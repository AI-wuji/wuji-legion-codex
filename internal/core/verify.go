package core

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultProbeTimeout = 120 * time.Second
const maxProbeOutputBytes = 64 * 1024
const probeTerminationWait = 2 * time.Second

type limitedOutput struct {
	mu        sync.Mutex
	data      []byte
	truncated bool
}

func (output *limitedOutput) Write(data []byte) (int, error) {
	output.mu.Lock()
	defer output.mu.Unlock()
	remaining := maxProbeOutputBytes - len(output.data)
	if remaining > 0 {
		keep := len(data)
		if keep > remaining {
			keep = remaining
		}
		output.data = append(output.data, data[:keep]...)
	}
	if len(data) > remaining {
		output.truncated = true
	}
	return len(data), nil
}

func (output *limitedOutput) String() string {
	output.mu.Lock()
	defer output.mu.Unlock()
	value := string(output.data)
	if output.truncated {
		value += fmt.Sprintf("\n[output truncated at %d bytes]", maxProbeOutputBytes)
	}
	return value
}

func Verify(root string, manifest Manifest) VerifyResult {
	result := VerifyResult{Capability: manifest.ID, Claimed: manifest.Status, Effective: "known", Passed: true}
	if manifest.Root == "" {
		manifest.Root = root
	}
	if err := ValidateManifest(manifest); err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, "invalid manifest: "+err.Error())
		return result
	}
	if len(manifest.Engines) > 0 {
		result.Checks = append(result.Checks, fmt.Sprintf("engine coverage valid: %d engines", len(manifest.Engines)))
	}
	retainedSources := 0
	for _, source := range manifest.Sources {
		path, complete := ResolveCompleteSourceAt(root, source)
		if !complete {
			partialPath, present := ResolveSourceAt(root, source)
			priority := sourcePriority(source)
			if priority != "primary" {
				// Secondary material is deliberately cold unless the caller asks for
				// a full capability audit. The primary route must remain verifiable on
				// hosts that do not have every retained integration installed.
				state := "unavailable"
				if present {
					state = "incomplete"
				}
				result.Checks = append(result.Checks, priority+" source "+state+": "+source.ID)
				continue
			}
			if !present {
				result.Passed = false
				result.Errors = append(result.Errors, "source not found: "+source.ID)
				continue
			}
			path = partialPath
		}
		if path == "" {
			result.Passed = false
			result.Errors = append(result.Errors, "source not found: "+source.ID)
			continue
		}
		retainedSources++
		result.Sources = append(result.Sources, path)
		for _, required := range source.Required {
			matches, _ := filepath.Glob(filepath.Join(path, filepath.FromSlash(required)))
			if len(matches) == 0 {
				result.Passed = false
				result.Errors = append(result.Errors, fmt.Sprintf("%s missing %s", source.ID, required))
			} else {
				result.Checks = append(result.Checks, fmt.Sprintf("%s retains %s", source.ID, required))
			}
		}
	}
	if result.Passed && retainedSources > 0 {
		result.Effective = "assets-retained"
	}
	if manifest.Genome != nil {
		genome, assets, issues := verifyFusionGenome(manifest)
		result.Genome = &genome
		result.Assets = assets
		if len(issues) > 0 {
			result.Passed = false
			result.Errors = append(result.Errors, issues...)
		} else {
			result.Checks = append(result.Checks, fmt.Sprintf("fusion genome entrypoints and assets reachable: %d adapters", len(genome.Adapters)))
		}
	}
	if manifest.Probe != nil && result.Passed {
		kind := strings.ToLower(strings.TrimSpace(manifest.Probe.Kind))
		if kind == "" {
			kind = "behavior"
		}
		command := ExpandPathAt(root, manifest.Probe.Command)
		args := make([]string, len(manifest.Probe.Args))
		for i, arg := range manifest.Probe.Args {
			args[i] = ExpandPathAt(root, arg)
		}
		timeout := defaultProbeTimeout
		if manifest.Probe.TimeoutSeconds > 0 {
			timeout = time.Duration(manifest.Probe.TimeoutSeconds) * time.Second
		}
		probeContext, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		cmd := exec.CommandContext(probeContext, command, args...)
		// CommandContext only terminates the direct process by default. Make its
		// cancellation hook terminate the probe tree, and bound Wait so inherited
		// stdout/stderr from a surviving descendant cannot hold verification open.
		cmd.Cancel = func() error { return terminateProbeProcessTree(cmd) }
		cmd.WaitDelay = probeTerminationWait
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "WUJI_ROOT="+root)
		var evidenceDir string
		if kind == "behavior" {
			var evidenceErr error
			evidenceDir, evidenceErr = os.MkdirTemp("", "wuji-probe-evidence-*")
			if evidenceErr != nil {
				result.Passed = false
				result.Errors = append(result.Errors, "create probe evidence directory: "+evidenceErr.Error())
				return result
			}
			defer os.RemoveAll(evidenceDir)
			cmd.Env = append(cmd.Env, "WUJI_PROBE_EVIDENCE_DIR="+evidenceDir)
		}
		output := &limitedOutput{}
		cmd.Stdout = output
		cmd.Stderr = output
		err := runProbeCommand(probeContext, cmd)
		outputText := strings.TrimSpace(output.String())
		if probeContext.Err() == context.DeadlineExceeded {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("probe timed out after %s", timeout))
		} else if err != nil {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("probe failed: %v: %s", err, outputText))
		} else {
			result.Checks = append(result.Checks, kind+" probe passed: "+outputText)
			switch kind {
			case "behavior":
				evidence, evidenceErr := parseProbeEvidence(outputText, manifest.Probe.Fixture, evidenceDir, manifest.Probe.RequiredEvidence)
				if evidenceErr != nil {
					result.Passed = false
					result.Errors = append(result.Errors, "behavior probe returned invalid evidence: "+evidenceErr.Error())
					break
				}
				result.Probe = &evidence
				result.Effective = "behavior-verified"
			case "smoke", "mount":
				if rank(result.Effective) < rank("callable") {
					result.Effective = "callable"
				}
			}
		}
	}
	if result.Passed && manifest.Status == "primary" && result.Effective == "behavior-verified" {
		if err := verifyPromotionReceipt(root, manifest, *result.Probe); err != nil {
			result.Passed = false
			result.Errors = append(result.Errors, "primary promotion evidence is invalid: "+err.Error())
		} else {
			result.Effective = "primary"
			result.Checks = append(result.Checks, "content-addressed promotion receipt verified: "+manifest.PromotionReceipt)
		}
	}
	if rank(manifest.Status) > rank(result.Effective) {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("claimed %s but evidence supports only %s", manifest.Status, result.Effective))
	}
	return result
}

func runProbeCommand(ctx context.Context, cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		// CommandContext invokes the single cmd.Cancel hook installed by Verify
		// and enforces WaitDelay. Do not race it with a second taskkill here.
		// Plain-command callers must install their own bounded cancellation hook.
		select {
		case err := <-done:
			return err
		case <-time.After(probeTerminationWait):
			return fmt.Errorf("probe did not exit within %s of cancellation", probeTerminationWait)
		}
	}
}

func terminateProbeProcessTree(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	if runtime.GOOS == "windows" {
		// taskkill is also external work: give it an independent finite budget.
		// This is best-effort cleanup, not a hard containment boundary.
		killContext, cancel := context.WithTimeout(context.Background(), probeTerminationWait)
		defer cancel()
		return exec.CommandContext(killContext, "taskkill", "/PID", strconv.Itoa(cmd.Process.Pid), "/T", "/F").Run()
	}
	return cmd.Process.Kill()
}

func verifyPromotionReceipt(root string, manifest Manifest, evidence ProbeEvidence) error {
	cleanReceipt := filepath.Clean(filepath.FromSlash(manifest.PromotionReceipt))
	if !safeRelativePath(manifest.PromotionReceipt) || !strings.HasPrefix(cleanReceipt, "releases"+string(filepath.Separator)) {
		return fmt.Errorf("promotion receipt must be under releases/")
	}
	receiptPath := filepath.Join(root, "capabilities", manifest.ID, cleanReceipt)
	receiptData, err := readRegularFile(receiptPath)
	if err != nil {
		return err
	}
	receiptHash := fmt.Sprintf("%x", sha256.Sum256(receiptData))
	if filepath.Base(cleanReceipt) != receiptHash+".json" {
		return fmt.Errorf("promotion receipt filename does not match its sha256")
	}
	receipt, err := decodePromotionReceipt(receiptData)
	if err != nil {
		return err
	}
	if receipt.SchemaVersion != 1 || receipt.Capability != manifest.ID || receipt.Decision != "replace" {
		return fmt.Errorf("promotion receipt identity or decision is invalid")
	}
	if manifest.Probe == nil || receipt.Fixture != manifest.Probe.Fixture || receipt.Fixture != evidence.Fixture {
		return fmt.Errorf("promotion receipt fixture does not match the live behavior probe")
	}
	contractHash, err := manifestContractSHA256(manifest)
	if err != nil {
		return err
	}
	if receipt.Candidate.ContractSHA256 != contractHash || receipt.Candidate.EffectiveStatus != "behavior-verified" {
		return fmt.Errorf("promotion receipt does not bind the installed candidate contract")
	}
	comparison := findProbeArtifact(evidence.Evidence, manifest.Probe.ComparisonEvidence)
	if comparison == nil || receipt.Candidate.Signature != evidence.Signature || !samePromotionArtifact(receipt.Candidate.ComparisonEvidence, *comparison) {
		return fmt.Errorf("promotion receipt does not match live candidate evidence")
	}
	if rank(receipt.Baseline.EffectiveStatus) < rank("behavior-verified") || receipt.Baseline.Signature != receipt.Candidate.Signature || receipt.Baseline.ComparisonEvidence.ID != receipt.Candidate.ComparisonEvidence.ID || receipt.Baseline.ComparisonEvidence.SHA256 != receipt.Candidate.ComparisonEvidence.SHA256 {
		return fmt.Errorf("promotion receipt does not prove equivalent baseline behavior")
	}
	cleanBaseline := filepath.Clean(filepath.FromSlash(receipt.BaselineManifest))
	requiredPrefix := filepath.Join("retired", manifest.ID) + string(filepath.Separator)
	if !safeRelativePath(receipt.BaselineManifest) || !strings.HasPrefix(cleanBaseline, requiredPrefix) || filepath.Base(cleanBaseline) != "manifest.json" {
		return fmt.Errorf("promotion receipt baseline manifest path is invalid")
	}
	baselineData, err := readRegularFile(filepath.Join(root, cleanBaseline))
	if err != nil {
		return fmt.Errorf("archived baseline manifest: %w", err)
	}
	baseline, err := decodeManifest(baselineData)
	if err != nil {
		return fmt.Errorf("archived baseline manifest: %w", err)
	}
	if baseline.ID != manifest.ID {
		return fmt.Errorf("archived baseline capability id does not match")
	}
	baselineHash, err := manifestContractSHA256(baseline)
	if err != nil {
		return err
	}
	if baselineHash != receipt.Baseline.ContractSHA256 {
		return fmt.Errorf("archived baseline contract sha256 does not match receipt")
	}
	return nil
}

func decodePromotionReceipt(data []byte) (PromotionReceipt, error) {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var receipt PromotionReceipt
	if err := decoder.Decode(&receipt); err != nil {
		return PromotionReceipt{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return PromotionReceipt{}, fmt.Errorf("multiple JSON values are not allowed")
		}
		return PromotionReceipt{}, err
	}
	return receipt, nil
}

func readRegularFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", path)
	}
	return os.ReadFile(path)
}

func manifestContractSHA256(manifest Manifest) (string, error) {
	manifest.Root = ""
	manifest.Status = "behavior-verified"
	manifest.PromotionReceipt = ""
	data, err := json.Marshal(manifest)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func samePromotionArtifact(expected PromotionArtifact, actual ProbeArtifact) bool {
	return expected.ID == actual.ID && expected.SHA256 == actual.SHA256 && expected.Size == actual.Size
}

func parseProbeEvidence(output, fixture, evidenceDir string, required []string) (ProbeEvidence, error) {
	var evidence ProbeEvidence
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		var candidate ProbeEvidence
		if err := json.Unmarshal([]byte(line), &candidate); err != nil {
			continue
		}
		if candidate.WujiProbe == "behavior" {
			evidence = candidate
		}
	}
	if evidence.WujiProbe != "behavior" {
		return ProbeEvidence{}, fmt.Errorf("missing wuji_probe=behavior receipt")
	}
	if evidence.Fixture != fixture {
		return ProbeEvidence{}, fmt.Errorf("fixture %q does not match %q", evidence.Fixture, fixture)
	}
	if !evidence.Passed {
		return ProbeEvidence{}, fmt.Errorf("receipt marked passed=false")
	}
	if strings.TrimSpace(evidence.Signature) == "" {
		return ProbeEvidence{}, fmt.Errorf("receipt contains no behavior signature")
	}
	verified := make(map[string]bool, len(evidence.Evidence))
	for index := range evidence.Evidence {
		artifact := &evidence.Evidence[index]
		if !componentIDPattern.MatchString(artifact.ID) || verified[artifact.ID] {
			return ProbeEvidence{}, fmt.Errorf("invalid or duplicate evidence id %q", artifact.ID)
		}
		path, err := resolveProbeEvidencePath(evidenceDir, artifact.Path)
		if err != nil {
			return ProbeEvidence{}, fmt.Errorf("evidence %s: %w", artifact.ID, err)
		}
		file, err := os.Open(path)
		if err != nil {
			return ProbeEvidence{}, fmt.Errorf("evidence %s cannot be opened: %w", artifact.ID, err)
		}
		hash := sha256.New()
		size, copyErr := io.Copy(hash, file)
		closeErr := file.Close()
		if copyErr != nil {
			return ProbeEvidence{}, fmt.Errorf("evidence %s cannot be hashed: %w", artifact.ID, copyErr)
		}
		if closeErr != nil {
			return ProbeEvidence{}, fmt.Errorf("evidence %s cannot be closed: %w", artifact.ID, closeErr)
		}
		actualHash := fmt.Sprintf("%x", hash.Sum(nil))
		if !strings.EqualFold(actualHash, artifact.SHA256) {
			return ProbeEvidence{}, fmt.Errorf("evidence %s sha256 mismatch", artifact.ID)
		}
		artifact.SHA256 = actualHash
		artifact.Size = size
		verified[artifact.ID] = true
	}
	for _, id := range required {
		if !verified[id] {
			return ProbeEvidence{}, fmt.Errorf("required evidence %q is missing", id)
		}
	}
	return evidence, nil
}

func resolveProbeEvidencePath(root, relative string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(strings.TrimSpace(relative)))
	if clean == "." || filepath.IsAbs(clean) || filepath.VolumeName(clean) != "" || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path must stay relative to the evidence directory")
	}
	path := filepath.Join(root, clean)
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("path is not a regular file")
	}
	return path, nil
}
