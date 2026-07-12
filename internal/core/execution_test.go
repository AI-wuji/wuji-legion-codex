package core

import "testing"

func TestValidateWorkerReceiptEnforcesFallbackAndSavings(t *testing.T) {
	worker := testReceiptWorker()
	receipt := validReceipt(worker)
	if err := ValidateWorkerReceipt(worker, receipt); err != nil {
		t.Fatal(err)
	}

	paidRetry := receipt
	paidRetry.Attempts = append([]WorkerAttempt(nil), receipt.Attempts...)
	paidRetry.Attempts[0].GenerationStarted = true
	if err := ValidateWorkerReceipt(worker, paidRetry); err == nil {
		t.Fatal("paid retry was accepted")
	}

	noSaving := receipt
	noSaving.TotalCostMicrounits = noSaving.AjiBaselineMicrounits
	noSaving.SavingsMicrounits = 0
	if err := ValidateWorkerReceipt(worker, noSaving); err == nil {
		t.Fatal("non-saving delegation was accepted")
	}
}

func validReceipt(worker WorkerTask) WorkerExecutionReceipt {
	fallback := worker.FallbackModels[0]
	return WorkerExecutionReceipt{
		SchemaVersion: workerReceiptSchemaVersion, WorkerID: worker.ID, RequestedModel: worker.Model,
		Attempts: []WorkerAttempt{
			{Model: worker.Model, FailureKind: "model-unavailable", CacheDomain: "model-local:" + worker.Model, ContextBytes: worker.AllocatedContextBytes, StablePrefixBytes: worker.StablePrefixBytes, TaskContractBytes: worker.AllocatedTaskContractBytes},
			{Model: fallback, GenerationStarted: true, InputTokens: 100, CachedInputTokens: 80, OutputTokens: 20, CacheDomain: "model-local:" + fallback, ContextBytes: worker.AllocatedContextBytes, StablePrefixBytes: worker.StablePrefixBytes, TaskContractBytes: worker.AllocatedTaskContractBytes},
		},
		EffectiveModel: fallback, ResultHandle: "wuji-result://sha256/result", ContextHandleIDs: worker.ContextHandles,
		StablePrefixBytesSent: worker.StablePrefixBytes * 2, StablePrefixSHA256: worker.StablePrefixSHA256,
		ContextBytesSent: worker.AllocatedContextBytes * 2, ContextPayloadSHA256: worker.ContextPayloadSHA256,
		TaskContractBytes: worker.AllocatedTaskContractBytes * 2, TaskContractSHA256: worker.TaskContractSHA256,
		InputTokens: 100, CachedInputTokens: 80, OutputTokens: 20, RetryCount: 1, AcceptedByAji: true,
		AttemptFailureKinds: []string{"model-unavailable"}, CacheDomain: "model-local:" + fallback, DelegationGateReason: worker.DelegationGateReason,
		BillingUnit: "usd-micro", TotalCostMicrounits: 60, AjiBaselineMicrounits: 100, SavingsMicrounits: 40,
	}
}

func testReceiptWorker() WorkerTask {
	query := "code task parallel"
	items := []Manifest{{ID: "code", Triggers: []string{"code"}, Status: "callable", PrimarySkill: "native", Experts: []Expert{{ID: "implementation", Independent: true, ModelClass: "terra"}}}}
	return RouteWithContext(query, items, delegationContextForTest(query, 512)).Workers[0]
}
