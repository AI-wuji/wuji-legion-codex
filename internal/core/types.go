package core

type Source struct {
	ID       string   `json:"id"`
	Engine   string   `json:"engine,omitempty"`
	Priority string   `json:"priority,omitempty"`
	Globs    []string `json:"globs"`
	Required []string `json:"required"`
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
	ID       string `json:"id"`
	Path     string `json:"path"`
	Priority string `json:"priority,omitempty"`
}

type WorkerTask struct {
	ID                         string   `json:"id"`
	Purpose                    string   `json:"purpose"`
	ModelClass                 string   `json:"model_class"`
	Model                      string   `json:"model"`
	FallbackModels             []string `json:"fallback_models,omitempty"`
	Inputs                     []string `json:"inputs"`
	ContextMode                string   `json:"context_mode"`
	ContextHandles             []string `json:"context_handles,omitempty"`
	ContextArtifact            string   `json:"context_artifact,omitempty"`
	AllocatedContextBytes      int      `json:"allocated_context_bytes"`
	AllocatedTaskContractBytes int      `json:"allocated_task_contract_bytes"`
	MaxTaskContractBytes       int      `json:"max_task_contract_bytes"`
	DelegationGateReason       string   `json:"delegation_gate_reason"`
	Writes                     bool     `json:"writes"`
	ExecutionEvidenceRequired  bool     `json:"execution_evidence_required"`
	ExecutionEvidenceFields    []string `json:"execution_evidence_fields"`
}

type DelegationPolicy struct {
	CrossModelCacheAssumed bool   `json:"cross_model_cache_assumed"`
	MaxTaskContractBytes   int    `json:"max_task_contract_bytes"`
	MaxSharedContextBytes  int    `json:"max_shared_context_bytes"`
	MaxTotalReplayBytes    int    `json:"max_total_replay_bytes"`
	OnGateFailure          string `json:"on_gate_failure"`
}

type DelegationDecision struct {
	Allowed              bool   `json:"allowed"`
	Reason               string `json:"reason"`
	ContextHandle        string `json:"context_handle,omitempty"`
	TaskContractBytes    int    `json:"task_contract_bytes"`
	SelectedContextBytes int    `json:"selected_context_bytes"`
	EstimatedReplayBytes int    `json:"estimated_replay_bytes"`
}

type ModelPolicy struct {
	MainModel      string              `json:"main_model"`
	ClassModels    map[string]string   `json:"class_models"`
	FallbackModels map[string][]string `json:"fallback_models"`
	Delegation     string              `json:"delegation"`
}

type RouteResult struct {
	Version                 string             `json:"version"`
	Brain                   string             `json:"brain"`
	MainModel               string             `json:"main_model"`
	ModelPolicy             ModelPolicy        `json:"model_policy"`
	DelegationPolicy        DelegationPolicy   `json:"delegation_policy"`
	DelegationDecision      DelegationDecision `json:"delegation_decision"`
	Reasoning               string             `json:"reasoning"`
	WriteAuthority          string             `json:"write_authority"`
	Nuwa                    bool               `json:"nuwa"`
	Capability              string             `json:"capability"`
	CapabilityStatus        string             `json:"capability_status"`
	PrimarySkill            string             `json:"primary_skill"`
	Fallback                string             `json:"fallback,omitempty"`
	Engine                  string             `json:"engine,omitempty"`
	Provider                string             `json:"provider,omitempty"`
	ProviderFallback        string             `json:"provider_fallback,omitempty"`
	SecondaryCapabilities   []string           `json:"secondary_capabilities,omitempty"`
	MountedSources          []MountedSource    `json:"mounted_sources"`
	ExecutionLane           string             `json:"execution_lane"`
	Parallel                bool               `json:"parallel"`
	Workers                 []WorkerTask       `json:"workers,omitempty"`
	Officers                []string           `json:"officers,omitempty"`
	InternalAdversarialPass bool               `json:"internal_adversarial_pass"`
	FinishLine              []string           `json:"finish_line"`
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
	Workspace        string           `json:"workspace"`
	Query            string           `json:"query"`
	QueryFingerprint string           `json:"query_fingerprint"`
	BudgetBytes      int              `json:"budget_bytes"`
	SelectedBytes    int              `json:"selected_bytes"`
	ScannedFiles     int              `json:"scanned_files"`
	ContextHandle    string           `json:"context_handle"`
	ContentSHA256    string           `json:"content_sha256"`
	ArtifactPath     string           `json:"artifact_path,omitempty"`
	Excerpts         []ContextExcerpt `json:"excerpts"`
	Policy           []string         `json:"policy"`
}

type ContextArtifact struct {
	SchemaVersion    int              `json:"schema_version"`
	Workspace        string           `json:"workspace"`
	QueryFingerprint string           `json:"query_fingerprint"`
	Handle           string           `json:"handle"`
	ContentSHA256    string           `json:"content_sha256"`
	SelectedBytes    int              `json:"selected_bytes"`
	Excerpts         []ContextExcerpt `json:"excerpts"`
}

type DelegationContext struct {
	Handle                string
	ArtifactPath          string
	QueryFingerprint      string
	SelectedBytes         int
	ParentContextRequired bool
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
