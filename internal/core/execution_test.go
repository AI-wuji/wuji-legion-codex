package core

import "testing"

func TestValidateWorkerReceiptEnforcesOrderedAvailabilityFallbackAndSavings(t *testing.T) {
	worker := testLunaReceiptWorker()
	receipt := validReceipt(worker)
	if err := ValidateWorkerReceipt(worker, receipt); err != nil {
		t.Fatal(err)
	}

	missingDispatch := receipt
	missingDispatch.HostDispatchID = ""
	if err := ValidateWorkerReceipt(worker, missingDispatch); err == nil {
		t.Fatal("receipt without a host dispatch identifier was accepted")
	}

	extraAttempt := receipt
	extraAttempt.Attempts = append(append([]WorkerAttempt(nil), receipt.Attempts...), WorkerAttempt{Model: worker.Model, GenerationStarted: true, CacheDomain: "model-local:" + worker.Model, ContextBytes: worker.AllocatedContextBytes, StablePrefixBytes: worker.StablePrefixBytes, SourceExecutionBytes: worker.SourceExecutionBytes, TaskContractBytes: worker.AllocatedTaskContractBytes})
	if err := ValidateWorkerReceipt(worker, extraAttempt); err == nil {
		t.Fatal("out-of-order worker attempt was accepted")
	}

	noSaving := receipt
	noSaving.TotalCostMicrounits = noSaving.ExecutionBaselineMicrounits
	noSaving.SavingsMicrounits = 0
	if err := ValidateWorkerReceipt(worker, noSaving); err == nil {
		t.Fatal("non-saving delegation was accepted")
	}
}

func TestValidateWorkerReceiptEnforcesSessionAndGPTWorkerAllowlist(t *testing.T) {
	worker := testReceiptWorker()
	receipt := validReceipt(worker)

	wrongSession := receipt
	wrongSession.SessionKey = "wuji-session://sha256/wrong"
	if err := ValidateWorkerReceipt(worker, wrongSession); err == nil {
		t.Fatal("receipt with a different task session was accepted")
	}

	wrongSwitchCount := receipt
	wrongSwitchCount.ModelSwitchCount = 2
	if err := ValidateWorkerReceipt(worker, wrongSwitchCount); err == nil {
		t.Fatal("receipt with false model-switch telemetry was accepted")
	}

	legacyWorker := worker
	legacyWorker.Model = "gpt-5.4"
	legacyReceipt := validReceipt(legacyWorker)
	if err := ValidateWorkerReceipt(legacyWorker, legacyReceipt); err == nil {
		t.Fatal("legacy GPT worker model was accepted")
	}

	externalWorker := worker
	externalWorker.Model = "grok-4.5"
	externalReceipt := validReceipt(externalWorker)
	if err := ValidateWorkerReceipt(externalWorker, externalReceipt); err != nil {
		t.Fatalf("explicit non-GPT provider was incorrectly blocked: %v", err)
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

func TestValidateWorkerReceiptEnforcesExecutionNodeWriteBoundary(t *testing.T) {
	worker := testReceiptWorker()
	worker.Writes = true
	receipt := validReceipt(worker)
	receipt.WriteBoundary = "scoped-artifact-write"
	if err := ValidateWorkerReceipt(worker, receipt); err != nil {
		t.Fatalf("scoped execution-node artifact write was rejected: %v", err)
	}

	receipt.WriteBoundary = "read-only"
	if err := ValidateWorkerReceipt(worker, receipt); err == nil {
		t.Fatal("write receipt outside its execution-node boundary was accepted")
	}
}

func TestValidateSolJudgmentReceiptAllowsBoundedCostEscalation(t *testing.T) {
	worker := Route("architecture decision: use Sol", nil).Workers[0]
	receipt := validReceipt(worker)
	receipt.TotalCostMicrounits = 160
	receipt.ExecutionBaselineMicrounits = 100
	receipt.SavingsMicrounits = -60
	if err := ValidateWorkerReceipt(worker, receipt); err != nil {
		t.Fatalf("bounded Sol judgment was rejected for costing more than Terra: %v", err)
	}
}

func validReceipt(worker WorkerTask) WorkerExecutionReceipt {
	attempts := []WorkerAttempt{}
	effectiveModel := worker.Model
	modelSwitches := 0
	retryCount := 0
	failureKinds := []string{}
	models := append([]string{worker.Model}, worker.AvailabilityFallbackModels...)
	for index, model := range models {
		attempt := WorkerAttempt{Model: model, CacheDomain: "model-local:" + model, ContextBytes: worker.AllocatedContextBytes, StablePrefixBytes: worker.StablePrefixBytes, SourceExecutionBytes: worker.SourceExecutionBytes, TaskContractBytes: worker.AllocatedTaskContractBytes}
		if index < len(models)-1 {
			attempt.FailureKind = "model-unavailable"
			failureKinds = append(failureKinds, attempt.FailureKind)
		} else {
			attempt.GenerationStarted = true
			attempt.InputTokens = 100
			attempt.CachedInputTokens = 80
			attempt.OutputTokens = 20
			effectiveModel = model
		}
		attempts = append(attempts, attempt)
	}
	if len(worker.AvailabilityFallbackModels) > 0 {
		modelSwitches = len(worker.AvailabilityFallbackModels)
		retryCount = len(worker.AvailabilityFallbackModels)
	}
	return WorkerExecutionReceipt{
		SchemaVersion: workerReceiptSchemaVersion, WorkerID: worker.ID, RequestedModel: worker.Model, SessionKey: worker.SessionKey,
		HostDispatchID: "codex-agent://test/worker", WriteBoundary: "read-only", Attempts: attempts,
		EffectiveModel: effectiveModel, ModelSwitchCount: modelSwitches, ResultHandle: "wuji-result://sha256/0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", ContextHandleIDs: worker.ContextHandles,
		StablePrefixBytesSent: worker.StablePrefixBytes * len(attempts), StablePrefixSHA256: worker.StablePrefixSHA256,
		SourceExecutionBytesSent: worker.SourceExecutionBytes * len(attempts),
		ContextBytesSent:         worker.AllocatedContextBytes * len(attempts), ContextPayloadSHA256: worker.ContextPayloadSHA256,
		TaskContractBytes: worker.AllocatedTaskContractBytes * len(attempts), TaskContractSHA256: worker.TaskContractSHA256,
		InputTokens: 100, CachedInputTokens: 80, OutputTokens: 20, RetryCount: retryCount,
		AttemptFailureKinds: failureKinds, CacheDomain: "model-local:" + effectiveModel, DelegationGateReason: worker.DelegationGateReason,
		BillingUnit: "usd-micro", TotalCostMicrounits: 60, ExecutionBaselineMicrounits: 100, SavingsMicrounits: 40,
	}
}

func TestValidateWorkerReceiptRejectsFallbackAfterGeneration(t *testing.T) {
	worker := testLunaReceiptWorker()
	receipt := validReceipt(worker)
	receipt.Attempts[0].GenerationStarted = true
	if err := ValidateWorkerReceipt(worker, receipt); err == nil {
		t.Fatal("fallback after generation was accepted")
	}
}

func TestValidateWorkerReceiptRejectsInvalidFallbackOrder(t *testing.T) {
	worker := testLunaReceiptWorker()
	receipt := validReceipt(worker)
	receipt.Attempts[1].Model = worker.AvailabilityFallbackModels[1]
	receipt.Attempts[1].CacheDomain = "model-local:" + receipt.Attempts[1].Model
	if err := ValidateWorkerReceipt(worker, receipt); err == nil {
		t.Fatal("non-ascending fallback order was accepted")
	}
}

func TestValidateWorkerReceiptRejectsSkippedFallbackModel(t *testing.T) {
	worker := testLunaReceiptWorker()
	receipt := validReceipt(worker)
	// A provider can return unavailable before generation, but it cannot skip a
	// declared availability rung to report a stronger model as the next attempt.
	receipt.Attempts = receipt.Attempts[:2]
	receipt.Attempts[1].Model = worker.AvailabilityFallbackModels[1]
	receipt.Attempts[1].CacheDomain = "model-local:" + receipt.Attempts[1].Model
	receipt.EffectiveModel = receipt.Attempts[1].Model
	receipt.ModelSwitchCount = 1
	receipt.RetryCount = 1
	receipt.StablePrefixBytesSent = worker.StablePrefixBytes * len(receipt.Attempts)
	receipt.SourceExecutionBytesSent = worker.SourceExecutionBytes * len(receipt.Attempts)
	receipt.ContextBytesSent = worker.AllocatedContextBytes * len(receipt.Attempts)
	receipt.TaskContractBytes = worker.AllocatedTaskContractBytes * len(receipt.Attempts)
	if err := ValidateWorkerReceipt(worker, receipt); err == nil {
		t.Fatal("receipt that skipped a declared fallback model was accepted")
	}
}

func testReceiptWorker() WorkerTask {
	query := "code task parallel"
	items := []Manifest{{ID: "code", Triggers: []string{"code"}, Status: "callable", PrimarySkill: "native", Experts: []Expert{{ID: "implementation", Independent: true, ModelClass: "terra"}}}}
	return RouteWithContext(query, items, delegationContextForTest(query, 512)).Workers[0]
}

func testLunaReceiptWorker() WorkerTask {
	return RouteWithContext("list files and count occurrences", nil, DelegationContext{SelfContained: true}).Workers[0]
}
