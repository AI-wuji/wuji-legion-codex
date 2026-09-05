package core

type Source struct {
	ID         string   `json:"id"`
	Engine     string   `json:"engine,omitempty"`
	Priority   string   `json:"priority,omitempty"`
	Lifecycle  string   `json:"lifecycle,omitempty"`
	Version    string   `json:"version,omitempty"`
	Revision   string   `json:"revision,omitempty"`
	ReleaseID  string   `json:"release_id,omitempty"`
	License    string   `json:"license,omitempty"`
	Activation []string `json:"activation,omitempty"`
	Entrypoint string   `json:"entrypoint,omitempty"`
	Fallback   string   `json:"fallback,omitempty"`
	Globs      []string `json:"globs"`
	Required   []string `json:"required"`
}

// FusionGenome is an optional, declarative description of how a capability
// composes existing callable sources. Existing manifests remain valid without
// one and can be introduced incrementally.
type FusionGenome struct {
	SchemaVersion int             `json:"schema_version"`
	Species       string          `json:"species"`
	Revision      string          `json:"revision"`
	ReleaseID     string          `json:"release_id,omitempty"`
	Generation    int             `json:"generation,omitempty"`
	Parents       []string        `json:"parents,omitempty"`
	Adapters      []FusionAdapter `json:"adapters"`
}

// FusionAdapter binds one source entrypoint and its retained assets to a
// domain. It never grants an execution host new filesystem authority.
type FusionAdapter struct {
	ID             string        `json:"id"`
	Domain         string        `json:"domain"`
	Source         string        `json:"source"`
	Entrypoint     string        `json:"entrypoint"`
	SourceVersion  string        `json:"source_version,omitempty"`
	AtomRevision   string        `json:"atom_revision,omitempty"`
	ReleaseID      string        `json:"release_id,omitempty"`
	License        string        `json:"license,omitempty"`
	Assets         []string      `json:"assets,omitempty"`
	AssetContracts []FusionAsset `json:"asset_contracts,omitempty"`
}

type FusionAsset struct {
	ID            string   `json:"asset_id"`
	Path          string   `json:"path"`
	Compatibility []string `json:"compatibility,omitempty"`
}

type FusionAdapterVerification struct {
	ID               string `json:"id"`
	Source           string `json:"source"`
	Entrypoint       string `json:"entrypoint"`
	EntrypointSHA256 string `json:"entrypoint_sha256,omitempty"`
	Reachable        bool   `json:"reachable"`
	Reason           string `json:"reason,omitempty"`
}

type FusionGenomeVerification struct {
	Species  string                      `json:"species"`
	Revision string                      `json:"revision"`
	Adapters []FusionAdapterVerification `json:"adapters"`
}

type Probe struct {
	Command            string   `json:"command"`
	Args               []string `json:"args"`
	Fixture            string   `json:"fixture,omitempty"`
	Kind               string   `json:"kind,omitempty"`
	TimeoutSeconds     int      `json:"timeout_seconds,omitempty"`
	RequiredEvidence   []string `json:"required_evidence,omitempty"`
	ComparisonEvidence string   `json:"comparison_evidence,omitempty"`
}

type ProbeArtifact struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size,omitempty"`
}

type ProbeEvidence struct {
	WujiProbe string          `json:"wuji_probe"`
	Fixture   string          `json:"fixture"`
	Passed    bool            `json:"passed"`
	Evidence  []ProbeArtifact `json:"evidence"`
	Signature string          `json:"signature"`
}

type PromotionArtifact struct {
	ID     string `json:"id"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

type PromotionProof struct {
	ContractSHA256     string            `json:"contract_sha256"`
	EffectiveStatus    string            `json:"effective_status"`
	Signature          string            `json:"signature"`
	ComparisonEvidence PromotionArtifact `json:"comparison_evidence"`
}

type PromotionReceipt struct {
	SchemaVersion    int            `json:"schema_version"`
	Capability       string         `json:"capability"`
	Decision         string         `json:"decision"`
	Fixture          string         `json:"fixture"`
	BaselineManifest string         `json:"baseline_manifest"`
	Baseline         PromotionProof `json:"baseline"`
	Candidate        PromotionProof `json:"candidate"`
}

type Expert struct {
	ID          string `json:"id"`
	Purpose     string `json:"purpose"`
	Independent bool   `json:"independent"`
	ModelClass  string `json:"model_class"`
}

type Provider struct {
	ID       string   `json:"id"`
	Default  bool     `json:"default"`
	Triggers []string `json:"triggers,omitempty"`
	Fallback string   `json:"fallback,omitempty"`
}

type Engine struct {
	ID           string   `json:"id"`
	Default      bool     `json:"default"`
	Triggers     []string `json:"triggers,omitempty"`
	PrimarySkill string   `json:"primary_skill,omitempty"`
}

type Manifest struct {
	Root             string        `json:"-"`
	ID               string        `json:"id"`
	Description      string        `json:"description"`
	Triggers         []string      `json:"triggers"`
	Status           string        `json:"status"`
	PromotionReceipt string        `json:"promotion_receipt,omitempty"`
	PrimarySkill     string        `json:"primary_skill"`
	HostCallable     bool          `json:"host_callable"`
	DirectMount      bool          `json:"direct_mount"`
	Fallback         string        `json:"fallback"`
	Sources          []Source      `json:"sources"`
	Genome           *FusionGenome `json:"fusion_genome,omitempty"`
	Probe            *Probe        `json:"probe,omitempty"`
	Experts          []Expert      `json:"experts,omitempty"`
	Providers        []Provider    `json:"providers,omitempty"`
	Engines          []Engine      `json:"engines,omitempty"`
}

type MountedSource struct {
	ID               string `json:"id"`
	Path             string `json:"path"`
	Priority         string `json:"priority,omitempty"`
	Lifecycle        string `json:"lifecycle"`
	Entrypoint       string `json:"entrypoint,omitempty"`
	ActivationReason string `json:"activation_reason,omitempty"`
}

// SourceExecutionContract is the bounded, content-addressed instruction
// surface that turns a routed source into something the execution host can
// actually load. EntrypointContent is deliberately excluded from route JSON.
type SourceExecutionContract struct {
	SourceID         string `json:"source_id"`
	Capability       string `json:"capability"`
	InvocationKind   string `json:"invocation_kind"`
	Entrypoint       string `json:"entrypoint"`
	EntrypointSHA256 string `json:"entrypoint_sha256"`
	EntrypointBytes  int    `json:"entrypoint_bytes"`
	ActivationReason string `json:"activation_reason"`
	// ResolvedEntrypointPath is host-only. Route JSON never grants filesystem
	// authority; dispatch resolves this again from trusted manifests.
	ResolvedEntrypointPath string `json:"-"`
	EntrypointContent      string `json:"-"`
}

// ResponsePolicyContract is a bounded, content-addressed rule overlay for
// Aji's user-facing response. Unlike SourceExecutionContract, it applies to
// the final writer and must not replace the task's domain capability.
type ResponsePolicyContract struct {
	ID               string              `json:"id"`
	Revision         string              `json:"revision"`
	Active           bool                `json:"active"`
	ActivationReason string              `json:"activation_reason"`
	SourceCommit     string              `json:"source_commit"`
	RulesSHA256      string              `json:"rules_sha256"`
	RulesBytes       int                 `json:"rules_bytes"`
	Precedence       []string            `json:"precedence"`
	Directives       []ResponseDirective `json:"directives,omitempty"`
	Suppressed       []string            `json:"suppressed,omitempty"`
	ExitTriggers     []string            `json:"exit_triggers"`
}

type ResponseDirective struct {
	ID         string         `json:"id"`
	Phase      string         `json:"phase"`
	Priority   int            `json:"priority"`
	Conditions []string       `json:"conditions"`
	Directive  string         `json:"directive"`
	Overrides  []string       `json:"overrides,omitempty"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

// SourceEntrypointVerification proves only that the dispatcher resolved the
// routed entrypoint. It is never proof that a Skill or MCP executed.
type SourceEntrypointVerification struct {
	SourceID         string `json:"source_id"`
	Capability       string `json:"capability"`
	InvocationKind   string `json:"invocation_kind"`
	Entrypoint       string `json:"entrypoint"`
	EntrypointSHA256 string `json:"entrypoint_sha256"`
	EntrypointBytes  int    `json:"entrypoint_bytes"`
}

// SourceAuditEntry distinguishes an executable routing atom from retained
// upstream material. It deliberately does not claim that a host executed it.
type SourceAuditEntry struct {
	Capability          string              `json:"capability"`
	Source              string              `json:"source"`
	Lifecycle           string              `json:"lifecycle"`
	State               string              `json:"state"`
	ExecutionMode       string              `json:"execution_mode"`
	ExecutionEvidence   string              `json:"execution_evidence"`
	Entrypoint          string              `json:"entrypoint,omitempty"`
	EntrypointReachable bool                `json:"entrypoint_reachable"`
	EntrypointSHA256    string              `json:"entrypoint_sha256,omitempty"`
	Assets              []AssetReachability `json:"assets,omitempty"`
	BehaviorCoverage    string              `json:"behavior_coverage"`
	Reason              string              `json:"reason"`
}

// AssetReachability reports only whether a retained asset resolves to a local
// trusted path. It is deliberately separate from behavior-probe evidence.
type AssetReachability struct {
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	Reachable bool   `json:"reachable"`
	SHA256    string `json:"sha256,omitempty"`
	Bytes     int64  `json:"bytes,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

type LineageNode struct {
	ID             string   `json:"id"`
	Kind           string   `json:"kind"`
	Capability     string   `json:"capability,omitempty"`
	SourceID       string   `json:"source_id,omitempty"`
	AssetID        string   `json:"asset_id,omitempty"`
	Entrypoint     string   `json:"entrypoint,omitempty"`
	Compatibility  []string `json:"compatibility,omitempty"`
	Parents        []string `json:"parents,omitempty"`
	SHA256         string   `json:"sha256,omitempty"`
	Path           string   `json:"path,omitempty"`
	State          string   `json:"state"`
	SourceVersion  string   `json:"source_version,omitempty"`
	AtomRevision   string   `json:"atom_revision,omitempty"`
	Species        string   `json:"species,omitempty"`
	FusionRevision string   `json:"fusion_revision,omitempty"`
	ReleaseID      string   `json:"release_id,omitempty"`
	Generation     int      `json:"generation,omitempty"`
	License        string   `json:"license,omitempty"`
}

// LineageRejection retains only the smallest reviewable reason for an
// unadmitted source. It intentionally excludes transcripts and raw logs.
type LineageRejection struct {
	ID            string `json:"id"`
	Source        string `json:"source"`
	Reason        string `json:"reason"`
	SHA256        string `json:"sha256"`
	SourceVersion string `json:"source_version,omitempty"`
	AtomRevision  string `json:"atom_revision,omitempty"`
	ReleaseID     string `json:"release_id,omitempty"`
	State         string `json:"state"`
}

type LineageCatalog struct {
	SchemaVersion int                `json:"schema_version"`
	Nodes         []LineageNode      `json:"nodes"`
	Rejections    []LineageRejection `json:"rejections,omitempty"`
}

type LineageSyncResult struct {
	CatalogPath    string         `json:"catalog_path"`
	CatalogSHA256  string         `json:"catalog_sha256"`
	NodeCount      int            `json:"node_count"`
	RejectionCount int            `json:"rejection_count"`
	Catalog        LineageCatalog `json:"catalog"`
}

type AssetSelectionRequest struct {
	Capability    string   `json:"capability,omitempty"`
	Domain        string   `json:"domain,omitempty"`
	AssetID       string   `json:"asset_id,omitempty"`
	Compatibility []string `json:"compatibility,omitempty"`
}

type AssetInvocationContract struct {
	AssetID          string                  `json:"asset_id"`
	Capability       string                  `json:"capability"`
	AdapterID        string                  `json:"adapter_id"`
	Domain           string                  `json:"domain"`
	SourceID         string                  `json:"source_id"`
	Entrypoint       string                  `json:"entrypoint"`
	EntrypointSHA256 string                  `json:"entrypoint_sha256"`
	AssetPath        string                  `json:"asset_path"`
	AssetSHA256      string                  `json:"asset_sha256"`
	AssetBytes       int64                   `json:"asset_bytes"`
	Compatibility    []string                `json:"compatibility,omitempty"`
	Invocation       SourceExecutionContract `json:"invocation"`
}

// ExecutionGraph is the durable execution contract kept separate from the
// requirement graph. Nodes reference exact requirement revisions and carry
// bounded host, model, dependency, and verification policies.
type ExecutionGraph struct {
	SchemaVersion int                  `json:"schema_version"`
	Nodes         []ExecutionGraphNode `json:"nodes"`
}

type ExecutionGraphNode struct {
	ID                   string   `json:"id"`
	VersionID            string   `json:"version_id"`
	Revision             int      `json:"revision"`
	Status               string   `json:"status"`
	Authority            string   `json:"authority"`
	Goal                 string   `json:"goal"`
	Avoids               []string `json:"avoids,omitempty"`
	RequirementRevisions []string `json:"requirement_revisions"`
	DependsOn            []string `json:"depends_on,omitempty"`
	Inputs               []string `json:"inputs,omitempty"`
	AllowedContext       []string `json:"allowed_context,omitempty"`
	Outputs              []string `json:"outputs,omitempty"`
	Model                string   `json:"model"`
	ModelReason          string   `json:"model_reason"`
	Acceptance           []string `json:"acceptance,omitempty"`
	Verification         []string `json:"verification,omitempty"`
	EvidenceRequired     []string `json:"evidence_required,omitempty"`
	TimeBudgetSeconds    int      `json:"time_budget_seconds,omitempty"`
	CostBudgetMicrounits int64    `json:"cost_budget_microunits,omitempty"`
	MaxAttempts          int      `json:"max_attempts,omitempty"`
	NetworkBoundary      string   `json:"network_boundary"`
	WriteBoundary        string   `json:"write_boundary"`
	BranchBoundary       string   `json:"branch_boundary"`
	Failure              string   `json:"failure,omitempty"`
	Recovery             string   `json:"recovery,omitempty"`
	ArtifactHandles      []string `json:"artifact_handles,omitempty"`
	VerificationHandles  []string `json:"verification_handles,omitempty"`
	Supersedes           string   `json:"supersedes,omitempty"`
}

type ExecutionNodeInput struct {
	ID                   string
	Authority            string
	Goal                 string
	Avoids               []string
	RequirementRevisions []string
	DependsOn            []string
	Inputs               []string
	AllowedContext       []string
	Outputs              []string
	Model                string
	ModelReason          string
	Acceptance           []string
	Verification         []string
	EvidenceRequired     []string
	TimeBudgetSeconds    int
	CostBudgetMicrounits int64
	MaxAttempts          int
	NetworkBoundary      string
	WriteBoundary        string
	BranchBoundary       string
	Failure              string
	Recovery             string
}

type ExecutionResultInput struct {
	ID                  string
	Status              string
	Failure             string
	Recovery            string
	ArtifactHandles     []string
	VerificationHandles []string
}

type ExecutionGraphProjection struct {
	SchemaVersion int                  `json:"schema_version"`
	Handle        string               `json:"handle"`
	PayloadSHA256 string               `json:"payload_sha256"`
	PayloadBytes  int                  `json:"payload_bytes"`
	Target        string               `json:"target"`
	Nodes         []ExecutionGraphNode `json:"nodes"`
	Payload       string               `json:"payload"`
	ArtifactPath  string               `json:"artifact_path,omitempty"`
}

type ExecutionGraphResult struct {
	Graph        ExecutionGraph            `json:"graph"`
	Node         ExecutionGraphNode        `json:"node,omitempty"`
	Projection   *ExecutionGraphProjection `json:"projection,omitempty"`
	ArtifactPath string                    `json:"artifact_path,omitempty"`
}

type AuditEvent struct {
	EventID         string   `json:"event_id"`
	EventType       string   `json:"event_type"`
	Actor           string   `json:"actor"`
	Authority       string   `json:"authority"`
	Target          string   `json:"target"`
	InputRevision   string   `json:"input_revision,omitempty"`
	ResultHandle    string   `json:"result_handle,omitempty"`
	EvidenceHandles []string `json:"evidence_handles,omitempty"`
	RecordSHA256    string   `json:"record_sha256"`
	OccurredAt      string   `json:"occurred_at"`
}

type SecurityAction struct {
	Kind               string `json:"kind"`
	Target             string `json:"target"`
	Workspace          string `json:"workspace,omitempty"`
	ExplicitUserIntent bool   `json:"explicit_user_intent"`
}

type SecurityGateResult struct {
	Action   SecurityAction `json:"action"`
	Decision string         `json:"decision"`
	Allowed  bool           `json:"allowed"`
	Checks   []string       `json:"checks"`
	Reason   string         `json:"reason,omitempty"`
}

type OfficerContract struct {
	Role                     string   `json:"role"`
	TaskTypes                []string `json:"task_types"`
	Stages                   []string `json:"stages"`
	RiskSignals              []string `json:"risk_signals"`
	ArtifactTypes            []string `json:"artifact_types"`
	EvidenceGaps             []string `json:"evidence_gaps"`
	RequiresUserConfirmation bool     `json:"requires_user_confirmation"`
}

type OfficerRecommendation struct {
	Role     string          `json:"role"`
	Decision string          `json:"decision"`
	Reason   string          `json:"reason"`
	Contract OfficerContract `json:"contract"`
}

type WorkerTask struct {
	ID         string `json:"id"`
	Stage      string `json:"stage"`
	Purpose    string `json:"purpose"`
	ModelClass string `json:"model_class"`
	Model      string `json:"model"`
	// AvailabilityFallbackModels is an ordered model-selection chain that the
	// host may consult before generation starts. It is deliberately separate
	// from the execution retry fields below: a worker execution is always one
	// generation attempt once a model has been selected.
	AvailabilityFallbackModels []string                  `json:"availability_fallback_models,omitempty"`
	AvailabilityFallbackOn     []string                  `json:"availability_fallback_on,omitempty"`
	FallbackModels             []string                  `json:"fallback_models,omitempty"`
	SessionKey                 string                    `json:"session_key"`
	SessionAffinity            string                    `json:"session_affinity"`
	EscalationPolicy           string                    `json:"escalation_policy"`
	MaxModelSwitches           int                       `json:"max_model_switches"`
	MaxSources                 int                       `json:"max_sources,omitempty"`
	TimeBudgetSeconds          int                       `json:"time_budget_seconds,omitempty"`
	StopConditions             []string                  `json:"stop_conditions,omitempty"`
	Inputs                     []string                  `json:"inputs"`
	Protocol                   []string                  `json:"protocol,omitempty"`
	TaskContract               string                    `json:"task_contract"`
	TaskContractSHA256         string                    `json:"task_contract_sha256"`
	ContextMode                string                    `json:"context_mode"`
	ContextHandles             []string                  `json:"context_handles,omitempty"`
	ContextArtifact            string                    `json:"context_artifact,omitempty"`
	ContextPayload             string                    `json:"context_payload,omitempty"`
	ContextPayloadSHA256       string                    `json:"context_payload_sha256,omitempty"`
	StableCapabilityPrefix     string                    `json:"stable_capability_prefix"`
	StablePrefixSHA256         string                    `json:"stable_prefix_sha256"`
	StablePrefixBytes          int                       `json:"stable_prefix_bytes"`
	SourceExecution            []SourceExecutionContract `json:"source_execution,omitempty"`
	SourceExecutionBytes       int                       `json:"source_execution_bytes"`
	AssetContracts             []AssetInvocationContract `json:"asset_contracts,omitempty"`
	PromptOrder                []string                  `json:"prompt_order"`
	AllocatedContextBytes      int                       `json:"allocated_context_bytes"`
	AllocatedTaskContractBytes int                       `json:"allocated_task_contract_bytes"`
	MaxTaskContractBytes       int                       `json:"max_task_contract_bytes"`
	DelegationGateReason       string                    `json:"delegation_gate_reason"`
	MaxAttempts                int                       `json:"max_attempts"`
	FallbackOn                 []string                  `json:"fallback_on"`
	Writes                     bool                      `json:"writes"`
	// These fields define the required host-attestation shape. They do not make
	// a route or caller-supplied receipt into proof that a worker executed.
	ExecutionEvidenceRequired bool     `json:"execution_evidence_required"`
	ExecutionEvidenceFields   []string `json:"execution_evidence_fields"`
}

type WorkerAttempt struct {
	Model                string `json:"model"`
	FailureKind          string `json:"failure_kind,omitempty"`
	GenerationStarted    bool   `json:"generation_started"`
	InputTokens          int    `json:"input_tokens"`
	CachedInputTokens    int    `json:"cached_input_tokens"`
	OutputTokens         int    `json:"output_tokens"`
	ContextBytes         int    `json:"context_bytes"`
	StablePrefixBytes    int    `json:"stable_prefix_bytes"`
	SourceExecutionBytes int    `json:"source_execution_bytes"`
	TaskContractBytes    int    `json:"task_contract_bytes"`
	CacheDomain          string `json:"cache_domain"`
}

type WorkerExecutionReceipt struct {
	SchemaVersion               int             `json:"schema_version"`
	WorkerID                    string          `json:"worker_id"`
	RequestedModel              string          `json:"requested_model"`
	SessionKey                  string          `json:"session_key"`
	HostDispatchID              string          `json:"host_dispatch_id"`
	WriteBoundary               string          `json:"write_boundary"`
	Attempts                    []WorkerAttempt `json:"attempts"`
	EffectiveModel              string          `json:"effective_model"`
	ModelSwitchCount            int             `json:"model_switch_count"`
	ResultHandle                string          `json:"result_handle"`
	ContextHandleIDs            []string        `json:"context_handle_ids"`
	StablePrefixBytesSent       int             `json:"stable_prefix_bytes"`
	StablePrefixSHA256          string          `json:"stable_prefix_sha256"`
	SourceExecutionBytesSent    int             `json:"source_execution_bytes"`
	ContextBytesSent            int             `json:"context_bytes_sent"`
	ContextPayloadSHA256        string          `json:"context_payload_sha256"`
	TaskContractBytes           int             `json:"task_contract_bytes"`
	TaskContractSHA256          string          `json:"task_contract_sha256"`
	InputTokens                 int             `json:"input_tokens"`
	CachedInputTokens           int             `json:"cached_input_tokens"`
	OutputTokens                int             `json:"output_tokens"`
	RetryCount                  int             `json:"retry_count"`
	AttemptFailureKinds         []string        `json:"attempt_failure_kinds"`
	CacheDomain                 string          `json:"cache_domain"`
	DelegationGateReason        string          `json:"delegation_gate_reason"`
	BillingUnit                 string          `json:"billing_unit"`
	TotalCostMicrounits         int64           `json:"total_cost_microunits"`
	ExecutionBaselineMicrounits int64           `json:"execution_baseline_microunits"`
	SavingsMicrounits           int64           `json:"savings_microunits"`
}

type DelegationPolicy struct {
	CrossModelCacheAssumed          bool   `json:"cross_model_cache_assumed"`
	CacheScope                      string `json:"cache_scope"`
	MaxTaskContractBytes            int    `json:"max_task_contract_bytes"`
	MaxSharedContextBytes           int    `json:"max_shared_context_bytes"`
	MaxTotalReplayBytes             int    `json:"max_total_replay_bytes"`
	MinContextCoverageBasisPoints   int    `json:"min_context_coverage_basis_points"`
	RequireCodeExcerpt              bool   `json:"require_code_excerpt"`
	RequireContentAnchor            bool   `json:"require_content_anchor"`
	RequireSelfContainedHandoff     bool   `json:"require_self_contained_handoff"`
	FallbackOnlyOnAvailabilityError bool   `json:"fallback_only_on_availability_error"`
	OnGateFailure                   string `json:"on_gate_failure"`
}

type TaskExecutionPolicy struct {
	TaskShape                string `json:"task_shape"`
	ModelSelectionTiming     string `json:"model_selection_timing"`
	SessionAffinity          string `json:"session_affinity"`
	EscalationPolicy         string `json:"escalation_policy"`
	MaxModelSwitches         int    `json:"max_model_switches"`
	DowngradeAfterGeneration bool   `json:"downgrade_after_generation"`
	PreflightBeforeExecution bool   `json:"preflight_before_execution"`
}

type SearchFirstPolicy struct {
	Required                 bool     `json:"required"`
	Reason                   string   `json:"reason,omitempty"`
	SourceOrder              []string `json:"source_order,omitempty"`
	MaxSources               int      `json:"max_sources,omitempty"`
	TimeBudgetSeconds        int      `json:"time_budget_seconds,omitempty"`
	StopConditions           []string `json:"stop_conditions,omitempty"`
	CancelStaleExecutionPlan bool     `json:"cancel_stale_execution_plan"`
}

// ChangeCapsuleGate makes a bounded high-risk change contract an explicit
// routing requirement instead of a dormant CLI command.
type ChangeCapsuleGate struct {
	Required bool   `json:"required"`
	Strict   bool   `json:"strict"`
	Reason   string `json:"reason,omitempty"`
}

type DelegationDecision struct {
	// Allowed means the selected independent worker fan-out passed its
	// delegation gates. A bounded task-judgment fallback is reported
	// separately so route JSON cannot be mistaken for implementation approval.
	Allowed               bool   `json:"allowed"`
	FallbackAllowed       bool   `json:"fallback_allowed"`
	ImplementationAllowed bool   `json:"implementation_allowed"`
	Reason                string `json:"reason"`
	ContextHandle         string `json:"context_handle,omitempty"`
	TaskContractBytes     int    `json:"task_contract_bytes"`
	TotalContractBytes    int    `json:"total_task_contract_bytes"`
	SelectedContextBytes  int    `json:"selected_context_bytes"`
	EstimatedReplayBytes  int    `json:"estimated_replay_bytes"`
	ContextCoverageBPS    int    `json:"context_coverage_basis_points"`
	CodeExcerptCount      int    `json:"code_excerpt_count"`
	ContentAnchorCount    int    `json:"content_anchor_count"`
	SelfContained         bool   `json:"self_contained"`
}

type ModelPolicy struct {
	RoutingMode        string              `json:"routing_mode"`
	UserSelectedModel  string              `json:"user_selected_model,omitempty"`
	MainModel          string              `json:"main_model"`
	MainFallbackModels []string            `json:"main_fallback_models,omitempty"`
	GeneralStaffModel  string              `json:"general_staff_model,omitempty"`
	ClassModels        map[string]string   `json:"class_models"`
	FallbackModels     map[string][]string `json:"fallback_models"`
	Delegation         string              `json:"delegation"`
}

// AjiTaskIntent is the deterministic, user-facing interpretation of a task.
// It carries PonyTail's minimum-correct judgment across every capability, not
// just code work. It is a routing contract, not a model-generated completion.
type AjiTaskIntent struct {
	Objective               string   `json:"objective"`
	Constraints             []string `json:"constraints,omitempty"`
	AcceptanceCriteria      []string `json:"acceptance_criteria,omitempty"`
	Complexity              string   `json:"complexity"`
	MinimumCorrectPath      string   `json:"minimum_correct_path"`
	ReuseCandidates         []string `json:"reuse_candidates,omitempty"`
	SelectedCapabilities    []string `json:"selected_capabilities,omitempty"`
	RejectedComplexity      []string `json:"rejected_complexity,omitempty"`
	Risks                   []string `json:"risks,omitempty"`
	EvidenceRequirements    []string `json:"evidence_requirements,omitempty"`
	SideEffects             []string `json:"side_effects,omitempty"`
	IndependentVerification bool     `json:"independent_verification"`
}

type RouteResult struct {
	Version                 string                    `json:"version"`
	Brain                   string                    `json:"brain"`
	MainModel               string                    `json:"main_model"`
	GeneralStaffModel       string                    `json:"general_staff_model,omitempty"`
	ModelPolicy             ModelPolicy               `json:"model_policy"`
	TaskIntent              AjiTaskIntent             `json:"task_intent"`
	DelegationPolicy        DelegationPolicy          `json:"delegation_policy"`
	DelegationDecision      DelegationDecision        `json:"delegation_decision"`
	TaskExecutionPolicy     TaskExecutionPolicy       `json:"task_execution_policy"`
	SearchFirstPolicy       SearchFirstPolicy         `json:"search_first_policy"`
	ChangeCapsule           ChangeCapsuleGate         `json:"change_capsule"`
	Reasoning               string                    `json:"reasoning"`
	WriteAuthority          string                    `json:"write_authority"`
	Nuwa                    bool                      `json:"nuwa"`
	Capability              string                    `json:"capability"`
	CapabilityStatus        string                    `json:"capability_status"`
	PrimarySkill            string                    `json:"primary_skill"`
	Fallback                string                    `json:"fallback,omitempty"`
	Engine                  string                    `json:"engine,omitempty"`
	Provider                string                    `json:"provider,omitempty"`
	ProviderFallback        string                    `json:"provider_fallback,omitempty"`
	SecondaryCapabilities   []string                  `json:"secondary_capabilities,omitempty"`
	MountedSources          []MountedSource           `json:"mounted_sources"`
	SourceExecution         []SourceExecutionContract `json:"source_execution,omitempty"`
	ResponsePolicy          *ResponsePolicyContract   `json:"response_policy,omitempty"`
	ResponsePolicyError     string                    `json:"response_policy_error,omitempty"`
	AssetContracts          []AssetInvocationContract `json:"asset_contracts,omitempty"`
	SourceActivationError   string                    `json:"source_activation_error,omitempty"`
	ExecutionLane           string                    `json:"execution_lane"`
	GeneralStaffWorker      *WorkerTask               `json:"general_staff_worker,omitempty"`
	Parallel                bool                      `json:"parallel"`
	PreflightWorkers        []WorkerTask              `json:"preflight_workers,omitempty"`
	Workers                 []WorkerTask              `json:"workers,omitempty"`
	Officers                []string                  `json:"officers,omitempty"`
	OfficerRecommendations  []OfficerRecommendation   `json:"officer_recommendations,omitempty"`
	OfficerWorkers          []WorkerTask              `json:"officer_workers,omitempty"`
	InternalAdversarialPass bool                      `json:"internal_adversarial_pass"`
	FinishLine              []string                  `json:"finish_line"`
}

type VerifyResult struct {
	Capability string                    `json:"capability"`
	Claimed    string                    `json:"claimed_status"`
	Effective  string                    `json:"effective_status"`
	Passed     bool                      `json:"passed"`
	Sources    []string                  `json:"sources"`
	Assets     []AssetReachability       `json:"assets,omitempty"`
	Genome     *FusionGenomeVerification `json:"fusion_genome,omitempty"`
	Checks     []string                  `json:"checks"`
	Errors     []string                  `json:"errors,omitempty"`
	Probe      *ProbeEvidence            `json:"probe_evidence,omitempty"`
}

type ContextExcerpt struct {
	Path          string   `json:"path"`
	Score         int      `json:"score"`
	LineRanges    []string `json:"line_ranges"`
	Text          string   `json:"text"`
	SourceSHA256  string   `json:"source_sha256"`
	ContentSHA256 string   `json:"content_sha256"`
}

type ContextResult struct {
	Workspace          string           `json:"workspace"`
	Query              string           `json:"query"`
	QueryFingerprint   string           `json:"query_fingerprint"`
	BudgetBytes        int              `json:"budget_bytes"`
	SelectedBytes      int              `json:"selected_bytes"`
	ScannedFiles       int              `json:"scanned_files"`
	IndexedFiles       int              `json:"indexed_files"`
	CandidateFiles     int              `json:"candidate_files"`
	GraphLookups       int              `json:"graph_lookups"`
	RetrievalTruncated bool             `json:"retrieval_truncated"`
	GraphSourceBytes   int64            `json:"graph_source_bytes"`
	RetrievalMode      string           `json:"retrieval_mode"`
	FallbackReason     string           `json:"fallback_reason,omitempty"`
	ContextHandle      string           `json:"context_handle"`
	ContentSHA256      string           `json:"content_sha256"`
	ArtifactPath       string           `json:"artifact_path,omitempty"`
	RetrievalTerms     []string         `json:"retrieval_terms"`
	MatchedTerms       []string         `json:"matched_terms"`
	CoverageBPS        int              `json:"coverage_basis_points"`
	CodeExcerptCount   int              `json:"code_excerpt_count"`
	ContentAnchorCount int              `json:"content_anchor_count"`
	PayloadSHA256      string           `json:"payload_sha256"`
	PayloadBytes       int              `json:"payload_bytes"`
	Excerpts           []ContextExcerpt `json:"excerpts"`
	Policy             []string         `json:"policy"`
}

type ContextArtifact struct {
	SchemaVersion      int              `json:"schema_version"`
	Workspace          string           `json:"workspace"`
	QueryFingerprint   string           `json:"query_fingerprint"`
	Handle             string           `json:"handle"`
	ContentSHA256      string           `json:"content_sha256"`
	SelectedBytes      int              `json:"selected_bytes"`
	RetrievalTerms     []string         `json:"retrieval_terms"`
	MatchedTerms       []string         `json:"matched_terms"`
	CoverageBPS        int              `json:"coverage_basis_points"`
	CodeExcerptCount   int              `json:"code_excerpt_count"`
	ContentAnchorCount int              `json:"content_anchor_count"`
	PayloadSHA256      string           `json:"payload_sha256"`
	PayloadBytes       int              `json:"payload_bytes"`
	Excerpts           []ContextExcerpt `json:"excerpts"`
}

type DelegationContext struct {
	Handle                string
	ArtifactPath          string
	QueryFingerprint      string
	SelectedBytes         int
	RetrievalTerms        []string
	MatchedTerms          []string
	CoverageBPS           int
	CodeExcerptCount      int
	ContentAnchorCount    int
	Payload               string
	PayloadSHA256         string
	ParentContextRequired bool
	SelfContained         bool
	verified              bool
}

type EvolutionResult struct {
	Candidate       string        `json:"candidate"`
	Decision        string        `json:"decision"`
	Applied         bool          `json:"applied"`
	ExistingStatus  string        `json:"existing_status,omitempty"`
	ExistingProof   *VerifyResult `json:"existing_proof,omitempty"`
	CandidateProof  VerifyResult  `json:"candidate_proof"`
	Comparison      string        `json:"comparison,omitempty"`
	RequiredActions []string      `json:"required_actions,omitempty"`
}
