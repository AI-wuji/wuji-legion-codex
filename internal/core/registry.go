package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var capabilityIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var componentIDPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9._-]*[a-z0-9])?$`)
var naturalPartPattern = regexp.MustCompile(`\d+|\D+`)

var lifecycleStatuses = map[string]bool{
	"known":             true,
	"doctrine-only":     true,
	"assets-retained":   true,
	"callable":          true,
	"behavior-verified": true,
	"primary":           true,
}

func ValidateManifest(item Manifest) error {
	if item.ID == "" {
		return fmt.Errorf("id is required")
	}
	if len(item.ID) > 64 || !capabilityIDPattern.MatchString(item.ID) {
		return fmt.Errorf("invalid capability id %q: use lowercase letters, numbers, and single hyphens", item.ID)
	}
	if !lifecycleStatuses[item.Status] {
		return fmt.Errorf("invalid lifecycle status %q", item.Status)
	}
	if strings.TrimSpace(item.Description) == "" {
		return fmt.Errorf("description is required")
	}
	if err := validateStringList("triggers", item.Triggers, true); err != nil {
		return err
	}
	if strings.TrimSpace(item.PrimarySkill) == "" {
		return fmt.Errorf("primary_skill is required")
	}
	if strings.TrimSpace(item.Fallback) == "" {
		return fmt.Errorf("fallback is required")
	}
	if rank(item.Status) >= rank("callable") && !item.HostCallable && !item.DirectMount {
		return fmt.Errorf("%s capability requires host_callable or direct_mount", item.Status)
	}
	if item.DirectMount && len(item.Sources) == 0 {
		return fmt.Errorf("direct_mount requires at least one source")
	}
	if item.Status == "assets-retained" && len(item.Sources) == 0 {
		return fmt.Errorf("assets-retained capability requires at least one source")
	}
	if err := validateSources(item.Sources, item.Engines); err != nil {
		return err
	}
	if err := validateProviders(item.Providers); err != nil {
		return err
	}
	if err := validateEngines(item.Engines, item.Sources); err != nil {
		return err
	}
	if err := validateExperts(item.Experts); err != nil {
		return err
	}
	if item.Probe != nil {
		if strings.TrimSpace(item.Probe.Command) == "" {
			return fmt.Errorf("probe command is required")
		}
		if item.Probe.TimeoutSeconds < 0 || item.Probe.TimeoutSeconds > 3600 {
			return fmt.Errorf("probe timeout_seconds must be between 0 and 3600")
		}
		switch strings.ToLower(strings.TrimSpace(item.Probe.Kind)) {
		case "", "behavior", "smoke", "mount":
		default:
			return fmt.Errorf("probe kind %q is invalid", item.Probe.Kind)
		}
	}
	if rank(item.Status) >= rank("behavior-verified") {
		if item.Probe == nil {
			return fmt.Errorf("%s capability requires a behavior probe", item.Status)
		}
		if strings.TrimSpace(item.Probe.Fixture) == "" {
			return fmt.Errorf("%s capability probe requires a fixture id", item.Status)
		}
		if !componentIDPattern.MatchString(item.Probe.Fixture) {
			return fmt.Errorf("probe fixture id %q is invalid", item.Probe.Fixture)
		}
		if strings.ToLower(strings.TrimSpace(item.Probe.Kind)) != "behavior" {
			return fmt.Errorf("%s capability requires probe kind behavior", item.Status)
		}
	}
	return nil
}

func validateSources(sources []Source, engines []Engine) error {
	ids := map[string]bool{}
	engineIDs := map[string]bool{}
	for _, engine := range engines {
		engineIDs[engine.ID] = true
	}
	for _, source := range sources {
		if err := validateComponentID("source", source.ID, ids); err != nil {
			return err
		}
		if source.Engine != "" && !engineIDs[source.Engine] {
			return fmt.Errorf("source %q references unknown engine %q", source.ID, source.Engine)
		}
		switch strings.ToLower(strings.TrimSpace(source.Priority)) {
		case "", "primary", "secondary", "optional":
		default:
			return fmt.Errorf("source %q has invalid priority %q", source.ID, source.Priority)
		}
		if err := validateStringList("source "+source.ID+" globs", source.Globs, true); err != nil {
			return err
		}
		for _, pattern := range source.Globs {
			if _, err := filepath.Glob(ExpandPath(pattern)); err != nil {
				return fmt.Errorf("source %q has invalid glob %q: %w", source.ID, pattern, err)
			}
		}
		if err := validateStringList("source "+source.ID+" required", source.Required, true); err != nil {
			return err
		}
		for _, required := range source.Required {
			clean := filepath.Clean(filepath.FromSlash(required))
			if filepath.IsAbs(clean) || filepath.VolumeName(clean) != "" || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
				return fmt.Errorf("source %q required path must stay relative: %q", source.ID, required)
			}
			if _, err := filepath.Glob(filepath.Join("root", clean)); err != nil {
				return fmt.Errorf("source %q has invalid required glob %q: %w", source.ID, required, err)
			}
		}
	}
	return nil
}

func validateProviders(providers []Provider) error {
	if len(providers) == 0 {
		return nil
	}
	ids := map[string]bool{}
	defaults := 0
	for _, provider := range providers {
		if err := validateComponentID("provider", provider.ID, ids); err != nil {
			return err
		}
		if provider.Default {
			defaults++
		} else if len(provider.Triggers) == 0 {
			return fmt.Errorf("non-default provider %q requires triggers", provider.ID)
		}
		if err := validateStringList("provider "+provider.ID+" triggers", provider.Triggers, false); err != nil {
			return err
		}
	}
	if defaults != 1 {
		return fmt.Errorf("exactly one default provider is required, got %d", defaults)
	}
	return nil
}

func validateEngines(engines []Engine, sources []Source) error {
	if len(engines) == 0 {
		return nil
	}
	ids := map[string]bool{}
	coverage := map[string]int{}
	defaults := 0
	for _, source := range sources {
		if source.Engine != "" {
			coverage[source.Engine]++
		}
	}
	for _, engine := range engines {
		if err := validateComponentID("engine", engine.ID, ids); err != nil {
			return err
		}
		if strings.TrimSpace(engine.PrimarySkill) == "" {
			return fmt.Errorf("engine %q primary_skill is required", engine.ID)
		}
		if engine.Default {
			defaults++
		} else if len(engine.Triggers) == 0 {
			return fmt.Errorf("non-default engine %q requires triggers", engine.ID)
		}
		if err := validateStringList("engine "+engine.ID+" triggers", engine.Triggers, false); err != nil {
			return err
		}
		if coverage[engine.ID] == 0 {
			return fmt.Errorf("engine has no complete source package: %s", engine.ID)
		}
	}
	if defaults != 1 {
		return fmt.Errorf("exactly one default engine is required, got %d", defaults)
	}
	return nil
}

func validateExperts(experts []Expert) error {
	ids := map[string]bool{}
	for _, expert := range experts {
		if err := validateComponentID("expert", expert.ID, ids); err != nil {
			return err
		}
		if strings.TrimSpace(expert.Purpose) == "" || strings.TrimSpace(expert.ModelClass) == "" {
			return fmt.Errorf("expert %q requires purpose and model_class", expert.ID)
		}
		if _, ok := modelPolicies[strings.ToLower(strings.TrimSpace(expert.ModelClass))]; !ok {
			return fmt.Errorf("expert %q uses unsupported model_class %q", expert.ID, expert.ModelClass)
		}
	}
	return nil
}

func validateComponentID(kind, id string, seen map[string]bool) error {
	if !componentIDPattern.MatchString(id) {
		return fmt.Errorf("%s id %q is invalid", kind, id)
	}
	key := strings.ToLower(id)
	if seen[key] {
		return fmt.Errorf("%s id %q is duplicated", kind, id)
	}
	seen[key] = true
	return nil
}

func validateStringList(name string, values []string, required bool) error {
	if required && len(values) == 0 {
		return fmt.Errorf("%s must not be empty", name)
	}
	seen := map[string]bool{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return fmt.Errorf("%s contains an empty value", name)
		}
		key := strings.ToLower(trimmed)
		if seen[key] {
			return fmt.Errorf("%s contains duplicate value %q", name, value)
		}
		seen[key] = true
	}
	return nil
}

func LoadManifests(root string) ([]Manifest, error) {
	paths, err := filepath.Glob(filepath.Join(root, "capabilities", "*", "manifest.json"))
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no capability manifests under %s", root)
	}
	sort.Strings(paths)
	items := make([]Manifest, 0, len(paths))
	for _, path := range paths {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, readErr
		}
		item, jsonErr := decodeManifest(data)
		if jsonErr != nil {
			return nil, fmt.Errorf("%s: %w", path, jsonErr)
		}
		item.Root = root
		if validationErr := ValidateManifest(item); validationErr != nil {
			return nil, fmt.Errorf("%s: %w", path, validationErr)
		}
		directoryID := filepath.Base(filepath.Dir(path))
		if directoryID != item.ID {
			return nil, fmt.Errorf("%s: capability id %q must match directory %q", path, item.ID, directoryID)
		}
		items = append(items, item)
	}
	return items, nil
}

func ExpandPath(value string) string {
	return ExpandPathAt("", value)
}

func ExpandPathAt(root, value string) string {
	if root != "" {
		value = strings.ReplaceAll(value, "${ROOT}", root)
	}
	projects := os.Getenv("WUJI_PROJECTS")
	if projects == "" {
		if root != "" {
			projects = filepath.Clean(filepath.Join(root, ".."))
		} else if profile := os.Getenv("USERPROFILE"); profile != "" {
			candidate := filepath.Join(profile, "wuji-projects")
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				projects = candidate
			}
		}
	}
	if projects != "" {
		value = strings.ReplaceAll(value, "${WUJI_PROJECTS}", projects)
	}
	if profile := os.Getenv("USERPROFILE"); profile != "" {
		value = strings.ReplaceAll(value, "${USERPROFILE}", profile)
	}
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		value = strings.ReplaceAll(value, "${LOCALAPPDATA}", localAppData)
	}
	return filepath.Clean(value)
}

func ResolveSource(source Source) (string, bool) {
	return ResolveSourceAt("", source)
}

func ResolveSourceAt(root string, source Source) (string, bool) {
	return resolveSourceAt(root, source, false)
}

func ResolveCompleteSourceAt(root string, source Source) (string, bool) {
	return resolveSourceAt(root, source, true)
}

func resolveSourceAt(root string, source Source, requireComplete bool) (string, bool) {
	for _, raw := range source.Globs {
		matches, _ := filepath.Glob(ExpandPathAt(root, raw))
		sort.Slice(matches, func(i, j int) bool { return naturalCompare(matches[i], matches[j]) > 0 })
		for _, match := range matches {
			if info, err := os.Stat(match); err == nil && info.IsDir() {
				if !requireComplete || sourceComplete(match, source.Required) {
					return match, true
				}
			}
		}
	}
	return "", false
}

func sourceComplete(path string, required []string) bool {
	for _, pattern := range required {
		matches, err := filepath.Glob(filepath.Join(path, filepath.FromSlash(pattern)))
		if err != nil || len(matches) == 0 {
			return false
		}
	}
	return true
}

func decodeManifest(data []byte) (Manifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var item Manifest
	if err := decoder.Decode(&item); err != nil {
		return Manifest{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return Manifest{}, fmt.Errorf("multiple JSON values are not allowed")
		}
		return Manifest{}, err
	}
	return item, nil
}

func naturalCompare(left, right string) int {
	leftParts := naturalPartPattern.FindAllString(strings.ToLower(left), -1)
	rightParts := naturalPartPattern.FindAllString(strings.ToLower(right), -1)
	for i := 0; i < len(leftParts) && i < len(rightParts); i++ {
		leftPart, rightPart := leftParts[i], rightParts[i]
		leftDigit := leftPart[0] >= '0' && leftPart[0] <= '9'
		rightDigit := rightPart[0] >= '0' && rightPart[0] <= '9'
		if leftDigit && rightDigit {
			leftNumber := strings.TrimLeft(leftPart, "0")
			rightNumber := strings.TrimLeft(rightPart, "0")
			if leftNumber == "" {
				leftNumber = "0"
			}
			if rightNumber == "" {
				rightNumber = "0"
			}
			if len(leftNumber) != len(rightNumber) {
				if len(leftNumber) > len(rightNumber) {
					return 1
				}
				return -1
			}
			if leftNumber != rightNumber {
				if leftNumber > rightNumber {
					return 1
				}
				return -1
			}
		} else if leftPart != rightPart {
			if leftPart > rightPart {
				return 1
			}
			return -1
		}
	}
	if len(leftParts) > len(rightParts) {
		return 1
	}
	if len(leftParts) < len(rightParts) {
		return -1
	}
	return 0
}

func FindManifest(items []Manifest, id string) (Manifest, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return Manifest{}, false
}
