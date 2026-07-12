package core

import (
	"encoding/hex"
	"fmt"
	"strings"
)

const workerReceiptSchemaVersion = 2

func ValidateWorkerReceipt(worker WorkerTask, receipt WorkerExecutionReceipt) error {
	if receipt.SchemaVersion != workerReceiptSchemaVersion {
		return fmt.Errorf("unsupported worker receipt schema: %d", receipt.SchemaVersion)
	}
	if receipt.WorkerID != worker.ID || receipt.RequestedModel != worker.Model {
		return fmt.Errorf("worker receipt identity does not match route")
	}
	if receipt.SessionKey == "" || receipt.SessionKey != worker.SessionKey {
		return fmt.Errorf("worker receipt session key does not match route")
	}
	if strings.TrimSpace(receipt.HostDispatchID) == "" {
		return fmt.Errorf("worker receipt lacks a host dispatch identifier")
	}
	if worker.Writes || receipt.WriteBoundary != "read-only" {
		return fmt.Errorf("worker receipt violates Aji-only write authority")
	}
	if len(receipt.Attempts) == 0 || len(receipt.Attempts) > worker.MaxAttempts {
		return fmt.Errorf("worker receipt attempt count exceeds route policy")
	}
	if receipt.Attempts[0].Model != worker.Model {
		return fmt.Errorf("first worker attempt did not use requested model")
	}

	allowedFallback := make(map[string]bool, len(worker.FallbackModels))
	for _, model := range worker.FallbackModels {
		allowedFallback[model] = true
	}
	allowedFailure := make(map[string]bool, len(worker.FallbackOn))
	for _, kind := range worker.FallbackOn {
		allowedFailure[kind] = true
	}

	totalInput, totalCached, totalOutput := 0, 0, 0
	totalContext, totalPrefix, totalContract := 0, 0, 0
	failureKinds := make([]string, 0, len(receipt.Attempts))
	modelSwitches := 0
	for index, attempt := range receipt.Attempts {
		if attempt.Model == "" || attempt.CacheDomain != "model-local:"+attempt.Model {
			return fmt.Errorf("attempt %d has an invalid model-local cache domain", index+1)
		}
		if attempt.InputTokens < 0 || attempt.CachedInputTokens < 0 || attempt.OutputTokens < 0 || attempt.CachedInputTokens > attempt.InputTokens {
			return fmt.Errorf("attempt %d has invalid token accounting", index+1)
		}
		if attempt.ContextBytes != worker.AllocatedContextBytes || attempt.StablePrefixBytes != worker.StablePrefixBytes || attempt.TaskContractBytes != worker.AllocatedTaskContractBytes {
			return fmt.Errorf("attempt %d replay bytes do not match route payload", index+1)
		}
		if index > 0 {
			previous := receipt.Attempts[index-1]
			if !allowedFallback[attempt.Model] || previous.GenerationStarted || !allowedFailure[previous.FailureKind] {
				return fmt.Errorf("attempt %d violates generation-before-fallback policy", index+1)
			}
			if attempt.Model != previous.Model {
				modelSwitches++
			}
			if modelStrength(attempt.Model) <= modelStrength(previous.Model) {
				return fmt.Errorf("attempt %d violates upgrade-only model policy", index+1)
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
		totalContract += attempt.TaskContractBytes
	}
	if receipt.ModelSwitchCount != modelSwitches || modelSwitches > worker.MaxModelSwitches {
		return fmt.Errorf("worker receipt model switch count violates route policy")
	}

	last := receipt.Attempts[len(receipt.Attempts)-1]
	if last.FailureKind != "" || !last.GenerationStarted || !validResultHandle(receipt.ResultHandle) {
		return fmt.Errorf("worker receipt has no successful generated result")
	}
	if receipt.EffectiveModel != last.Model || receipt.RetryCount != len(receipt.Attempts)-1 {
		return fmt.Errorf("worker receipt effective model or retry count is inconsistent")
	}
	if receipt.StablePrefixBytesSent != totalPrefix || receipt.ContextBytesSent != totalContext || receipt.TaskContractBytes != totalContract {
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
	if !receipt.AcceptedByAji {
		return fmt.Errorf("worker result was not accepted by Aji")
	}
	if strings.TrimSpace(receipt.BillingUnit) == "" || receipt.TotalCostMicrounits < 0 || receipt.AjiBaselineMicrounits <= 0 {
		return fmt.Errorf("worker receipt lacks usable billing evidence")
	}
	wantSavings := receipt.AjiBaselineMicrounits - receipt.TotalCostMicrounits
	if receipt.SavingsMicrounits != wantSavings {
		return fmt.Errorf("worker receipt savings calculation is inconsistent")
	}
	if worker.ModelClass != "sol" && receipt.SavingsMicrounits <= 0 {
		return fmt.Errorf("delegation did not beat the Aji cost baseline")
	}
	return nil
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
	case "gpt-5.6-luna":
		return 1
	case "gpt-5.6-terra":
		return 2
	case "gpt-5.6-sol":
		return 3
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
