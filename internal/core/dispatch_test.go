package core

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestDispatchWorkerPreparesNativeHostContractWithoutStartingExternalCLI(t *testing.T) {
	worker := testReceiptWorker()
	outputDir := t.TempDir()
	result, err := DispatchWorker(worker, DispatchOptions{
		Workspace: t.TempDir(), OutputDir: outputDir,
		Runner: func(_ context.Context, _ string, arguments []string) (CodexCommandResult, error) {
			t.Fatalf("external CLI must not run without explicit compatibility mode: %#v", arguments)
			return CodexCommandResult{}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "native-host-dispatch-required" || !result.NativeHostRequired || result.PreparedPromptSHA256 == "" || result.PreparedPromptBytes == 0 {
		t.Fatalf("unexpected dispatch result: %#v", result)
	}
}

func TestCompatibilityExecIsExplicitAndUntrusted(t *testing.T) {
	worker := testReceiptWorker()
	result, err := DispatchWorker(worker, DispatchOptions{
		Workspace: t.TempDir(), OutputDir: t.TempDir(), CompatibilityExec: true,
		Runner: func(_ context.Context, _ string, arguments []string) (CodexCommandResult, error) {
			if argumentValue(arguments, "-m") != worker.Model {
				t.Fatalf("compatibility command changed requested model: %#v", arguments)
			}
			if err := os.WriteFile(outputLastMessagePath(arguments), []byte("evidence"), 0o600); err != nil {
				t.Fatal(err)
			}
			return CodexCommandResult{}, nil
		},
	})
	if err != nil || result.Status != "compatibility-exec-completed-untrusted" || result.SucceededCLIModelRequest != worker.Model || len(result.Attempts) != 1 {
		t.Fatalf("compatibility dispatch was incorrectly reported: %#v err=%v", result, err)
	}
}

func TestCodexArgumentsPassWorkerPromptAsFinalPositionalArgument(t *testing.T) {
	arguments := codexArguments("gpt-5.6-luna", "workspace", "result.txt", "{\"objective\":\"review\"}")
	if containsString(arguments, "--") {
		t.Fatalf("worker prompt was suppressed by option termination: %#v", arguments)
	}
	if got, want := arguments[len(arguments)-1], "{\"objective\":\"review\"}"; got != want {
		t.Fatalf("worker prompt was not the final positional argument: got %q want %q", got, want)
	}
}

func TestDefaultCodexPathIsNeverEmpty(t *testing.T) {
	if strings.TrimSpace(defaultCodexPath()) == "" {
		t.Fatal("default Codex executable path is empty")
	}
}

func TestWorkerPromptMakesTheTaskContractActionable(t *testing.T) {
	worker := testReceiptWorker()
	prompt := workerPrompt(worker)
	if !strings.Contains(prompt, "Execute its objective now; do not ask for a separate task.") {
		t.Fatalf("worker prompt does not instruct the worker to execute its task contract: %q", prompt)
	}
}

func TestWorkerPromptDoesNotLoadUnselectedSources(t *testing.T) {
	worker := testReceiptWorker()
	worker.SourceExecution = []SourceExecutionContract{{
		SourceID: "selected", Capability: "focused", InvocationKind: sourceEntrypointInvocationKind, Entrypoint: "SKILL.md",
		EntrypointSHA256: strings.Repeat("a", 64), EntrypointBytes: 8, ActivationReason: "primary-source", EntrypointContent: "selected-body",
	}}
	prompt := workerPrompt(worker)
	if !strings.Contains(prompt, "selected-body") || strings.Contains(prompt, "unselected-body") {
		t.Fatalf("worker prompt did not stay scoped to the selected entrypoint: %q", prompt)
	}
}

func TestDispatchWorkerRejectsAutomaticFallback(t *testing.T) {
	worker := testReceiptWorker()
	worker.FallbackModels = []string{"gpt-5.6-sol"}
	if _, err := DispatchWorker(worker, DispatchOptions{Workspace: t.TempDir(), OutputDir: t.TempDir()}); err == nil {
		t.Fatal("dispatch accepted an automatic fallback")
	}
}

func TestDispatchWorkerDoesNotRetryAfterGeneration(t *testing.T) {
	worker := testReceiptWorker()
	outputDir := t.TempDir()
	result, err := DispatchWorker(worker, DispatchOptions{
		Workspace: t.TempDir(), OutputDir: outputDir, CompatibilityExec: true,
		Runner: func(_ context.Context, _ string, arguments []string) (CodexCommandResult, error) {
			if err := os.WriteFile(outputLastMessagePath(arguments), []byte("partial response"), 0o600); err != nil {
				t.Fatal(err)
			}
			return CodexCommandResult{ExitCode: 1, Stderr: "provider disconnected"}, errors.New("provider disconnected")
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Attempts) != 1 || !result.Attempts[0].GenerationStarted || result.Status != "compatibility-exec-failed-after-generation" {
		t.Fatalf("generated partial response was retried: %#v", result)
	}
}

func TestDispatchWorkerHonorsRouteFallbackFailures(t *testing.T) {
	worker := testReceiptWorker()
	var calls int
	result, err := DispatchWorker(worker, DispatchOptions{
		Workspace: t.TempDir(), OutputDir: t.TempDir(), CompatibilityExec: true,
		Runner: func(_ context.Context, _ string, _ []string) (CodexCommandResult, error) {
			calls++
			return CodexCommandResult{ExitCode: 1, Stderr: "unexpected command failure"}, errors.New("unexpected command failure")
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || len(result.Attempts) != 1 || result.Attempts[0].FailureKind != "provider-error-before-generation" {
		t.Fatalf("dispatch ignored route fallback policy: %#v", result)
	}
}

func TestDispatchWorkerRejectsLegacyGPTModelsButAllowsExplicitOtherProviders(t *testing.T) {
	legacy := testReceiptWorker()
	legacy.Model = "gpt-5.4-mini"
	if _, err := DispatchWorker(legacy, DispatchOptions{Workspace: t.TempDir(), OutputDir: t.TempDir()}); err == nil {
		t.Fatal("dispatch accepted a legacy GPT worker model")
	}

	external := testReceiptWorker()
	external.Model = "grok-4.5"
	result, err := DispatchWorker(external, DispatchOptions{Workspace: t.TempDir(), OutputDir: t.TempDir()})
	if err != nil || result.RequestedModel != "grok-4.5" || result.Status != "native-host-dispatch-required" {
		t.Fatalf("explicit non-GPT provider was incorrectly blocked: result=%#v err=%v", result, err)
	}
}

func TestDispatchWorkerRequiresACompletedCodexCommand(t *testing.T) {
	worker := testReceiptWorker()
	result, err := DispatchWorker(worker, DispatchOptions{
		Workspace: t.TempDir(), OutputDir: t.TempDir(), CompatibilityExec: true,
		Runner: func(_ context.Context, _ string, arguments []string) (CodexCommandResult, error) {
			if err := os.WriteFile(outputLastMessagePath(arguments), []byte("final output"), 0o600); err != nil {
				t.Fatal(err)
			}
			return CodexCommandResult{ExitCode: 1, OutputReady: true}, errors.New("process did not exit")
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "compatibility-exec-failed-after-generation" || result.Attempts[0].CompletionMode != "output-file-observed" {
		t.Fatalf("an interrupted command was accepted as worker evidence: %#v", result)
	}
}

func TestDispatchWorkerRejectsGenericInteractiveOutput(t *testing.T) {
	worker := testReceiptWorker()
	result, err := DispatchWorker(worker, DispatchOptions{
		Workspace: t.TempDir(), OutputDir: t.TempDir(), CompatibilityExec: true,
		Runner: func(_ context.Context, _ string, arguments []string) (CodexCommandResult, error) {
			if err := os.WriteFile(outputLastMessagePath(arguments), []byte("How can I help with the repository?"), 0o600); err != nil {
				t.Fatal(err)
			}
			return CodexCommandResult{ExitCode: 0}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "compatibility-exec-failed-after-generation" || result.Attempts[0].FailureKind != "non-actionable-output" {
		t.Fatalf("generic interactive output was accepted as worker evidence: %#v", result)
	}
}

func TestDispatchWorkerRejectsSandboxBlockedOutput(t *testing.T) {
	worker := testReceiptWorker()
	result, err := DispatchWorker(worker, DispatchOptions{
		Workspace: t.TempDir(), OutputDir: t.TempDir(), CompatibilityExec: true,
		Runner: func(_ context.Context, _ string, arguments []string) (CodexCommandResult, error) {
			content := "Paths: unavailable. Filesystem read/list operation was blocked by sandbox policy."
			if err := os.WriteFile(outputLastMessagePath(arguments), []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			return CodexCommandResult{ExitCode: 0}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "compatibility-exec-failed-after-generation" || result.Attempts[0].FailureKind != "worker-blocked" {
		t.Fatalf("sandbox-blocked output was accepted as worker evidence: %#v", result)
	}
}

func argumentValue(arguments []string, flag string) string {
	for index := 0; index+1 < len(arguments); index++ {
		if arguments[index] == flag {
			return arguments[index+1]
		}
	}
	return ""
}

func TestRunCodexCommandWaitsForProcessExitAfterObservedOutput(t *testing.T) {
	if os.Getenv("WUJI_DISPATCH_HELPER") == "1" {
		path := os.Getenv("WUJI_DISPATCH_RESULT")
		if err := os.WriteFile(path, []byte("final output"), 0o600); err != nil {
			os.Exit(2)
		}
		time.Sleep(200 * time.Millisecond)
		os.Exit(0)
	}

	resultPath := t.TempDir() + "/result.txt"
	t.Setenv("WUJI_DISPATCH_HELPER", "1")
	t.Setenv("WUJI_DISPATCH_RESULT", resultPath)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	started := time.Now()
	result, err := runCodexCommand(ctx, os.Args[0], []string{"-test.run=TestRunCodexCommandWaitsForProcessExitAfterObservedOutput", "--", "--output-last-message", resultPath})
	if err != nil || !result.OutputReady || time.Since(started) < 150*time.Millisecond {
		t.Fatalf("dispatcher accepted an output file before the worker process exited: result=%#v err=%v", result, err)
	}
}
