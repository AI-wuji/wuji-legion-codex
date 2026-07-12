package core

type Source struct {
	ID       string   `json:"id"`
	Engine   string   `json:"engine,omitempty"`
	Priority string   `json:"priority,omitempty"`
	Globs    []string `json:"globs"`
	Required []string `json:"required"`
}

type Probe struct {
	Command        string   `json:"command"`
	Args           []string `json:"args"`
	Fixture        string   `json:"fixture,omitempty"`
	Kind           string   `json:"kind,omitempty"`
	TimeoutSeconds int      `json:"timeout_seconds,omitempty"`
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
	Root         string     `json:"-"`
	ID           string     `json:"id"`
	Description  string     `json:"description"`
	Triggers     []string   `json:"triggers"`
	Status       string     `json:"status"`
	PrimarySkill string     `json:"primary_skill"`
	HostCallable bool       `json:"host_callable"`
	DirectMount  bool       `json:"direct_mount"`
	Fallback     string     `json:"fallback"`
	Sources      []Source   `json:"sources"`
	Probe        *Probe     `json:"probe,omitempty"`
	Experts      []Expert   `json:"experts,omitempty"`
	Providers    []Provider `json:"providers,omitempty"`
	Engines      []Engine   `json:"engines,omitempty"`
}

type MountedSource struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Priority string `json:"priority,omitempty"`
}

type WorkerTask struct {
	ID             string   `json:"id"`
	Purpose        string   `json:"purpose"`
	ModelClass     string   `json:"model_class"`
	Model          string   `json:"model"`
	FallbackModels []string `json:"fallback_models,omitempty"`
	Inputs         []string `json:"inputs"`
	Writes         bool     `json:"writes"`
}

type ModelPolicy struct {
	MainModel      string              `json:"main_model"`
	ClassModels    map[string]string   `json:"class_models"`
	FallbackModels map[string][]string `json:"fallback_models"`
	Delegation     string              `json:"delegation"`
}

type RouteResult struct {
	Version                 string          `json:"version"`
	Brain                   string          `json:"brain"`
	MainModel               string          `json:"main_model"`
	ModelPolicy             ModelPolicy     `json:"model_policy"`
	Reasoning               string          `json:"reasoning"`
	WriteAuthority          string          `json:"write_authority"`
	Nuwa                    bool            `json:"nuwa"`
	Capability              string          `json:"capability"`
	CapabilityStatus        string          `json:"capability_status"`
	PrimarySkill            string          `json:"primary_skill"`
	Fallback                string          `json:"fallback,omitempty"`
	Engine                  string          `json:"engine,omitempty"`
	Provider                string          `json:"provider,omitempty"`
	ProviderFallback        string          `json:"provider_fallback,omitempty"`
	SecondaryCapabilities   []string        `json:"secondary_capabilities,omitempty"`
	MountedSources          []MountedSource `json:"mounted_sources"`
	ExecutionLane           string          `json:"execution_lane"`
	Parallel                bool            `json:"parallel"`
	Workers                 []WorkerTask    `json:"workers,omitempty"`
	Officers                []string        `json:"officers,omitempty"`
	InternalAdversarialPass bool            `json:"internal_adversarial_pass"`
	FinishLine              []string        `json:"finish_line"`
}

type VerifyResult struct {
	Capability string   `json:"capability"`
	Claimed    string   `json:"claimed_status"`
	Effective  string   `json:"effective_status"`
	Passed     bool     `json:"passed"`
	Sources    []string `json:"sources"`
	Checks     []string `json:"checks"`
	Errors     []string `json:"errors,omitempty"`
}

type ContextExcerpt struct {
	Path       string   `json:"path"`
	Score      int      `json:"score"`
	LineRanges []string `json:"line_ranges"`
	Text       string   `json:"text"`
}

type ContextResult struct {
	Workspace     string           `json:"workspace"`
	Query         string           `json:"query"`
	BudgetBytes   int              `json:"budget_bytes"`
	SelectedBytes int              `json:"selected_bytes"`
	ScannedFiles  int              `json:"scanned_files"`
	Excerpts      []ContextExcerpt `json:"excerpts"`
	Policy        []string         `json:"policy"`
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
