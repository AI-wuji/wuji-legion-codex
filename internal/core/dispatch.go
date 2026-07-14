package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CodexCommandResult struct {
	ExitCode    int
	Stderr      string
	OutputReady bool
}

type CodexCommandRunner func(context.Context, string, []string) (CodexCommandResult, error)

type DispatchOptions struct {
	CodexPath           string
	CodexArgumentPrefix []string
	Workspace           string
	OutputDir           string
	Timeout             time.Duration
	DryRun              bool
	// CompatibilityExec is deliberately opt-in. An external `codex exec`
	// process cannot inherit a Desktop host's native-agent identity or model
	// authentication, so it may help diagnose a local CLI installation but can
	// never be accepted as native worker execution evidence.
	CompatibilityExec bool
	TrustedManifests  []Manifest
	Runner            CodexCommandRunner
}

type DispatchAttempt struct {
	Model             string   `json:"model"`
	Arguments         []string `json:"arguments"`
	ExitCode          int      `json:"exit_code"`
	FailureKind       string   `json:"failure_kind,omitempty"`
	GenerationStarted bool     `json:"generation_started"`
	ResultPath        string   `json:"result_path,omitempty"`
	ResultSHA256      string   `json:"result_sha256,omitempty"`
	ResultHandle      string   `json:"result_handle,omitempty"`
	Stderr            string   `json:"stderr,omitempty"`
	CompletionMode    string   `json:"completion_mode,omitempty"`
}

type DispatchResult struct {
	WorkerID                 string                         `json:"worker_id"`
	SessionKey               string                         `json:"session_key"`
	ContractRequestID        string                         `json:"contract_request_id"`
	RequestedModel           string                         `json:"requested_model"`
	SucceededCLIModelRequest string                         `json:"succeeded_cli_model_request,omitempty"`
	Status                   string                         `json:"status"`
	WriteBoundary            string                         `json:"write_boundary"`
	Attempts                 []DispatchAttempt              `json:"attempts"`
	ModelEvidence            string                         `json:"model_evidence"`
	TelemetryStatus          string                         `json:"telemetry_status"`
	DispatchMode             string                         `json:"dispatch_mode"`
	NativeHostRequired       bool                           `json:"native_host_required"`
	PreparedPromptSHA256     string                         `json:"prepared_prompt_sha256"`
	PreparedPromptBytes      int                            `json:"prepared_prompt_bytes"`
	SourceContracts          []SourceEntrypointVerification `json:"source_contracts,omitempty"`
}

func DispatchWorker(worker WorkerTask, options DispatchOptions) (DispatchResult, error) {
	if strings.TrimSpace(worker.ID) == "" || strings.TrimSpace(worker.Model) == "" || strings.TrimSpace(worker.SessionKey) == "" {
		return DispatchResult{}, fmt.Errorf("worker is missing dispatch identity")
	}
	if worker.Writes {
		return DispatchResult{}, fmt.Errorf("worker dispatch violates Aji-only write authority")
	}
	if err := validateDelegatedModel(worker.Model); err != nil {
		return DispatchResult{}, err
	}
	if len(worker.FallbackModels) > 0 {
		return DispatchResult{}, fmt.Errorf("automatic worker model fallback is disabled")
	}
	if strings.TrimSpace(options.Workspace) == "" || strings.TrimSpace(options.OutputDir) == "" {
		return DispatchResult{}, fmt.Errorf("dispatch requires workspace and output directory")
	}
	workspace, err := filepath.Abs(options.Workspace)
	if err != nil {
		return DispatchResult{}, err
	}
	outputDir, err := filepath.Abs(options.OutputDir)
	if err != nil {
		return DispatchResult{}, err
	}
	if options.Timeout <= 0 {
		options.Timeout = 90 * time.Second
	}
	if strings.TrimSpace(options.CodexPath) == "" {
		invocation := defaultCodexInvocation()
		options.CodexPath = invocation.Path
		options.CodexArgumentPrefix = invocation.ArgumentPrefix
	}
	if options.Runner == nil {
		options.Runner = runCodexCommand
	}
	if !options.DryRun {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return DispatchResult{}, err
		}
	}
	if len(worker.SourceExecution) > 0 && len(options.TrustedManifests) == 0 {
		return DispatchResult{}, fmt.Errorf("trusted manifests are required for selected source dispatch")
	}
	verifiedSources, sourceVerification, err := VerifySourceExecutionContracts(options.TrustedManifests, worker.SourceExecution)
	if err != nil {
		return DispatchResult{}, err
	}
	worker.SourceExecution = verifiedSources

	prompt := workerPrompt(worker)
	promptDigest := sha256.Sum256([]byte(prompt))
	result := DispatchResult{
		WorkerID:             worker.ID,
		SessionKey:           worker.SessionKey,
		ContractRequestID:    dispatchID(worker, outputDir),
		RequestedModel:       worker.Model,
		Status:               "native-host-dispatch-required",
		WriteBoundary:        "read-only",
		ModelEvidence:        "Only a Desktop native child created with this exact model can prove model execution. An external codex exec process is compatibility-only and is never native execution evidence.",
		TelemetryStatus:      "unavailable-from-codex-cli",
		DispatchMode:         "native-host-contract",
		NativeHostRequired:   true,
		PreparedPromptSHA256: hex.EncodeToString(promptDigest[:]),
		PreparedPromptBytes:  len(prompt),
		SourceContracts:      sourceVerification,
	}
	// The Desktop host reads the verified contract and creates the actual
	// Sol/Luna child. Do not silently substitute a separately authenticated
	// command-line process and then call it a routed worker.
	if !options.CompatibilityExec {
		return result, nil
	}
	result.DispatchMode = "external-cli-compatibility-untrusted"
	result.NativeHostRequired = true
	result.Status = "compatibility-exec-failed-before-generation"
	models := append([]string{worker.Model}, worker.FallbackModels...)
	for index, model := range models {
		if worker.MaxAttempts > 0 && index >= worker.MaxAttempts {
			break
		}
		resultPath := filepath.Join(outputDir, fmt.Sprintf("%s-%s-%d.txt", safeDispatchName(worker.ID), dispatchSuffix(result.ContractRequestID), index+1))
		arguments := append(append([]string{}, options.CodexArgumentPrefix...), codexArguments(model, workspace, resultPath, prompt)...)
		attempt := DispatchAttempt{Model: model, Arguments: arguments}
		if options.DryRun {
			attempt.ExitCode = 0
			result.Attempts = append(result.Attempts, attempt)
			result.Status = "compatibility-exec-planned-untrusted"
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
		commandResult, commandErr := options.Runner(ctx, options.CodexPath, arguments)
		cancel()
		attempt.ExitCode = commandResult.ExitCode
		attempt.Stderr = compactDispatchStderr(commandResult.Stderr)
		if commandResult.OutputReady {
			attempt.CompletionMode = "output-file-observed"
		}
		content, readErr := os.ReadFile(resultPath)
		if readErr == nil && len(strings.TrimSpace(string(content))) > 0 {
			digest := sha256.Sum256(content)
			attempt.GenerationStarted = true
			attempt.ResultPath = resultPath
			attempt.ResultSHA256 = hex.EncodeToString(digest[:])
			attempt.ResultHandle = "wuji-result://sha256/" + attempt.ResultSHA256
			if failureKind := workerOutputFailureKind(content); failureKind != "" {
				attempt.FailureKind = failureKind
				result.Attempts = append(result.Attempts, attempt)
				result.Status = "compatibility-exec-failed-after-generation"
				return result, nil
			}
			result.Attempts = append(result.Attempts, attempt)
			if commandErr == nil && commandResult.ExitCode == 0 {
				result.Status = "compatibility-exec-completed-untrusted"
				result.SucceededCLIModelRequest = model
			} else {
				result.Status = "compatibility-exec-failed-after-generation"
			}
			return result, nil
		}
		attempt.FailureKind = dispatchFailureKind(commandResult.Stderr, commandErr)
		result.Attempts = append(result.Attempts, attempt)
		if index == len(models)-1 || (worker.MaxAttempts > 0 && index+1 >= worker.MaxAttempts) || !containsString(worker.FallbackOn, attempt.FailureKind) {
			return result, nil
		}
	}
	return result, nil
}

func dispatchSuffix(handle string) string {
	const prefix = "codex-exec://sha256/"
	value := strings.TrimPrefix(handle, prefix)
	if len(value) > 12 {
		return value[:12]
	}
	return safeDispatchName(value)
}

func codexArguments(model, workspace, resultPath, prompt string) []string {
	// `codex exec` takes the task as its final positional PROMPT. Adding `--`
	// before it suppresses that positional argument and makes the CLI read stdin,
	// producing a generic interactive response instead of executing the worker.
	return []string{"exec", "-m", model, "--sandbox", "read-only", "--ephemeral", "-C", workspace, "--output-last-message", resultPath, prompt}
}

func isActionableWorkerOutput(content []byte) bool {
	return workerOutputFailureKind(content) == ""
}

func workerOutputFailureKind(content []byte) string {
	value := strings.ToLower(strings.TrimSpace(string(content)))
	if value == "" {
		return "non-actionable-output"
	}
	for _, generic := range []string{
		"how can i help with the repository?",
		"what would you like me to work on in the repository?",
		"ready for the next task.",
		"ready. what would you like me to inspect, change, or verify in the repository?",
	} {
		if value == generic {
			return "non-actionable-output"
		}
	}
	for _, blocked := range []string{
		"blocked by sandbox policy",
		"blocked by the sandbox policy",
		"sandbox policy blocked",
		"filesystem read/list operation was blocked",
	} {
		if strings.Contains(value, blocked) {
			return "worker-blocked"
		}
	}
	return ""
}

type codexInvocation struct {
	Path           string
	ArgumentPrefix []string
}

var (
	globalCodexScriptOnce sync.Once
	globalCodexScriptPath string
)

func defaultCodexPath() string {
	return defaultCodexInvocation().Path
}

func defaultCodexInvocation() codexInvocation {
	// On Windows, invoking an npm .ps1/.cmd shim through os/exec can corrupt a
	// multiline task contract. Launch the package's Node entrypoint directly
	// when the global installation can be resolved. This avoids both the shim
	// and the desktop-app binary, which may be inaccessible to this process.
	if runtime.GOOS == "windows" {
		if nodePath, err := exec.LookPath("node.exe"); err == nil {
			if scriptPath, ok := globalCodexScript(); ok {
				return codexInvocation{Path: nodePath, ArgumentPrefix: []string{scriptPath}}
			}
		}
	}
	return codexInvocation{Path: "codex"}
}

func globalCodexScript() (string, bool) {
	globalCodexScriptOnce.Do(func() {
		command := exec.Command("npm", "root", "-g")
		output, err := command.Output()
		if err != nil {
			return
		}
		root := strings.TrimSpace(string(output))
		if root == "" {
			return
		}
		script := filepath.Join(root, "@openai", "codex", "bin", "codex.js")
		info, err := os.Stat(script)
		if err == nil && !info.IsDir() {
			globalCodexScriptPath = script
		}
	})
	return globalCodexScriptPath, globalCodexScriptPath != ""
}

func workerPrompt(worker WorkerTask) string {
	parts := []string{
		worker.StableCapabilityPrefix,
	}
	for _, source := range worker.SourceExecution {
		parts = append(parts, strings.Join([]string{
			"Selected capability entrypoint: " + source.SourceID,
			"Activation reason: " + source.ActivationReason,
			"Entrypoint path (host-resolved): " + source.ResolvedEntrypointPath,
			"Entrypoint SHA-256: " + source.EntrypointSHA256,
			"The following is the exact host-verified Skill entrypoint selected for this task. Apply it; do not load unrelated source trees:",
			source.EntrypointContent,
		}, "\n"))
	}
	parts = append(parts,
		"The task contract below is the complete user request for this worker. Execute its objective now; do not ask for a separate task. Return only the requested evidence and options.",
		worker.ContextPayload,
		worker.TaskContract,
		"Execution boundary: read-only. Return compact evidence and options only. Do not write workspace files or claim final completion.",
	)
	return strings.Join(parts, "\n\n")
}

func runCodexCommand(ctx context.Context, path string, arguments []string) (CodexCommandResult, error) {
	command := exec.CommandContext(ctx, path, arguments...)
	logFile, err := os.CreateTemp("", "wuji-dispatch-*.log")
	if err != nil {
		return CodexCommandResult{}, err
	}
	logPath := logFile.Name()
	defer func() {
		_ = logFile.Close()
		_ = os.Remove(logPath)
	}()
	command.Stdout = logFile
	command.Stderr = logFile
	if err := command.Start(); err != nil {
		return CodexCommandResult{}, err
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	commandOutput := func() string {
		_ = logFile.Sync()
		content, readErr := os.ReadFile(logPath)
		if readErr != nil {
			return ""
		}
		return string(content)
	}
	resultPath := outputLastMessagePath(arguments)
	for {
		select {
		case err := <-done:
			result := completedCodexResult(command, commandOutput())
			result.OutputReady = resultPath != "" && hasDispatchOutput(resultPath)
			return result, err
		case <-ctx.Done():
			terminateCommandTree(command)
			err := <-done
			return completedCodexResult(command, commandOutput()), err
		}
	}
}

func terminateCommandTree(command *exec.Cmd) {
	if command == nil || command.Process == nil {
		return
	}
	if runtime.GOOS == "windows" {
		_ = exec.Command("taskkill", "/PID", strconv.Itoa(command.Process.Pid), "/T", "/F").Run()
		return
	}
	_ = command.Process.Kill()
}

func completedCodexResult(command *exec.Cmd, stderr string) CodexCommandResult {
	result := CodexCommandResult{Stderr: stderr}
	if command.ProcessState != nil {
		result.ExitCode = command.ProcessState.ExitCode()
	}
	return result
}

func outputLastMessagePath(arguments []string) string {
	for index := 0; index+1 < len(arguments); index++ {
		if arguments[index] == "--output-last-message" {
			return arguments[index+1]
		}
	}
	return ""
}

func hasDispatchOutput(path string) bool {
	content, err := os.ReadFile(path)
	return err == nil && len(strings.TrimSpace(string(content))) > 0
}

func dispatchFailureKind(stderr string, commandErr error) string {
	text := strings.ToLower(stderr)
	if strings.Contains(text, "model") && (strings.Contains(text, "not found") || strings.Contains(text, "unavailable") || strings.Contains(text, "unsupported")) {
		return "model-unavailable"
	}
	if commandErr != nil {
		return "provider-error-before-generation"
	}
	return "provider-error-before-generation"
}

func dispatchID(worker WorkerTask, outputDir string) string {
	payload := worker.SessionKey + "\x00" + worker.ID + "\x00" + outputDir + "\x00" + time.Now().UTC().Format(time.RFC3339Nano)
	digest := sha256.Sum256([]byte(payload))
	return "codex-exec://sha256/" + hex.EncodeToString(digest[:])
}

func safeDispatchName(value string) string {
	value = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, value)
	value = strings.Trim(value, "-")
	if value == "" {
		return "worker"
	}
	return value
}

func compactDispatchStderr(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 2048 {
		return value
	}
	return value[:2048]
}
