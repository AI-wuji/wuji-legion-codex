package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	responsePolicyCapabilityID = "interaction"
	responsePolicyDomain       = "response-policy"
	maxResponseRuleBytes       = 64 * 1024
)

type responseRuleSet struct {
	SchemaVersion int                 `json:"schema_version"`
	ID            string              `json:"id"`
	Revision      string              `json:"revision"`
	Mode          string              `json:"mode"`
	Source        responseRuleSource  `json:"source"`
	Activation    []string            `json:"activation"`
	Exit          []string            `json:"exit"`
	Precedence    []string            `json:"precedence"`
	Atoms         []ResponseDirective `json:"atoms"`
}

type responseRuleSource struct {
	Repository string `json:"repository"`
	Commit     string `json:"commit"`
	License    string `json:"license"`
}

// CompileResponsePolicy loads the trusted rule asset and compiles a compact
// final-writer contract. active carries session state from the host; explicit
// exit language always wins over both carried state and activation language.
func CompileResponsePolicy(manifest Manifest, query string, active bool) (*ResponsePolicyContract, error) {
	if manifest.ID != responsePolicyCapabilityID {
		return nil, fmt.Errorf("response policy requires capability %q", responsePolicyCapabilityID)
	}
	if manifest.Genome == nil {
		return nil, fmt.Errorf("interaction capability has no fusion genome")
	}
	rules, raw, err := loadResponseRuleSet(manifest)
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(strings.TrimSpace(query))
	exit := matchesResponseTrigger(lower, rules.Exit)
	activate := matchesResponseTrigger(lower, rules.Activation)
	if !active && !activate && !exit {
		return nil, nil
	}
	reason := "session-active"
	if exit {
		active = false
		reason = "explicit-exit"
	} else if activate {
		active = true
		reason = "explicit-activation"
	}
	digest := sha256.Sum256(raw)
	contract := &ResponsePolicyContract{
		ID: rules.ID, Revision: rules.Revision, Active: active, ActivationReason: reason,
		SourceCommit: rules.Source.Commit, RulesSHA256: hex.EncodeToString(digest[:]), RulesBytes: len(raw),
		Precedence: append([]string(nil), rules.Precedence...), ExitTriggers: append([]string(nil), rules.Exit...),
	}
	if !active {
		return contract, nil
	}

	contexts := responseRuleContexts(lower)
	suppressed := responseRuleSuppressions(contexts, rules.Atoms)
	for _, atom := range rules.Atoms {
		if suppressed[atom.ID] {
			contract.Suppressed = append(contract.Suppressed, atom.ID)
			continue
		}
		if !responseDirectiveApplies(atom, contexts) {
			continue
		}
		contract.Directives = append(contract.Directives, atom)
	}
	sort.SliceStable(contract.Directives, func(i, j int) bool {
		return contract.Directives[i].Priority > contract.Directives[j].Priority
	})
	sort.Strings(contract.Suppressed)
	return contract, nil
}

func responsePolicyRequested(query string) bool {
	return containsAny(strings.ToLower(query),
		"开启专注执行模式", "开启专注模式", "启用行动专注模式", "启用专注执行模式", "enable action focus", "start action focus", "start focused execution", "use i-have-adhd", "使用 i-have-adhd",
		"停止专注模式", "退出专注模式", "恢复正常模式", "stop focus mode", "disable action focus", "return to normal mode", "switch to normal mode")
}

func capabilityManifest(manifests []Manifest, id string) (Manifest, bool) {
	for _, manifest := range manifests {
		if manifest.ID == id {
			return manifest, true
		}
	}
	return Manifest{}, false
}

func loadResponseRuleSet(manifest Manifest) (responseRuleSet, []byte, error) {
	var adapter *FusionAdapter
	var assetPath string
	for index := range manifest.Genome.Adapters {
		candidate := &manifest.Genome.Adapters[index]
		if candidate.Domain != responsePolicyDomain {
			continue
		}
		assets, err := fusionAdapterAssets(manifest.ID, *candidate)
		if err != nil {
			return responseRuleSet{}, nil, err
		}
		for _, asset := range assets {
			if asset.ID == "response-rules" {
				adapter, assetPath = candidate, asset.Path
				break
			}
		}
	}
	if adapter == nil || assetPath == "" {
		return responseRuleSet{}, nil, fmt.Errorf("interaction fusion genome has no response-rules asset")
	}
	var source *Source
	for index := range manifest.Sources {
		if manifest.Sources[index].ID == adapter.Source {
			source = &manifest.Sources[index]
			break
		}
	}
	if source == nil {
		return responseRuleSet{}, nil, fmt.Errorf("response policy source %q is not registered", adapter.Source)
	}
	root, ok := ResolveCompleteSourceAt(manifest.Root, *source)
	if !ok {
		return responseRuleSet{}, nil, fmt.Errorf("response policy source %q is unavailable", source.ID)
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return responseRuleSet{}, nil, fmt.Errorf("resolve response policy root: %w", err)
	}
	resolvedAsset, err := filepath.EvalSymlinks(filepath.Join(resolvedRoot, filepath.FromSlash(assetPath)))
	if err != nil {
		return responseRuleSet{}, nil, fmt.Errorf("resolve response policy asset: %w", err)
	}
	relative, err := filepath.Rel(resolvedRoot, resolvedAsset)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return responseRuleSet{}, nil, fmt.Errorf("response policy asset escapes its trusted source")
	}
	raw, err := os.ReadFile(resolvedAsset)
	if err != nil {
		return responseRuleSet{}, nil, fmt.Errorf("read response policy asset: %w", err)
	}
	if len(raw) == 0 || len(raw) > maxResponseRuleBytes {
		return responseRuleSet{}, nil, fmt.Errorf("response policy asset must be between 1 and %d bytes", maxResponseRuleBytes)
	}
	var rules responseRuleSet
	if err := json.Unmarshal(raw, &rules); err != nil {
		return responseRuleSet{}, nil, fmt.Errorf("decode response policy asset: %w", err)
	}
	if err := validateResponseRuleSet(rules); err != nil {
		return responseRuleSet{}, nil, err
	}
	return rules, raw, nil
}

func validateResponseRuleSet(rules responseRuleSet) error {
	if rules.SchemaVersion != 1 || !componentIDPattern.MatchString(rules.ID) || !componentIDPattern.MatchString(rules.Revision) {
		return fmt.Errorf("response rule identity is invalid")
	}
	if rules.Mode != "explicit-session" {
		return fmt.Errorf("response rule mode must be explicit-session")
	}
	if len(rules.Source.Commit) != 40 || rules.Source.License == "" || rules.Source.Repository == "" {
		return fmt.Errorf("response rule provenance is incomplete")
	}
	if _, err := hex.DecodeString(rules.Source.Commit); err != nil {
		return fmt.Errorf("response rule source commit is invalid")
	}
	if err := validateStringList("response activation", rules.Activation, true); err != nil {
		return err
	}
	if err := validateStringList("response exit", rules.Exit, true); err != nil {
		return err
	}
	wantPrecedence := []string{"host-safety", "explicit-user-instruction", "task-specific-contract", "action-focus-default"}
	if strings.Join(rules.Precedence, "\x00") != strings.Join(wantPrecedence, "\x00") {
		return fmt.Errorf("response rule precedence must preserve host, user, task, then default order")
	}
	if len(rules.Atoms) == 0 || len(rules.Atoms) > 32 {
		return fmt.Errorf("response rules require between 1 and 32 atoms")
	}
	ids := map[string]bool{}
	for _, atom := range rules.Atoms {
		if err := validateComponentID("response rule atom", atom.ID, ids); err != nil {
			return err
		}
		if atom.Priority < 1 || atom.Priority > 1000 || strings.TrimSpace(atom.Phase) == "" || strings.TrimSpace(atom.Directive) == "" || len(atom.Conditions) == 0 {
			return fmt.Errorf("response rule atom %q is incomplete", atom.ID)
		}
	}
	for _, atom := range rules.Atoms {
		for _, overridden := range atom.Overrides {
			if !ids[overridden] || overridden == atom.ID {
				return fmt.Errorf("response rule atom %q has invalid override %q", atom.ID, overridden)
			}
		}
	}
	return nil
}

func matchesResponseTrigger(query string, triggers []string) bool {
	for _, trigger := range triggers {
		trigger = strings.ToLower(trigger)
		for offset := 0; offset < len(query); {
			index := strings.Index(query[offset:], trigger)
			if index < 0 {
				break
			}
			index += offset
			if !responseTriggerNegated(query, index) {
				return true
			}
			offset = index + len(trigger)
		}
	}
	return false
}

func responseTriggerNegated(query string, index int) bool {
	start := index - 32
	if start < 0 {
		start = 0
	}
	prefix := strings.TrimSpace(query[start:index])
	for _, negation := range []string{"do not", "don't", "dont", "not", "不要", "不许", "别", "勿"} {
		if strings.HasSuffix(prefix, negation) {
			return true
		}
	}
	return false
}

func responseRuleContexts(query string) map[string]bool {
	return map[string]bool{
		"always":                   true,
		"explanation-request":      containsAny(query, "解释", "为什么", "原理", "explain", "why", "what is"),
		"destructive-or-ambiguous": containsAny(query, "删除", "清空", "覆盖", "重置", "delete", "remove", "overwrite", "reset", "ambiguous", "不明确"),
		"action-request":           containsAny(query, "修复", "创建", "生成", "实现", "修改", "更新", "完成", "处理", "fix", "create", "build", "implement", "write", "update", "change", "complete"),
		"multi-step":               containsAny(query, "多步骤", "多个步骤", "分步骤", "multi-step", "multiple steps"),
		"continuation-or-status":   containsAny(query, "继续", "进度", "状态", "接着", "continue", "status", "progress", "resume"),
		"error":                    containsAny(query, "错误", "报错", "失败", "故障", "bug", "error", "failed", "failure"),
	}
}

func responseDirectiveApplies(atom ResponseDirective, contexts map[string]bool) bool {
	for _, condition := range atom.Conditions {
		if contexts[condition] {
			return true
		}
	}
	return false
}

func responseRuleSuppressions(contexts map[string]bool, atoms []ResponseDirective) map[string]bool {
	suppressed := map[string]bool{}
	for _, atom := range atoms {
		applies := false
		for _, condition := range atom.Conditions {
			if contexts[condition] {
				applies = true
				break
			}
		}
		if applies {
			for _, id := range atom.Overrides {
				suppressed[id] = true
			}
		}
	}
	return suppressed
}

// ResponseDraft is a deterministic behavior fixture used by the independent
// probe. It keeps verification about observable response properties rather
// than accepting a model's self-reported compliance.
type ResponseDraft struct {
	FirstLineAction       bool
	StepCount             int
	ClosingActionCount    int
	ClosingActionMinutes  int
	Continuation          bool
	RestatesCurrentState  bool
	Error                 bool
	ErrorCausePresent     bool
	ErrorFixPresent       bool
	ContainsOffTopic      bool
	Destructive           bool
	ConfirmationRequested bool
}

func ValidateResponseDraft(contract *ResponsePolicyContract, draft ResponseDraft) []string {
	if contract == nil || !contract.Active {
		return nil
	}
	directives := map[string]ResponseDirective{}
	for _, directive := range contract.Directives {
		directives[directive.ID] = directive
	}
	violations := []string{}
	if _, ok := directives["first-action"]; ok && !draft.FirstLineAction {
		violations = append(violations, "first-action")
	}
	if directive, ok := directives["bounded-steps"]; ok {
		maximum := 5
		if value, ok := directive.Parameters["max_items"].(float64); ok {
			maximum = int(value)
		}
		if draft.StepCount > maximum {
			violations = append(violations, "bounded-steps")
		}
	}
	if _, ok := directives["single-next-action"]; ok && (draft.ClosingActionCount != 1 || draft.ClosingActionMinutes > 2) {
		violations = append(violations, "single-next-action")
	}
	if _, ok := directives["progress-state"]; ok && draft.Continuation && !draft.RestatesCurrentState {
		violations = append(violations, "progress-state")
	}
	if _, ok := directives["cause-fix"]; ok && draft.Error && (!draft.ErrorCausePresent || !draft.ErrorFixPresent) {
		violations = append(violations, "cause-fix")
	}
	if _, ok := directives["topic-filter"]; ok && draft.ContainsOffTopic {
		violations = append(violations, "topic-filter")
	}
	if _, ok := directives["destructive-clarity"]; ok && draft.Destructive && !draft.ConfirmationRequested {
		violations = append(violations, "destructive-clarity")
	}
	sort.Strings(violations)
	return violations
}
