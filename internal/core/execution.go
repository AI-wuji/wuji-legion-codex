package core

import (
	"encoding/hex"
	"fmt"
	"strings"
)

const workerReceiptSchemaVersion = 2

// ValidateWorkerReceiptConsistency checks a receipt against its route contract.
// It cannot attest that the Desktop host created the worker or that a model,
// Skill, or MCP call actually ran: receipt fields are caller supplied.
func ValidateWorkerReceiptConsistency(worker WorkerTask, receipt WorkerExecutionReceipt) error {
	if err := validateDelegatedModel(worker.Model); err != nil {
		return err
	}
	if err := validateWorkerExecutionPolicy(worker); err != nil {
		return err
	}
	if receipt.SchemaVersion != workerReceiptSchemaVersion {
		return fmt.Errorf("unsupported worker receipt schema: %d", receipt.SchemaVersion)
	}
	if receipt.WorkerID != worker.ID || receipt.RequestedModel != worker.Model {
		return fmt.Errorf("worker receipt identity does not match route")
	}
	if receipt.RequestedReasoningEffort != worker.ReasoningEffort {
		return fmt.Errorf("worker receipt requested reasoning effort does not match route")
	}
	if receipt.SessionKey == "" || receipt.SessionKey != worker.SessionKey {
		return fmt.Errorf("worker receipt session key does not match route")
	}
	if strings.TrimSpace(receipt.HostDispatchID) == "" {
		return fmt.Errorf("worker receipt lacks a host dispatch identifier")
	}
	expectedWriteBoundary := "read-only"
	if worker.Writes {
		expectedWriteBoundary = "scoped-artifact-write"
	}
	if receipt.WriteBoundary != expectedWriteBoundary {
		return fmt.Errorf("worker receipt write boundary does not match execution-node contract")
	}
	maxFallbacks := len(worker.AvailabilityFallbackModels)
	if len(worker.AvailabilityFallbackEfforts) > maxFallbacks {
		maxFallbacks = len(worker.AvailabilityFallbackEfforts)
	}
	maxAvailabilityAttempts := 1 + maxFallbacks
	if len(receipt.Attempts) == 0 || len(receipt.Attempts) > maxAvailabilityAttempts {
		return fmt.Errorf("worker receipt attempt count exceeds route policy")
	}
	if receipt.Attempts[0].Model != worker.Model {
		return fmt.Errorf("first worker attempt did not use requested model")
	}
	if receipt.Attempts[0].ReasoningEffort != worker.ReasoningEffort {
		return fmt.Errorf("first worker attempt did not use requested reasoning effort")
	}

	allowedFailure := make(map[string]bool, len(worker.AvailabilityFallbackOn))
	for _, kind := range worker.AvailabilityFallbackOn {
		allowedFailure[kind] = true
	}

	totalInput, totalCached, totalOutput := 0, 0, 0
	totalContext, totalPrefix, totalSources, totalContract := 0, 0, 0, 0
	failureKinds := make([]string, 0, len(receipt.Attempts))
	modelSwitches := 0
	reasoningSwitches := 0
	for index, attempt := range receipt.Attempts {
		if attempt.Model == "" || attempt.CacheDomain != "model-local:"+attempt.Model {
			return fmt.Errorf("attempt %d has an invalid model-local cache domain", index+1)
		}
		expectedReasoning := worker.ReasoningEffort
		if index > 0 && len(worker.AvailabilityFallbackEfforts) > 0 {
			expectedReasoning = worker.AvailabilityFallbackEfforts[index-1]
		}
		if attempt.ReasoningEffort != expectedReasoning {
			return fmt.Errorf("attempt %d uses reasoning effort %q, expected %q", index+1, attempt.ReasoningEffort, expectedReasoning)
		}
		if attempt.InputTokens < 0 || attempt.CachedInputTokens < 0 || attempt.OutputTokens < 0 || attempt.CachedInputTokens > attempt.InputTokens {
			return fmt.Errorf("attempt %d has invalid token accounting", index+1)
		}
		if attempt.ContextBytes != worker.AllocatedContextBytes || attempt.StablePrefixBytes != worker.StablePrefixBytes || attempt.SourceExecutionBytes != worker.SourceExecutionBytes || attempt.TaskContractBytes != worker.AllocatedTaskContractBytes {
			return fmt.Errorf("attempt %d replay bytes do not match route payload", index+1)
		}
		if index > 0 {
			previous := receipt.Attempts[index-1]
			expectedModel := worker.Model
			if len(worker.AvailabilityFallbackModels) > 0 {
				expectedModel = worker.AvailabilityFallbackModels[index-1]
			}
			if attempt.Model != expectedModel || previous.GenerationStarted || !allowedFailure[previous.FailureKind] {
				return fmt.Errorf("attempt %d violates generation-before-fallback policy", index+1)
			}
			if attempt.Model != previous.Model {
				modelSwitches++
			}
			if attempt.ReasoningEffort != previous.ReasoningEffort {
				reasoningSwitches++
			}
			if len(worker.AvailabilityFallbackModels) > 0 && modelStrength(attempt.Model) <= modelStrength(previous.Model) {
				return fmt.Errorf("attempt %d violates ordered fallback policy", index+1)
			}
		}
		if index < len(receipt.Attempts)-1 && attempt.FailureKind == "" {
			return fmt.Errorf("attempt %d has no fallback-eligible failure", index+1)
		}
		if attempt.FailureKind != "" {
			failureKinds = append(failureKinds, attempt.FailureKind)
		}
		totalInput += attempt.InputTokens
		totalCached += attempt.CachedInputTokens
		totalOutput += attempt.OutputTokens
		totalContext += attempt.ContextBytes
		totalPrefix += attempt.StablePrefixBytes
		totalSources += attempt.SourceExecutionBytes
		totalContract += attempt.TaskContractBytes
	}
	if receipt.ModelSwitchCount != modelSwitches || modelSwitches > len(worker.AvailabilityFallbackModels) {
		return fmt.Errorf("worker receipt model switch count violates route policy")
	}
	if receipt.ReasoningSwitchCount != reasoningSwitches || reasoningSwitches > len(worker.AvailabilityFallbackEfforts) {
		return fmt.Errorf("worker receipt reasoning switch count violates route policy")
	}

	last := receipt.Attempts[len(receipt.Attempts)-1]
	if last.FailureKind != "" || !last.GenerationStarted || !validResultHandle(receipt.ResultHandle) {
		return fmt.Errorf("worker receipt has no successful generated result")
	}
	if receipt.EffectiveModel != last.Model || receipt.EffectiveReasoningEffort != last.ReasoningEffort || receipt.RetryCount != len(receipt.Attempts)-1 {
		return fmt.Errorf("worker receipt effective model or retry count is inconsistent")
	}
	if receipt.StablePrefixBytesSent != totalPrefix || receipt.SourceExecutionBytesSent != totalSources || receipt.ContextBytesSent != totalContext || receipt.TaskContractBytes != totalContract {
		return fmt.Errorf("worker receipt replay byte totals are inconsistent")
	}
	if receipt.InputTokens != totalInput || receipt.CachedInputTokens != totalCached || receipt.OutputTokens != totalOutput {
		return fmt.Errorf("worker receipt token totals are inconsistent")
	}
	if !equalStringSlices(receipt.ContextHandleIDs, worker.ContextHandles) || !equalStringSlices(receipt.AttemptFailureKinds, failureKinds) {
		return fmt.Errorf("worker receipt context handles or failure kinds are inconsistent")
	}
	if receipt.CacheDomain != last.CacheDomain || receipt.DelegationGateReason != worker.DelegationGateReason {
		return fmt.Errorf("worker receipt cache domain or delegation reason is inconsistent")
	}
	if receipt.StablePrefixSHA256 != worker.StablePrefixSHA256 || receipt.ContextPayloadSHA256 != worker.ContextPayloadSHA256 || receipt.TaskContractSHA256 != worker.TaskContractSHA256 {
		return fmt.Errorf("worker receipt payload hashes are inconsistent")
	}
	if worker.TaskGraphID != "" && (receipt.TaskGraphID != "" || receipt.TaskGraphSHA256 != "" || receipt.ParentGraphID != "") &&
		(receipt.ParentGraphID != worker.ParentGraphID || receipt.TaskGraphID != worker.TaskGraphID || receipt.TaskGraphSHA256 != worker.TaskGraphSHA256) {
		return fmt.Errorf("worker receipt role graph identity or hash is inconsistent")
	}
	if strings.TrimSpace(receipt.BillingUnit) == "" || receipt.TotalCostMicrounits < 0 || receipt.ExecutionBaselineMicrounits <= 0 {
		return fmt.Errorf("worker receipt lacks usable billing evidence")
	}
	wantSavings := receipt.ExecutionBaselineMicrounits - receipt.TotalCostMicrounits
	if receipt.SavingsMicrounits != wantSavings {
		return fmt.Errorf("worker receipt savings calculation is inconsistent")
	}
	if worker.ModelClass != "sol" && receipt.SavingsMicrounits <= 0 {
		return fmt.Errorf("delegation did not beat the execution cost baseline")
	}
	return nil
}

// validateWorkerExecutionPolicy keeps model availability selection separate
// from paid execution retries. The host may inspect at most the two declared
// availability rungs before generation; after generation, the execution
// contract has exactly one attempt and no model switch budget.
func validateWorkerExecutionPolicy(worker WorkerTask) error {
	if worker.MaxAttempts != 1 || len(worker.FallbackModels) != 0 || len(worker.FallbackOn) != 0 || worker.MaxModelSwitches != 0 {
		return fmt.Errorf("worker route has an invalid single-execution retry budget")
	}
	if len(worker.AvailabilityFallbackModels) > maxAvailabilityFallbacks || len(worker.AvailabilityFallbackEfforts) > maxAvailabilityFallbacks {
		return fmt.Errorf("worker availability fallback chain exceeds bound")
	}
	if len(worker.AvailabilityFallbackModels) == 0 && len(worker.AvailabilityFallbackEfforts) == 0 {
		if len(worker.AvailabilityFallbackOn) != 0 {
			return fmt.Errorf("worker availability failure policy has no fallback")
		}
		return nil
	}
	if worker.ReasoningEffort != "low" && worker.ReasoningEffort != "medium" && worker.ReasoningEffort != "high" && worker.ReasoningEffort != "xhigh" && worker.ReasoningEffort != "max" && worker.ReasoningEffort != "none" && worker.ReasoningEffort != "" {
		return fmt.Errorf("worker reasoning effort %q is unsupported", worker.ReasoningEffort)
	}
	if len(worker.AvailabilityFallbackOn) == 0 {
		return fmt.Errorf("worker availability fallback chain has no failure policy")
	}
	seen := map[string]bool{worker.Model: true}
	for _, model := range worker.AvailabilityFallbackModels {
		if strings.TrimSpace(model) == "" || seen[model] {
			return fmt.Errorf("worker availability fallback chain contains an empty or duplicate model")
		}
		if err := validateDelegatedModel(model); err != nil {
			return err
		}
		seen[model] = true
	}
	seenEfforts := map[string]bool{worker.ReasoningEffort: true}
	for _, effort := range worker.AvailabilityFallbackEfforts {
		if strings.TrimSpace(effort) == "" || seenEfforts[effort] {
			return fmt.Errorf("worker reasoning fallback chain contains an empty or duplicate effort")
		}
		if effort != "low" && effort != "medium" && effort != "high" && effort != "xhigh" && effort != "max" && effort != "none" {
			return fmt.Errorf("worker reasoning fallback chain contains unsupported effort %q", effort)
		}
		previous := worker.ReasoningEffort
		if len(seenEfforts) > 1 {
			previous = worker.AvailabilityFallbackEfforts[len(seenEfforts)-2]
		}
		if reasoningStrength(effort) >= reasoningStrength(previous) {
			return fmt.Errorf("worker reasoning fallback must lower effort before generation")
		}
		seenEfforts[effort] = true
	}
	for _, failure := range worker.AvailabilityFallbackOn {
		if failure != "model-unavailable" && failure != "provider-error-before-generation" {
			return fmt.Errorf("worker availability fallback policy allows non-availability failure %q", failure)
		}
	}
	return nil
}

func reasoningStrength(effort string) int {
	switch effort {
	case "none":
		return 0
	case "low":
		return 1
	case "medium":
		return 2
	case "high":
		return 3
	case "xhigh":
		return 4
	case "max":
		return 5
	}
	return -1
}

// ValidateWorkerReceipt is retained for existing callers. It checks contract
// consistency only and must not be used as native execution verification.
func ValidateWorkerReceipt(worker WorkerTask, receipt WorkerExecutionReceipt) error {
	return ValidateWorkerReceiptConsistency(worker, receipt)
}

func validateDelegatedModel(model string) error {
	model = strings.ToLower(strings.TrimSpace(model))
	if model == hostSelectedModel || model == "" {
		return nil
	}
	if !strings.Contains(model, "gpt-") {
		return nil
	}
	return fmt.Errorf("GPT worker model %q is not allowed in host-selected mode; choose it explicitly in the host or use the host-selected sentinel", model)
}

func validResultHandle(handle string) bool {
	const prefix = "wuji-result://sha256/"
	if !strings.HasPrefix(handle, prefix) {
		return false
	}
	digest := strings.TrimPrefix(handle, prefix)
	if len(digest) != 64 || strings.ToLower(digest) != digest {
		return false
	}
	_, err := hex.DecodeString(digest)
	return err == nil
}

func modelStrength(model string) int {
	switch model {
	case hostSelectedModel:
		return 1
	default:
		return 0
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}
