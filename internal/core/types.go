package core

type Source struct {
	ID         string   `json:"id"`
	Engine     string   `json:"engine,omitempty"`
	Priority   string   `json:"priority,omitempty"`
	Lifecycle  string   `json:"lifecycle,omitempty"`
	Activation []string `json:"activation,omitempty"`
	Entrypoint string   `json:"entrypoint,omitempty"`
	Fallback   string   `json:"fallback,omitempty"`
	Globs      []string `json:"globs"`
	Required   []string `json:"required"`
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
	Root             string     `json:"-"`
	ID               string     `json:"id"`
	Description      string     `json:"description"`
	Triggers         []string   `json:"triggers"`
	Status           string     `json:"status"`
	PromotionReceipt string     `json:"promotion_receipt,omitempty"`
	PrimarySkill     string     `json:"primary_skill"`
	HostCallable     bool       `json:"host_callable"`
	DirectMount      bool       `json:"direct_mount"`
	Fallback         string     `json:"fallback"`
	Sources          []Source   `json:"sources"`
	Probe            *Probe     `json:"probe,omitempty"`
	Experts          []Expert   `json:"experts,omitempty"`
	Providers        []Provider `json:"providers,omitempty"`
	Engines          []Engine   `json:"engines,omitempty"`
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
	Capability        string `json:"capability"`
	Source            string `json:"source"`
	Lifecycle         string `json:"lifecycle"`
	State             string `json:"state"`
	ExecutionMode     string `json:"execution_mode"`
	ExecutionEvidence string `json:"execution_evidence"`
	Entrypoint        string `json:"entrypoint,omitempty"`
	Reason            string `json:"reason"`
}

type WorkerTask struct {
	ID                         string                    `json:"id"`
	Stage                      string                    `json:"stage"`
	Purpose                    string                    `json:"purpose"`
	ModelClass                 string                    `json:"model_class"`
	Model                      string                    `json:"model"`
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
	SchemaVersion            int             `json:"schema_version"`
	WorkerID                 string          `json:"worker_id"`
	RequestedModel           string          `json:"requested_model"`
	SessionKey               string          `json:"session_key"`
	HostDispatchID           string          `json:"host_dispatch_id"`
	WriteBoundary            string          `json:"write_boundary"`
	Attempts                 []WorkerAttempt `json:"attempts"`
	EffectiveModel           string          `json:"effective_model"`
	ModelSwitchCount         int             `json:"model_switch_count"`
	ResultHandle             string          `json:"result_handle"`
	ContextHandleIDs         []string        `json:"context_handle_ids"`
	StablePrefixBytesSent    int             `json:"stable_prefix_bytes"`
	StablePrefixSHA256       string          `json:"stable_prefix_sha256"`
	SourceExecutionBytesSent int             `json:"source_execution_bytes"`
	ContextBytesSent         int             `json:"context_bytes_sent"`
	ContextPayloadSHA256     string          `json:"context_payload_sha256"`
	TaskContractBytes        int             `json:"task_contract_bytes"`
	TaskContractSHA256       string          `json:"task_contract_sha256"`
	InputTokens              int             `json:"input_tokens"`
	CachedInputTokens        int             `json:"cached_input_tokens"`
	OutputTokens             int             `json:"output_tokens"`
	RetryCount               int             `json:"retry_count"`
	AcceptedByAji            bool            `json:"accepted_by_aji"`
	AttemptFailureKinds      []string        `json:"attempt_failure_kinds"`
	CacheDomain              string          `json:"cache_domain"`
	DelegationGateReason     string          `json:"delegation_gate_reason"`
	BillingUnit              string          `json:"billing_unit"`
	TotalCostMicrounits      int64           `json:"total_cost_microunits"`
	AjiBaselineMicrounits    int64           `json:"aji_baseline_microunits"`
	SavingsMicrounits        int64           `json:"savings_microunits"`
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
	Allowed              bool   `json:"allowed"`
	Reason               string `json:"reason"`
	ContextHandle        string `json:"context_handle,omitempty"`
	TaskContractBytes    int    `json:"task_contract_bytes"`
	TotalContractBytes   int    `json:"total_task_contract_bytes"`
	SelectedContextBytes int    `json:"selected_context_bytes"`
	EstimatedReplayBytes int    `json:"estimated_replay_bytes"`
	ContextCoverageBPS   int    `json:"context_coverage_basis_points"`
	CodeExcerptCount     int    `json:"code_excerpt_count"`
	ContentAnchorCount   int    `json:"content_anchor_count"`
	SelfContained        bool   `json:"self_contained"`
}

type ModelPolicy struct {
	MainModel      string              `json:"main_model"`
	ClassModels    map[string]string   `json:"class_models"`
	FallbackModels map[string][]string `json:"fallback_models"`
	Delegation     string              `json:"delegation"`
}

type RouteResult struct {
	Version                 string                    `json:"version"`
	Brain                   string                    `json:"brain"`
	MainModel               string                    `json:"main_model"`
	ModelPolicy             ModelPolicy               `json:"model_policy"`
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
	SourceActivationError   string                    `json:"source_activation_error,omitempty"`
	ExecutionLane           string                    `json:"execution_lane"`
	Parallel                bool                      `json:"parallel"`
	PreflightWorkers        []WorkerTask              `json:"preflight_workers,omitempty"`
	Workers                 []WorkerTask              `json:"workers,omitempty"`
	Officers                []string                  `json:"officers,omitempty"`
	OfficerWorkers          []WorkerTask              `json:"officer_workers,omitempty"`
	InternalAdversarialPass bool                      `json:"internal_adversarial_pass"`
	FinishLine              []string                  `json:"finish_line"`
}

type VerifyResult struct {
	Capability string         `json:"capability"`
	Claimed    string         `json:"claimed_status"`
	Effective  string         `json:"effective_status"`
	Passed     bool           `json:"passed"`
	Sources    []string       `json:"sources"`
	Checks     []string       `json:"checks"`
	Errors     []string       `json:"errors,omitempty"`
	Probe      *ProbeEvidence `json:"probe_evidence,omitempty"`
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
