package core

import "testing"

func TestValidateWorkerReceiptEnforcesFallbackAndSavings(t *testing.T) {
	worker := testReceiptWorker()
	receipt := validReceipt(worker)
	if err := ValidateWorkerReceipt(worker, receipt); err != nil {
		t.Fatal(err)
	}

	missingDispatch := receipt
	missingDispatch.HostDispatchID = ""
	if err := ValidateWorkerReceipt(worker, missingDispatch); err == nil {
		t.Fatal("receipt without a host dispatch identifier was accepted")
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

func TestValidateWorkerReceiptEnforcesSessionAndUpgradeOnlySwitching(t *testing.T) {
	worker := testReceiptWorker()
	receipt := validReceipt(worker)

	wrongSession := receipt
	wrongSession.SessionKey = "wuji-session://sha256/wrong"
	if err := ValidateWorkerReceipt(worker, wrongSession); err == nil {
		t.Fatal("receipt with a different task session was accepted")
	}

	wrongSwitchCount := receipt
	wrongSwitchCount.ModelSwitchCount = 0
	if err := ValidateWorkerReceipt(worker, wrongSwitchCount); err == nil {
		t.Fatal("receipt with false model-switch telemetry was accepted")
	}

	downgradeWorker := worker
	downgradeWorker.Model = "gpt-5.6-sol"
	downgradeWorker.FallbackModels = []string{"gpt-5.6-luna"}
	downgradeReceipt := validReceipt(downgradeWorker)
	if err := ValidateWorkerReceipt(downgradeWorker, downgradeReceipt); err == nil {
		t.Fatal("Sol-to-Luna downgrade was accepted")
	}
}

func TestValidateWorkerReceiptRequiresContentAddressedResult(t *testing.T) {
	worker := testReceiptWorker()
	receipt := validReceipt(worker)
	receipt.ResultHandle = "completed"
	if err := ValidateWorkerReceipt(worker, receipt); err == nil {
		t.Fatal("arbitrary result claim was accepted as a receipt")
	}
}

func TestValidateSolJudgmentReceiptAllowsBoundedCostEscalation(t *testing.T) {
	worker := Route("architecture decision: use Sol", nil).Workers[0]
	receipt := validReceipt(worker)
	receipt.TotalCostMicrounits = 160
	receipt.AjiBaselineMicrounits = 100
	receipt.SavingsMicrounits = -60
	if err := ValidateWorkerReceipt(worker, receipt); err != nil {
		t.Fatalf("bounded Sol judgment was rejected for costing more than Terra: %v", err)
	}
}

func validReceipt(worker WorkerTask) WorkerExecutionReceipt {
	attempts := []WorkerAttempt{{Model: worker.Model, GenerationStarted: true, InputTokens: 100, CachedInputTokens: 80, OutputTokens: 20, CacheDomain: "model-local:" + worker.Model, ContextBytes: worker.AllocatedContextBytes, StablePrefixBytes: worker.StablePrefixBytes, TaskContractBytes: worker.AllocatedTaskContractBytes}}
	effectiveModel := worker.Model
	modelSwitches := 0
	retryCount := 0
	failureKinds := []string{}
	if len(worker.FallbackModels) > 0 {
		fallback := worker.FallbackModels[0]
		attempts = []WorkerAttempt{
			{Model: worker.Model, FailureKind: "model-unavailable", CacheDomain: "model-local:" + worker.Model, ContextBytes: worker.AllocatedContextBytes, StablePrefixBytes: worker.StablePrefixBytes, TaskContractBytes: worker.AllocatedTaskContractBytes},
			{Model: fallback, GenerationStarted: true, InputTokens: 100, CachedInputTokens: 80, OutputTokens: 20, CacheDomain: "model-local:" + fallback, ContextBytes: worker.AllocatedContextBytes, StablePrefixBytes: worker.StablePrefixBytes, TaskContractBytes: worker.AllocatedTaskContractBytes},
		}
		effectiveModel = fallback
		modelSwitches = 1
		retryCount = 1
		failureKinds = []string{"model-unavailable"}
	}
	return WorkerExecutionReceipt{
		SchemaVersion: workerReceiptSchemaVersion, WorkerID: worker.ID, RequestedModel: worker.Model, SessionKey: worker.SessionKey,
		HostDispatchID: "codex-agent://test/worker", WriteBoundary: "read-only", Attempts: attempts,
		EffectiveModel: effectiveModel, ModelSwitchCount: modelSwitches, ResultHandle: "wuji-result://sha256/0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", ContextHandleIDs: worker.ContextHandles,
		StablePrefixBytesSent: worker.StablePrefixBytes * len(attempts), StablePrefixSHA256: worker.StablePrefixSHA256,
		ContextBytesSent: worker.AllocatedContextBytes * len(attempts), ContextPayloadSHA256: worker.ContextPayloadSHA256,
		TaskContractBytes: worker.AllocatedTaskContractBytes * len(attempts), TaskContractSHA256: worker.TaskContractSHA256,
		InputTokens: 100, CachedInputTokens: 80, OutputTokens: 20, RetryCount: retryCount, AcceptedByAji: true,
		AttemptFailureKinds: failureKinds, CacheDomain: "model-local:" + effectiveModel, DelegationGateReason: worker.DelegationGateReason,
		BillingUnit: "usd-micro", TotalCostMicrounits: 60, AjiBaselineMicrounits: 100, SavingsMicrounits: 40,
	}
}

func testReceiptWorker() WorkerTask {
	query := "code task parallel"
	items := []Manifest{{ID: "code", Triggers: []string{"code"}, Status: "callable", PrimarySkill: "native", Experts: []Expert{{ID: "implementation", Independent: true, ModelClass: "terra"}}}}
	return RouteWithContext(query, items, delegationContextForTest(query, 512)).Workers[0]
}
