package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const defaultProbeTimeout = 120 * time.Second
const maxProbeOutputBytes = 64 * 1024

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
	if err := ValidateManifest(manifest); err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, "invalid manifest: "+err.Error())
		return result
	}
	if len(manifest.Engines) > 0 {
		result.Checks = append(result.Checks, fmt.Sprintf("engine coverage valid: %d engines", len(manifest.Engines)))
	}
	for _, source := range manifest.Sources {
		path, ok := ResolveCompleteSourceAt(root, source)
		if !ok {
			path, ok = ResolveSourceAt(root, source)
		}
		if !ok {
			result.Passed = false
			result.Errors = append(result.Errors, "source not found: "+source.ID)
			continue
		}
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
	if result.Passed && len(manifest.Sources) > 0 {
		result.Effective = "assets-retained"
	}
	if result.Passed && manifest.HostCallable && manifest.PrimarySkill != "" {
		result.Effective = "callable"
		result.Checks = append(result.Checks, "complete package is exposed through the current Codex host: "+manifest.PrimarySkill)
	}
	if result.Passed && manifest.DirectMount && manifest.PrimarySkill != "" && len(manifest.Sources) > 0 {
		result.Effective = "callable"
		result.Checks = append(result.Checks, "complete cold package is directly mountable: "+manifest.PrimarySkill)
	}
	if manifest.Probe != nil && result.Passed {
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
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "WUJI_ROOT="+root)
		output := &limitedOutput{}
		cmd.Stdout = output
		cmd.Stderr = output
		err := cmd.Run()
		outputText := strings.TrimSpace(output.String())
		if probeContext.Err() == context.DeadlineExceeded {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("probe timed out after %s", timeout))
		} else if err != nil {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("probe failed: %v: %s", err, outputText))
		} else {
			kind := strings.ToLower(strings.TrimSpace(manifest.Probe.Kind))
			if kind == "" {
				kind = "behavior"
			}
			result.Checks = append(result.Checks, kind+" probe passed: "+outputText)
			switch kind {
			case "behavior":
				result.Effective = "behavior-verified"
			case "smoke", "mount":
				if rank(result.Effective) < rank("callable") {
					result.Effective = "callable"
				}
			default:
				result.Effective = "callable"
			}
		}
	}
	if result.Passed && manifest.Status == "primary" && result.Effective == "behavior-verified" {
		result.Effective = "primary"
	}
	if rank(manifest.Status) > rank(result.Effective) {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("claimed %s but evidence supports only %s", manifest.Status, result.Effective))
	}
	return result
}
