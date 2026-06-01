package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type jsonObject map[string]any

type slideSummary struct {
	Name       string `json:"name"`
	TextCount  int    `json:"text_count"`
	PicCount   int    `json:"pic_count"`
	ShapeCount int    `json:"shape_count"`
}

type pptxSummary struct {
	PPTXPath string         `json:"pptx_path"`
	Slides   []slideSummary `json:"slides"`
	Media    []string       `json:"media"`
	Layouts  []string       `json:"layouts"`
	Themes   []string       `json:"themes"`
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  wuji-cli reference-guard --reference <file>... --output <path>...")
	fmt.Fprintln(os.Stderr, "  wuji-cli workflow-guard --workspace <dir> [--stage scaffold|final]")
	fmt.Fprintln(os.Stderr, "  wuji-cli claim-guard --claim <text> [--evidence <file>]...")
	fmt.Fprintln(os.Stderr, "  wuji-cli time-guard --kind <non-code|general> --elapsed-minutes <n> [--artifact <file>] [--phase <name>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli task --workspace <dir> --event <start|heartbeat|blocked|end> [--status <value>] [--artifact <file>]... [--note <text>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli sync --source <dir> --dest <dir>")
	fmt.Fprintln(os.Stderr, "  wuji-cli audit --path <dir> [--report <file>] [--sarif <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli bench --workspace <dir> --name <run> [--input-tokens <n>] [--output-tokens <n>] [--duration-ms <n>] [--tool-calls <n>] [--retries <n>] [--qa-pass <true|false>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli bench-report --workspace <dir> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli preview --command <exe> [--arg <value>]... --output <file>")
	fmt.Fprintln(os.Stderr, "  wuji-cli asset-map --pptx <file> --workspace <dir>")
	fmt.Fprintln(os.Stderr, "  wuji-cli pptx-audit --pptx <file> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli pptx-preflight --workspace <dir> [--generator <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli pptx-batch-gate --workspace <dir> [--generator <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli mcp-guard --manifest <file> [--workspace <dir>] [--allow-network true|false]")
	fmt.Fprintln(os.Stderr, "  wuji-cli mcp-distill --catalog <file> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli route-task --config <file> --query <text> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli context-pack --config <file> --workspace <dir> --query <text> [--artifact <file>]... [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli feedback-log --workspace <dir> --task <text> [--prefer <term>]... [--avoid <term>]... [--note <text>] [--source <user|qa|audit>] [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli feedback-dataset --log <file> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli prompt-candidate-audit --candidate <file> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli prompt-eval --candidate <file> --dataset <file> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli prompt-distill --baseline <file> --candidate <file> --dataset <file> [--report <file>]")
}

func argValue(args []string, name string) (string, bool) {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == name {
			return args[i+1], true
		}
	}
	return "", false
}

func argValues(args []string, name string) []string {
	values := []string{}
	for i := 0; i < len(args)-1; i++ {
		if args[i] == name {
			values = append(values, args[i+1])
		}
	}
	return values
}

func parseBoolValue(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean: %s", value)
	}
}

func parseIntArg(args []string, name string) (int, bool, error) {
	value, ok := argValue(args, name)
	if !ok {
		return 0, false, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, true, err
	}
	return parsed, true, nil
}

func printGate(name string, failures []string) int {
	if len(failures) == 0 {
		fmt.Printf("GO %s\n", name)
		return 0
	}
	fmt.Printf("NO-GO %s\n", name)
	for _, failure := range failures {
		fmt.Printf("- %s\n", failure)
	}
	return 1
}

func absClean(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(abs)
}

func sameOrDescendant(path string, parent string) bool {
	path = strings.ToLower(absClean(path))
	parent = strings.ToLower(absClean(parent))
	return path == parent || strings.HasPrefix(path, parent+string(os.PathSeparator))
}

func nonEmpty(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Size() >= 20
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func appendJSONLine(path string, obj jsonObject) error {
	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = file.Write(append(bytes, '\n'))
	return err
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func writeJSON(path string, value any) error {
	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}
	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(bytes, '\n'), 0o644)
}

func writeAuditSARIF(path string, findings []jsonObject) error {
	rulesByID := map[string]jsonObject{}
	results := []jsonObject{}
	for _, finding := range findings {
		kind, _ := finding["kind"].(string)
		file, _ := finding["file"].(string)
		pattern, _ := finding["pattern"].(string)
		if kind == "" {
			kind = "wuji-audit-finding"
		}
		if _, ok := rulesByID[kind]; !ok {
			rulesByID[kind] = jsonObject{
				"id": kind,
				"shortDescription": jsonObject{
					"text": kind,
				},
				"help": jsonObject{
					"text": "Wuji execution-base deterministic audit finding.",
				},
			}
		}
		results = append(results, jsonObject{
			"ruleId": kind,
			"level":  "error",
			"message": jsonObject{
				"text": pattern,
			},
			"locations": []jsonObject{
				{
					"physicalLocation": jsonObject{
						"artifactLocation": jsonObject{
							"uri": filepath.ToSlash(file),
						},
					},
				},
			},
		})
	}
	ruleIDs := make([]string, 0, len(rulesByID))
	for ruleID := range rulesByID {
		ruleIDs = append(ruleIDs, ruleID)
	}
	sort.Strings(ruleIDs)
	rules := []jsonObject{}
	for _, ruleID := range ruleIDs {
		rules = append(rules, rulesByID[ruleID])
	}
	sarif := jsonObject{
		"$schema": "https://json.schemastore.org/sarif-2.1.0.json",
		"version": "2.1.0",
		"runs": []jsonObject{
			{
				"tool": jsonObject{
					"driver": jsonObject{
						"name":           "wuji-cli audit",
						"informationUri": "https://go.dev/doc/security/",
						"rules":          rules,
					},
				},
				"results": results,
			},
		},
	}
	return writeJSON(path, sarif)
}

func objectString(obj map[string]any, key string) string {
	value, ok := obj[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func objectBool(obj map[string]any, key string) (bool, bool) {
	value, ok := obj[key]
	if !ok {
		return false, false
	}
	parsed, ok := value.(bool)
	return parsed, ok
}

func objectFloat(obj map[string]any, key string) (float64, bool) {
	value, ok := obj[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	default:
		return 0, false
	}
}

func objectMap(obj map[string]any, key string) (map[string]any, bool) {
	value, ok := obj[key]
	if !ok {
		return nil, false
	}
	parsed, ok := value.(map[string]any)
	return parsed, ok
}

func objectSlice(obj map[string]any, key string) ([]any, bool) {
	value, ok := obj[key]
	if !ok {
		return nil, false
	}
	parsed, ok := value.([]any)
	return parsed, ok
}

func stringSlice(obj map[string]any, key string) []string {
	values, ok := objectSlice(obj, key)
	if !ok {
		return []string{}
	}
	result := []string{}
	for _, value := range values {
		text, ok := value.(string)
		if ok && strings.TrimSpace(text) != "" {
			result = append(result, strings.TrimSpace(text))
		}
	}
	return result
}

func intFromAny(value any) (int, bool) {
	switch typed := value.(type) {
	case float64:
		return int(typed), true
	case int:
		return typed, true
	default:
		return 0, false
	}
}

func boolFromAny(value any) (bool, bool) {
	typed, ok := value.(bool)
	return typed, ok
}

func readZipFile(file *zip.File) (string, error) {
	reader, err := file.Open()
	if err != nil {
		return "", err
	}
	defer reader.Close()
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func analyzePPTX(pptxPath string) (pptxSummary, error) {
	summary := pptxSummary{PPTXPath: absClean(pptxPath)}
	reader, err := zip.OpenReader(pptxPath)
	if err != nil {
		return summary, err
	}
	defer reader.Close()
	for _, file := range reader.File {
		name := strings.ReplaceAll(file.Name, "\\", "/")
		switch {
		case strings.HasPrefix(name, "ppt/slides/slide") && strings.HasSuffix(name, ".xml") && !strings.Contains(name, "/_rels/"):
			text, err := readZipFile(file)
			if err != nil {
				return summary, err
			}
			summary.Slides = append(summary.Slides, slideSummary{
				Name:       filepath.Base(name),
				TextCount:  strings.Count(text, "<a:t>"),
				PicCount:   strings.Count(text, "<p:pic"),
				ShapeCount: strings.Count(text, "<p:sp"),
			})
		case strings.HasPrefix(name, "ppt/media/"):
			summary.Media = append(summary.Media, filepath.Base(name))
		case strings.HasPrefix(name, "ppt/slideLayouts/") && strings.HasSuffix(name, ".xml"):
			summary.Layouts = append(summary.Layouts, filepath.Base(name))
		case strings.HasPrefix(name, "ppt/theme/") && strings.HasSuffix(name, ".xml"):
			summary.Themes = append(summary.Themes, filepath.Base(name))
		}
	}
	sort.Slice(summary.Slides, func(i, j int) bool { return summary.Slides[i].Name < summary.Slides[j].Name })
	sort.Strings(summary.Media)
	sort.Strings(summary.Layouts)
	sort.Strings(summary.Themes)
	return summary, nil
}

func referenceGuard(args []string) int {
	references := argValues(args, "--reference")
	outputs := argValues(args, "--output")
	if len(references) == 0 || len(outputs) == 0 {
		usage()
		return 2
	}
	failures := []string{}
	for _, reference := range references {
		if _, err := os.Stat(reference); err != nil {
			failures = append(failures, fmt.Sprintf("reference_missing=%s", reference))
		}
		for _, output := range outputs {
			if sameOrDescendant(output, reference) {
				failures = append(failures, fmt.Sprintf("output_overlaps_reference=%s -> %s", output, reference))
			}
		}
	}
	return printGate("reference-guard", failures)
}

func requireNonEmpty(path string, failures *[]string) {
	if !nonEmpty(path) {
		if _, err := os.Stat(path); err != nil {
			*failures = append(*failures, fmt.Sprintf("missing_required_file=%s", path))
		} else {
			*failures = append(*failures, fmt.Sprintf("required_file_too_small=%s", path))
		}
	}
}

func hasMarkdownFile(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			return true
		}
	}
	return false
}

func workflowGuard(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	stage, ok := argValue(args, "--stage")
	if !ok {
		stage = "scaffold"
	}
	if stage != "scaffold" && stage != "final" {
		fmt.Fprintln(os.Stderr, "stage must be scaffold or final")
		return 2
	}
	failures := []string{}
	if info, err := os.Stat(workspace); err != nil || !info.IsDir() {
		failures = append(failures, fmt.Sprintf("workspace_missing=%s", workspace))
		return printGate("workflow-guard", failures)
	}
	for _, name := range []string{"contract.md", "state.json", "final-report.md"} {
		requireNonEmpty(filepath.Join(workspace, name), &failures)
	}
	for _, name := range []string{"packets", "results"} {
		if info, err := os.Stat(filepath.Join(workspace, name)); err != nil || !info.IsDir() {
			failures = append(failures, fmt.Sprintf("missing_required_dir=%s", filepath.Join(workspace, name)))
		}
	}
	if stage == "final" {
		if !hasMarkdownFile(filepath.Join(workspace, "packets")) {
			failures = append(failures, "final_requires_packet_file")
		}
		if !hasMarkdownFile(filepath.Join(workspace, "results")) {
			failures = append(failures, "final_requires_result_file")
		}
		report, _ := os.ReadFile(filepath.Join(workspace, "final-report.md"))
		if !strings.Contains(string(report), "Verification Evidence") {
			failures = append(failures, "final_report_missing_verification_evidence_heading")
		}
	}
	return printGate("workflow-guard", failures)
}

func claimGuard(args []string) int {
	claim, ok := argValue(args, "--claim")
	if !ok {
		usage()
		return 2
	}
	evidence := argValues(args, "--evidence")
	successWords := []string{"完成", "成功", "通过", "已融合", "已生成", "verified", "passed", "complete", "completed", "success"}
	lower := strings.ToLower(claim)
	makesSuccessClaim := false
	for _, word := range successWords {
		if strings.Contains(claim, word) || strings.Contains(lower, word) {
			makesSuccessClaim = true
			break
		}
	}
	failures := []string{}
	if makesSuccessClaim && len(evidence) == 0 {
		failures = append(failures, "success_claim_requires_evidence")
	}
	for _, path := range evidence {
		if !nonEmpty(path) {
			failures = append(failures, fmt.Sprintf("evidence_missing_or_too_small=%s", path))
		}
	}
	return printGate("claim-guard", failures)
}

func timeGuard(args []string) int {
	kind, ok := argValue(args, "--kind")
	if !ok {
		usage()
		return 2
	}
	minutesText, ok := argValue(args, "--elapsed-minutes")
	if !ok {
		usage()
		return 2
	}
	minutes, err := strconv.Atoi(minutesText)
	if err != nil || minutes < 0 {
		fmt.Fprintln(os.Stderr, "elapsed-minutes must be a non-negative integer")
		return 2
	}
	phase, ok := argValue(args, "--phase")
	if !ok {
		phase = "unknown"
	}
	artifact, hasArtifactArg := argValue(args, "--artifact")
	hasArtifact := hasArtifactArg && nonEmpty(artifact)
	failures := []string{}
	if kind == "non-code" {
		explorePhase := map[string]bool{"explore": true, "research": true, "probe": true, "prototype": true, "preflight": true}
		if minutes >= 30 && !hasArtifact {
			failures = append(failures, fmt.Sprintf("no_verifiable_artifact_after_30_minutes phase=%s", phase))
		} else if minutes >= 15 && !hasArtifact && explorePhase[phase] {
			failures = append(failures, fmt.Sprintf("non_code_exploration_timeout_after_15_minutes phase=%s", phase))
		} else if minutes >= 10 && !hasArtifact {
			failures = append(failures, fmt.Sprintf("non_code_missing_primary_artifact_after_10_minutes phase=%s", phase))
		}
	}
	return printGate("time-guard", failures)
}

func taskCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	event, ok := argValue(args, "--event")
	if !ok {
		usage()
		return 2
	}
	status, _ := argValue(args, "--status")
	note, _ := argValue(args, "--note")
	artifacts := argValues(args, "--artifact")
	if err := ensureDir(workspace); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	entry := jsonObject{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"event":     event,
		"status":    status,
		"note":      note,
		"artifacts": artifacts,
	}
	logPath := filepath.Join(workspace, "task-log.jsonl")
	if err := appendJSONLine(logPath, entry); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("GO task\n- log=%s\n", logPath)
	return 0
}

func syncCommand(args []string) int {
	source, okSource := argValue(args, "--source")
	dest, okDest := argValue(args, "--dest")
	if !okSource || !okDest {
		usage()
		return 2
	}
	required := []string{"GLOBAL_AGENTS.md", "SKILL.md", "config.json", "README.md", "scripts", "units", "experts"}
	failures := []string{}
	for _, rel := range required {
		src := filepath.Join(source, rel)
		dst := filepath.Join(dest, rel)
		srcInfo, srcErr := os.Stat(src)
		dstInfo, dstErr := os.Stat(dst)
		if srcErr != nil {
			failures = append(failures, fmt.Sprintf("source_missing=%s", src))
			continue
		}
		if dstErr != nil {
			failures = append(failures, fmt.Sprintf("dest_missing=%s", dst))
			continue
		}
		if srcInfo.IsDir() != dstInfo.IsDir() {
			failures = append(failures, fmt.Sprintf("type_mismatch=%s", rel))
			continue
		}
		if !srcInfo.IsDir() {
			srcHash, srcErr := fileSHA256(src)
			dstHash, dstErr := fileSHA256(dst)
			if srcErr != nil || dstErr != nil || srcHash != dstHash {
				failures = append(failures, fmt.Sprintf("file_mismatch=%s", rel))
			}
		}
	}
	return printGate("sync", failures)
}

func auditCommand(args []string) int {
	root, ok := argValue(args, "--path")
	if !ok {
		usage()
		return 2
	}
	report, hasReport := argValue(args, "--report")
	sarif, hasSARIF := argValue(args, "--sarif")
	failures := []string{}
	findings := []jsonObject{}
	textExt := map[string]bool{
		".md": true, ".ps1": true, ".py": true, ".go": true, ".json": true, ".toml": true, ".yaml": true, ".yml": true, ".txt": true,
	}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			failures = append(failures, fmt.Sprintf("walk_error=%s", path))
			return nil
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == ".wuji-tools" || base == "__pycache__" || base == "outputs" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !textExt[ext] {
			return nil
		}
		bytes, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("read_error=%s", path))
			return nil
		}
		text := string(bytes)
		rel, _ := filepath.Rel(root, path)
		check := func(kind string, pattern string) {
			findings = append(findings, jsonObject{"file": rel, "kind": kind, "pattern": pattern})
			failures = append(failures, fmt.Sprintf("%s=%s", kind, rel))
		}
		replacementChar := string(rune(0xfffd))
		if strings.Contains(text, replacementChar) {
			check("encoding_replacement_char", "replacement-char")
		}
		staleRefs := []string{"units/" + "rust.md", "experts/" + "rust/", "Rust" + "师", "Rust" + "主帅"}
		hasStaleRef := false
		for _, staleRef := range staleRefs {
			if strings.Contains(text, staleRef) {
				hasStaleRef = true
				break
			}
		}
		if hasStaleRef {
			if !strings.HasSuffix(rel, "CHANGELOG.md") && !strings.HasSuffix(rel, filepath.Join("units", "distillation.md")) && !strings.HasSuffix(rel, filepath.Join("units", "nuwa.md")) {
				check("stale_rust_mainline_ref", "Rust mainline")
			}
		}
		unfinishedRefs := []string{"待" + "开发", "后续" + "路线"}
		for _, unfinishedRef := range unfinishedRefs {
			if strings.Contains(text, unfinishedRef) {
				check("unfinished_marker", "unfinished marker")
				break
			}
		}
		stackingRefs := []string{"A" + "/B", "a" + "/b", "并行" + "主线"}
		for _, stackingRef := range stackingRefs {
			if strings.Contains(text, stackingRef) {
				check("stacking_marker", "stacking marker")
				break
			}
		}
		secretPatterns := []string{
			"sk-" + "proj-",
			"sk-" + "live-",
			"sk-" + "ant-",
			"gh" + "p_",
			"github" + "_pat_",
			"xox" + "b-",
			"xox" + "p-",
			"-----BEGIN " + "PRIVATE KEY-----",
		}
		for _, secretPattern := range secretPatterns {
			if strings.Contains(text, secretPattern) {
				check("secret_leak_marker", "secret leak marker")
				break
			}
		}
		incompleteRefs := []string{"to" + "do", "t" + "bd"}
		lowerText := strings.ToLower(text)
		for _, incompleteRef := range incompleteRefs {
			if strings.Contains(lowerText, incompleteRef) {
				check("incomplete_marker", "incomplete marker")
				break
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if hasReport {
		_ = writeJSON(report, jsonObject{"path": absClean(root), "findings": findings})
	}
	if hasSARIF {
		_ = writeAuditSARIF(sarif, findings)
	}
	return printGate("audit", failures)
}

func benchCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	name, ok := argValue(args, "--name")
	if !ok {
		usage()
		return 2
	}
	inputTokens, _, err := parseIntArg(args, "--input-tokens")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	outputTokens, _, err := parseIntArg(args, "--output-tokens")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	durationMs, _, err := parseIntArg(args, "--duration-ms")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	toolCalls, _, err := parseIntArg(args, "--tool-calls")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	retries, _, err := parseIntArg(args, "--retries")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	var qaPass any = nil
	if qaValue, ok := argValue(args, "--qa-pass"); ok {
		parsed, err := parseBoolValue(qaValue)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		qaPass = parsed
	}
	entry := jsonObject{
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"name":          name,
		"input_tokens":  inputTokens,
		"output_tokens": outputTokens,
		"duration_ms":   durationMs,
		"tool_calls":    toolCalls,
		"retries":       retries,
		"qa_pass":       qaPass,
	}
	logPath := filepath.Join(workspace, "bench.jsonl")
	if err := appendJSONLine(logPath, entry); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("GO bench\n- log=%s\n", logPath)
	return 0
}

func benchReportCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	logPath := filepath.Join(workspace, "bench.jsonl")
	bytes, err := os.ReadFile(logPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	lines := strings.Split(strings.TrimSpace(string(bytes)), "\n")
	totalRuns := 0
	totalInput := 0
	totalOutput := 0
	totalDuration := 0
	totalToolCalls := 0
	totalRetries := 0
	qaPassCount := 0
	qaSeen := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		totalRuns++
		if value, ok := intFromAny(row["input_tokens"]); ok {
			totalInput += value
		}
		if value, ok := intFromAny(row["output_tokens"]); ok {
			totalOutput += value
		}
		if value, ok := intFromAny(row["duration_ms"]); ok {
			totalDuration += value
		}
		if value, ok := intFromAny(row["tool_calls"]); ok {
			totalToolCalls += value
		}
		if value, ok := intFromAny(row["retries"]); ok {
			totalRetries += value
		}
		if value, ok := boolFromAny(row["qa_pass"]); ok {
			qaSeen++
			if value {
				qaPassCount++
			}
		}
	}
	if totalRuns == 0 {
		fmt.Fprintln(os.Stderr, "bench log has no runs")
		return 1
	}
	avgDuration := totalDuration / totalRuns
	totalTokens := totalInput + totalOutput
	tokensPerMinute := 0
	if totalDuration > 0 {
		tokensPerMinute = int((float64(totalTokens) * 60000.0) / float64(totalDuration))
	}
	qaPassRate := 0.0
	if qaSeen > 0 {
		qaPassRate = float64(qaPassCount) / float64(qaSeen)
	}
	report := jsonObject{
		"workspace":         absClean(workspace),
		"runs":              totalRuns,
		"total_input":       totalInput,
		"total_output":      totalOutput,
		"total_tokens":      totalTokens,
		"avg_duration_ms":   avgDuration,
		"total_tool_calls":  totalToolCalls,
		"total_retries":     totalRetries,
		"qa_pass_rate":      qaPassRate,
		"tokens_per_minute": tokensPerMinute,
	}
	if hasReport {
		if err := writeJSON(reportPath, report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	fmt.Printf("GO bench-report\n- runs=%d\n- tokens_per_minute=%d\n", totalRuns, tokensPerMinute)
	return 0
}

func previewCommand(args []string) int {
	command, ok := argValue(args, "--command")
	if !ok {
		usage()
		return 2
	}
	output, ok := argValue(args, "--output")
	if !ok {
		usage()
		return 2
	}
	commandArgs := argValues(args, "--arg")
	cmd := exec.Command(command, commandArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	failures := []string{}
	if !nonEmpty(output) {
		failures = append(failures, fmt.Sprintf("preview_output_missing_or_too_small=%s", output))
	}
	return printGate("preview", failures)
}

func assetMapCommand(args []string) int {
	pptxPath, ok := argValue(args, "--pptx")
	workspace, okWorkspace := argValue(args, "--workspace")
	if !ok || !okWorkspace {
		usage()
		return 2
	}
	summary, err := analyzePPTX(pptxPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := ensureDir(workspace); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	frameLines := []string{"# reference-frame-map", ""}
	illustrationLines := []string{"# illustration-plan", ""}
	for i, slide := range summary.Slides {
		frameLines = append(frameLines, fmt.Sprintf("- slide-%02d %s: text=%d pic=%d shape=%d", i+1, slide.Name, slide.TextCount, slide.PicCount, slide.ShapeCount))
		illustrationMode := "无需插图"
		if slide.PicCount > 0 {
			illustrationMode = "复用参考图或参考图框"
		}
		illustrationLines = append(illustrationLines, fmt.Sprintf("- slide-%02d: %s", i+1, illustrationMode))
	}
	assetLines := []string{"# reusable-asset-map", "", "## media"}
	if len(summary.Media) == 0 {
		assetLines = append(assetLines, "- none")
	} else {
		for _, media := range summary.Media {
			assetLines = append(assetLines, "- "+media)
		}
	}
	assetLines = append(assetLines, "", "## layouts")
	if len(summary.Layouts) == 0 {
		assetLines = append(assetLines, "- none")
	} else {
		for _, layout := range summary.Layouts {
			assetLines = append(assetLines, "- "+layout)
		}
	}
	if err := os.WriteFile(filepath.Join(workspace, "reference-frame-map.md"), []byte(strings.Join(frameLines, "\n")+"\n"), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := os.WriteFile(filepath.Join(workspace, "illustration-plan.md"), []byte(strings.Join(illustrationLines, "\n")+"\n"), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := os.WriteFile(filepath.Join(workspace, "reusable-asset-map.md"), []byte(strings.Join(assetLines, "\n")+"\n"), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("GO asset-map\n- workspace=%s\n", workspace)
	return 0
}

func pptxAuditCommand(args []string) int {
	pptxPath, ok := argValue(args, "--pptx")
	if !ok {
		usage()
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	summary, err := analyzePPTX(pptxPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	failures := []string{}
	if len(summary.Slides) == 0 {
		failures = append(failures, "pptx_has_no_slides")
	}
	imageOnlySlides := []string{}
	for _, slide := range summary.Slides {
		if slide.PicCount > 0 && slide.TextCount == 0 && slide.ShapeCount <= 1 {
			imageOnlySlides = append(imageOnlySlides, slide.Name)
		}
	}
	if len(imageOnlySlides) > 0 {
		failures = append(failures, "pptx_contains_image_only_slides="+strings.Join(imageOnlySlides, ","))
	}
	report := jsonObject{
		"pptx_path":         summary.PPTXPath,
		"slide_count":       len(summary.Slides),
		"media_count":       len(summary.Media),
		"layout_count":      len(summary.Layouts),
		"image_only_slides": imageOnlySlides,
		"slides":            summary.Slides,
	}
	if hasReport {
		if err := writeJSON(reportPath, report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return printGate("pptx-audit", failures)
}

func existingPlanFile(workspace string, stem string) (string, bool) {
	for _, ext := range []string{"json", "md", "txt"} {
		path := filepath.Join(workspace, stem+"."+ext)
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

func scanGenerator(path string) []string {
	failures := []string{}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return append(failures, fmt.Sprintf("generator_unreadable=%s", path))
	}
	lower := strings.ToLower(string(bytes))
	fullSlidePicture := strings.Contains(lower, "add_picture") &&
		(strings.Contains(lower, "slide_width") || strings.Contains(lower, "slide_height")) &&
		(strings.Contains(lower, "0, 0") || strings.Contains(lower, "left=0") || strings.Contains(lower, "top=0"))
	if fullSlidePicture {
		failures = append(failures, "generator_looks_like_full_slide_picture_route")
	}
	rasterSlideRoute := strings.Contains(lower, "full-slide-image") ||
		strings.Contains(lower, "full_slide_image") ||
		strings.Contains(lower, "slide as image") ||
		strings.Contains(lower, "render entire slide") ||
		strings.Contains(lower, "每页一张") ||
		(strings.Contains(lower, "image.new") && strings.Contains(lower, "canvas.save"))
	if rasterSlideRoute {
		failures = append(failures, "generator_looks_like_raster_slide_route")
	}
	return failures
}

func pptxPreflight(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	failures := []string{}
	if info, err := os.Stat(workspace); err != nil || !info.IsDir() {
		failures = append(failures, fmt.Sprintf("workspace_missing=%s", workspace))
	}
	for _, stem := range []string{"reference-frame-map", "reusable-asset-map", "illustration-plan"} {
		if path, ok := existingPlanFile(workspace, stem); ok {
			if !nonEmpty(path) {
				failures = append(failures, fmt.Sprintf("plan_file_too_small=%s", path))
			}
		} else {
			failures = append(failures, fmt.Sprintf("missing_required_plan=%s", stem))
		}
	}
	if generator, ok := argValue(args, "--generator"); ok {
		failures = append(failures, scanGenerator(generator)...)
	}
	return printGate("pptx-preflight", failures)
}

func existingPilotFile(workspace string, stem string) (string, bool) {
	for _, ext := range []string{"png", "jpg", "jpeg", "pptx", "json", "md", "txt"} {
		path := filepath.Join(workspace, stem+"."+ext)
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

func pptxBatchGate(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	failures := []string{}
	if info, err := os.Stat(workspace); err != nil || !info.IsDir() {
		failures = append(failures, fmt.Sprintf("workspace_missing=%s", workspace))
	}
	for _, stem := range []string{"reference-frame-map", "reusable-asset-map", "illustration-plan"} {
		if path, ok := existingPlanFile(workspace, stem); ok {
			if !nonEmpty(path) {
				failures = append(failures, fmt.Sprintf("plan_file_too_small=%s", path))
			}
		} else {
			failures = append(failures, fmt.Sprintf("missing_required_plan=%s", stem))
		}
	}
	for _, stem := range []string{"pilot-preview", "pilot-page"} {
		if path, ok := existingPilotFile(workspace, stem); ok {
			if !nonEmpty(path) {
				failures = append(failures, fmt.Sprintf("pilot_file_too_small=%s", path))
			}
		} else {
			failures = append(failures, fmt.Sprintf("missing_required_pilot=%s", stem))
		}
	}
	if path, ok := existingPlanFile(workspace, "pilot-score"); ok {
		if !nonEmpty(path) {
			failures = append(failures, fmt.Sprintf("pilot_score_too_small=%s", path))
		}
	} else {
		failures = append(failures, "missing_required_pilot=pilot-score")
	}
	if generator, ok := argValue(args, "--generator"); ok {
		failures = append(failures, scanGenerator(generator)...)
	}
	return printGate("pptx-batch-gate", failures)
}

func mcpGuard(args []string) int {
	manifestPath, ok := argValue(args, "--manifest")
	if !ok {
		usage()
		return 2
	}
	allowNetwork := false
	if value, ok := argValue(args, "--allow-network"); ok {
		parsed, err := parseBoolValue(value)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		allowNetwork = parsed
	}
	workspace := ""
	if value, ok := argValue(args, "--workspace"); ok {
		workspace = absClean(value)
	}
	rawManifest, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	rawManifest = bytes.TrimPrefix(rawManifest, []byte{0xef, 0xbb, 0xbf})
	var manifest map[string]any
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	failures := []string{}
	for _, field := range []string{"name", "version"} {
		if objectString(manifest, field) == "" {
			failures = append(failures, fmt.Sprintf("manifest_missing_%s", field))
		}
	}
	transport := strings.ToLower(objectString(manifest, "transport"))
	if transport == "" {
		failures = append(failures, "manifest_missing_transport")
	} else if transport != "stdio" && transport != "http" && transport != "sse" {
		failures = append(failures, "manifest_unknown_transport="+transport)
	}
	if (transport == "http" || transport == "sse") && !allowNetwork {
		failures = append(failures, "network_transport_requires_explicit_allow")
	}
	capabilities, hasCapabilities := objectMap(manifest, "capabilities")
	if !hasCapabilities {
		failures = append(failures, "manifest_missing_capabilities")
	} else {
		for _, name := range []string{"tools", "resources", "prompts"} {
			if _, ok := capabilities[name]; !ok {
				failures = append(failures, "capability_missing_"+name)
			}
		}
	}
	permissions, hasPermissions := objectMap(manifest, "permissions")
	if !hasPermissions {
		failures = append(failures, "manifest_missing_permissions")
	} else {
		if network, ok := objectBool(permissions, "network"); ok && network && !allowNetwork {
			failures = append(failures, "network_permission_requires_explicit_allow")
		}
		if filesystem, ok := objectSlice(permissions, "filesystem"); ok {
			if workspace == "" && len(filesystem) > 0 {
				failures = append(failures, "filesystem_permission_requires_workspace")
			}
			for _, rawPath := range filesystem {
				path, ok := rawPath.(string)
				if !ok || strings.TrimSpace(path) == "" {
					failures = append(failures, "filesystem_permission_invalid_path")
					continue
				}
				if workspace != "" && !sameOrDescendant(path, workspace) {
					failures = append(failures, "filesystem_permission_outside_workspace="+path)
				}
			}
		}
	}
	lower := strings.ToLower(string(rawManifest))
	secretMarkers := []string{"api_key", "apikey", "secret", "token", "password", "bearer "}
	for _, marker := range secretMarkers {
		if strings.Contains(lower, marker) {
			failures = append(failures, "manifest_contains_secret_marker="+marker)
			break
		}
	}
	return printGate("mcp-guard", failures)
}

func mcpDistill(args []string) int {
	catalogPath, ok := argValue(args, "--catalog")
	if !ok {
		usage()
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	rawCatalog, err := os.ReadFile(catalogPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	rawCatalog = bytes.TrimPrefix(rawCatalog, []byte{0xef, 0xbb, 0xbf})
	var catalog map[string]any
	if err := json.Unmarshal(rawCatalog, &catalog); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	items, ok := objectSlice(catalog, "candidates")
	if !ok || len(items) == 0 {
		fmt.Fprintln(os.Stderr, "catalog must contain candidates")
		return 2
	}
	failures := []string{}
	decisions := []jsonObject{}
	seenCapabilities := map[string]bool{}
	for index, rawItem := range items {
		item, ok := rawItem.(map[string]any)
		if !ok {
			failures = append(failures, fmt.Sprintf("candidate_%d_invalid", index+1))
			continue
		}
		name := objectString(item, "name")
		owner := objectString(item, "owner")
		source := objectString(item, "source")
		license := objectString(item, "license")
		capability := objectString(item, "capability")
		transport := strings.ToLower(objectString(item, "transport"))
		permissions := stringSlice(item, "permissions")
		risk := strings.ToLower(objectString(item, "risk"))
		if name == "" || owner == "" || source == "" || capability == "" {
			failures = append(failures, fmt.Sprintf("candidate_%d_missing_required_field", index+1))
		}
		if license == "" || strings.EqualFold(license, "unknown") {
			failures = append(failures, fmt.Sprintf("candidate_%s_license_unknown", name))
		}
		capabilityKey := strings.ToLower(owner + "::" + capability)
		if seenCapabilities[capabilityKey] {
			failures = append(failures, fmt.Sprintf("duplicate_capability=%s", capabilityKey))
		}
		seenCapabilities[capabilityKey] = true
		decision := "absorb"
		reason := "bounded local tool capability"
		if source == "" || license == "" || strings.EqualFold(license, "unknown") {
			decision = "reject"
			reason = "missing source or license"
		} else if strings.Contains(risk, "high") || containsString(permissions, "secrets") || containsString(permissions, "write-all") {
			decision = "reject"
			reason = "high risk permission"
		} else if transport == "http" || transport == "sse" || containsString(permissions, "network") || containsString(permissions, "oauth") {
			decision = "defer"
			reason = "network or account permission requires task-specific approval"
		}
		decisions = append(decisions, jsonObject{
			"name":        name,
			"owner":       owner,
			"source":      source,
			"license":     license,
			"capability":  capability,
			"transport":   transport,
			"permissions": permissions,
			"decision":    decision,
			"reason":      reason,
		})
	}
	sort.Slice(decisions, func(i, j int) bool {
		return fmt.Sprint(decisions[i]["name"]) < fmt.Sprint(decisions[j]["name"])
	})
	if hasReport {
		if err := writeJSON(reportPath, jsonObject{"catalog": absClean(catalogPath), "decisions": decisions}); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return printGate("mcp-distill", failures)
}

func containsString(values []string, needle string) bool {
	needle = strings.ToLower(needle)
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), needle) {
			return true
		}
	}
	return false
}

func containsSecretLikeContent(text string) bool {
	lower := strings.ToLower(text)
	markers := []string{
		`"api_key":`,
		`"apikey":`,
		`"password":`,
		`"cookie":`,
		`"authorization":`,
		"bearer ",
		"sk-",
		"gh" + "p_",
		"gh" + "o_",
		"xoxb" + "-",
		"secret=",
		`"secret":`,
	}
	for _, marker := range markers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func loadJSONObject(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	raw = bytes.TrimPrefix(raw, []byte{0xef, 0xbb, 0xbf})
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func loadJSONLines(path string) ([]map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	raw = bytes.TrimPrefix(raw, []byte{0xef, 0xbb, 0xbf})
	lines := strings.Split(string(raw), "\n")
	records := []map[string]any{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			return nil, err
		}
		records = append(records, obj)
	}
	return records, nil
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func feedbackLogCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	task, ok := argValue(args, "--task")
	if !ok || strings.TrimSpace(task) == "" {
		usage()
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	preferTerms := uniqueStrings(argValues(args, "--prefer"))
	avoidTerms := uniqueStrings(argValues(args, "--avoid"))
	note, _ := argValue(args, "--note")
	source := "user"
	if value, ok := argValue(args, "--source"); ok && strings.TrimSpace(value) != "" {
		source = strings.TrimSpace(value)
	}
	if len(preferTerms) == 0 && len(avoidTerms) == 0 {
		fmt.Fprintln(os.Stderr, "feedback-log requires at least one --prefer or --avoid term")
		return 2
	}
	secretChecks := append([]string{task, note, source}, preferTerms...)
	secretChecks = append(secretChecks, avoidTerms...)
	for _, item := range secretChecks {
		if containsSecretLikeContent(item) {
			fmt.Fprintln(os.Stderr, "feedback-log contains secret-like content")
			return 1
		}
	}
	logPath := filepath.Join(workspace, "feedback", "feedback-log.jsonl")
	entry := jsonObject{
		"task":         strings.TrimSpace(task),
		"prefer_terms": preferTerms,
		"avoid_terms":  avoidTerms,
		"note":         strings.TrimSpace(note),
		"source":       source,
		"logged_at":    time.Now().Format(time.RFC3339),
	}
	if err := appendJSONLine(logPath, entry); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	report := jsonObject{
		"log":          absClean(logPath),
		"task":         entry["task"],
		"prefer_terms": preferTerms,
		"avoid_terms":  avoidTerms,
		"source":       source,
	}
	if hasReport {
		if err := writeJSON(reportPath, report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	fmt.Printf("GO feedback-log\n- log=%s\n", absClean(logPath))
	return 0
}

func feedbackDatasetCommand(args []string) int {
	logPath, ok := argValue(args, "--log")
	if !ok {
		usage()
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	records, err := loadJSONLines(logPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	cases := []jsonObject{}
	allPrefer := []string{}
	allAvoid := []string{}
	for index, record := range records {
		preferTerms := uniqueStrings(stringSlice(record, "prefer_terms"))
		avoidTerms := uniqueStrings(stringSlice(record, "avoid_terms"))
		task := objectString(record, "task")
		if len(preferTerms) == 0 && len(avoidTerms) == 0 {
			continue
		}
		caseID := fmt.Sprintf("feedback-%02d", index+1)
		cases = append(cases, jsonObject{
			"id":              caseID,
			"task":            task,
			"required_terms":  preferTerms,
			"forbidden_terms": avoidTerms,
		})
		allPrefer = append(allPrefer, preferTerms...)
		allAvoid = append(allAvoid, avoidTerms...)
	}
	if len(cases) == 0 {
		fmt.Fprintln(os.Stderr, "feedback-dataset found no usable feedback cases")
		return 1
	}
	report := jsonObject{
		"log": absClean(logPath),
		"summary": jsonObject{
			"cases":        len(cases),
			"prefer_terms": uniqueStrings(allPrefer),
			"avoid_terms":  uniqueStrings(allAvoid),
		},
		"cases": cases,
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(filepath.Dir(logPath), "feedback-dataset.json")
	}
	if err := writeJSON(outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("GO feedback-dataset\n- report=%s\n", absClean(outputPath))
	return 0
}

func promptCandidateAudit(args []string) int {
	candidatePath, ok := argValue(args, "--candidate")
	if !ok {
		usage()
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	candidate, err := loadJSONObject(candidatePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	failures := []string{}
	warnings := []string{}
	requiredFields := []string{"name", "objective", "prompt_template", "metric"}
	for _, field := range requiredFields {
		if objectString(candidate, field) == "" {
			failures = append(failures, "candidate_missing_"+field)
		}
	}
	promptTemplate := objectString(candidate, "prompt_template")
	if strings.TrimSpace(promptTemplate) != "" {
		if len(promptTemplate) < 80 {
			warnings = append(warnings, "prompt_template_may_be_under_specified")
		}
		if strings.Contains(strings.ToLower(promptTemplate), "todo") || strings.Contains(promptTemplate, "待补") {
			failures = append(failures, "prompt_template_contains_placeholder")
		}
	}
	stablePrefix := objectString(candidate, "stable_prefix")
	if strings.TrimSpace(stablePrefix) == "" {
		warnings = append(warnings, "candidate_missing_stable_prefix")
	} else if len(stablePrefix) < 40 {
		warnings = append(warnings, "stable_prefix_may_be_too_short_for_cache")
	}
	variables := stringSlice(candidate, "variables")
	if len(variables) == 0 {
		warnings = append(warnings, "candidate_missing_variables")
	}
	candidateBytes, _ := json.Marshal(candidate)
	if containsSecretLikeContent(string(candidateBytes)) {
		failures = append(failures, "candidate_contains_secret_like_content")
	}
	report := jsonObject{
		"candidate": absClean(candidatePath),
		"name":      objectString(candidate, "name"),
		"objective": objectString(candidate, "objective"),
		"variables": variables,
		"warnings":  warnings,
		"failures":  failures,
	}
	if hasReport {
		if err := writeJSON(reportPath, report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return printGate("prompt-candidate-audit", failures)
}

func promptEvalCommand(args []string) int {
	candidatePath, ok := argValue(args, "--candidate")
	if !ok {
		usage()
		return 2
	}
	datasetPath, ok := argValue(args, "--dataset")
	if !ok {
		usage()
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	candidate, err := loadJSONObject(candidatePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	dataset, err := loadJSONObject(datasetPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	promptTemplate := strings.ToLower(objectString(candidate, "prompt_template"))
	stablePrefix := strings.TrimSpace(objectString(candidate, "stable_prefix"))
	cases, ok := objectSlice(dataset, "cases")
	if !ok || len(cases) == 0 {
		fmt.Fprintln(os.Stderr, "dataset must contain cases")
		return 2
	}
	failures := []string{}
	warnings := []string{}
	totalChecks := 0
	passedChecks := 0
	caseReports := []jsonObject{}
	for index, rawCase := range cases {
		item, ok := rawCase.(map[string]any)
		if !ok {
			failures = append(failures, fmt.Sprintf("case_%d_invalid", index+1))
			continue
		}
		caseID := objectString(item, "id")
		if caseID == "" {
			caseID = fmt.Sprintf("case-%d", index+1)
		}
		requiredTerms := stringSlice(item, "required_terms")
		forbiddenTerms := stringSlice(item, "forbidden_terms")
		caseFailures := []string{}
		casePasses := 0
		caseChecks := 0
		for _, term := range requiredTerms {
			caseChecks++
			totalChecks++
			if strings.Contains(promptTemplate, strings.ToLower(term)) {
				casePasses++
				passedChecks++
			} else {
				caseFailures = append(caseFailures, "missing_required_term="+term)
			}
		}
		for _, term := range forbiddenTerms {
			caseChecks++
			totalChecks++
			if strings.Contains(promptTemplate, strings.ToLower(term)) {
				caseFailures = append(caseFailures, "contains_forbidden_term="+term)
			} else {
				casePasses++
				passedChecks++
			}
		}
		caseReports = append(caseReports, jsonObject{
			"id":       caseID,
			"checks":   caseChecks,
			"passed":   casePasses,
			"failures": caseFailures,
		})
	}
	cacheScore := 0.0
	if stablePrefix != "" {
		cacheScore = 1.0
		if len(stablePrefix) < 80 {
			cacheScore = 0.7
		}
		if len(stablePrefix) < 40 {
			cacheScore = 0.4
		}
	}
	simplicityScore := 1.0
	if len(promptTemplate) > 3000 {
		simplicityScore = 0.5
	} else if len(promptTemplate) > 1800 {
		simplicityScore = 0.75
	}
	coverageScore := 0.0
	if totalChecks > 0 {
		coverageScore = float64(passedChecks) / float64(totalChecks)
	}
	compositeScore := (coverageScore * 0.6) + (cacheScore * 0.25) + (simplicityScore * 0.15)
	if coverageScore < 0.85 {
		warnings = append(warnings, "prompt_coverage_below_threshold")
	}
	report := jsonObject{
		"candidate":         absClean(candidatePath),
		"dataset":           absClean(datasetPath),
		"metric":            objectString(candidate, "metric"),
		"coverage_score":    coverageScore,
		"cache_score":       cacheScore,
		"simplicity_score":  simplicityScore,
		"composite_score":   compositeScore,
		"stable_prefix_len": len(stablePrefix),
		"template_len":      len(promptTemplate),
		"checks":            totalChecks,
		"passed":            passedChecks,
		"case_reports":      caseReports,
		"warnings":          warnings,
		"failures":          failures,
	}
	if hasReport {
		if err := writeJSON(reportPath, report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return printGate("prompt-eval", failures)
}

func promptDistillCommand(args []string) int {
	baselinePath, ok := argValue(args, "--baseline")
	if !ok {
		usage()
		return 2
	}
	candidatePath, ok := argValue(args, "--candidate")
	if !ok {
		usage()
		return 2
	}
	datasetPath, ok := argValue(args, "--dataset")
	if !ok {
		usage()
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	tmpDir := os.TempDir()
	baselineReport := filepath.Join(tmpDir, fmt.Sprintf("wuji-prompt-baseline-%d.json", time.Now().UnixNano()))
	candidateReport := filepath.Join(tmpDir, fmt.Sprintf("wuji-prompt-candidate-%d.json", time.Now().UnixNano()+1))
	defer os.Remove(baselineReport)
	defer os.Remove(candidateReport)
	if code := promptEvalCommand([]string{"--candidate", baselinePath, "--dataset", datasetPath, "--report", baselineReport}); code != 0 {
		return code
	}
	if code := promptEvalCommand([]string{"--candidate", candidatePath, "--dataset", datasetPath, "--report", candidateReport}); code != 0 {
		return code
	}
	baselineEval, err := loadJSONObject(baselineReport)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	candidateEval, err := loadJSONObject(candidateReport)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	baselineScore, _ := baselineEval["composite_score"].(float64)
	candidateScore, _ := candidateEval["composite_score"].(float64)
	improvement := candidateScore - baselineScore
	decision := "defer"
	reason := "needs stronger measured improvement"
	failures := []string{}
	if improvement >= 0.05 {
		decision = "absorb"
		reason = "candidate improves composite prompt score with cache-friendly structure"
	} else if improvement < 0 {
		decision = "reject"
		reason = "candidate regresses composite prompt score"
		failures = append(failures, "candidate_regressed")
	}
	report := jsonObject{
		"baseline":         absClean(baselinePath),
		"candidate":        absClean(candidatePath),
		"dataset":          absClean(datasetPath),
		"baseline_score":   baselineScore,
		"candidate_score":  candidateScore,
		"improvement":      improvement,
		"decision":         decision,
		"reason":           reason,
		"baseline_report":  baselineEval,
		"candidate_report": candidateEval,
	}
	if hasReport {
		if err := writeJSON(reportPath, report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return printGate("prompt-distill", failures)
}

func routeTaskCommand(args []string) int {
	configPath, ok := argValue(args, "--config")
	if !ok {
		usage()
		return 2
	}
	query, ok := argValue(args, "--query")
	if !ok {
		usage()
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	rawConfig, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	rawConfig = bytes.TrimPrefix(rawConfig, []byte{0xef, 0xbb, 0xbf})
	var config map[string]any
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	rawRules, ok := objectSlice(config, "routing_rules")
	if !ok || len(rawRules) == 0 {
		fmt.Fprintln(os.Stderr, "config missing routing_rules")
		return 2
	}
	modelProfiles, _ := objectMap(config, "model_profiles")
	defaultTier := "standard"
	if text := objectString(config, "default_model_tier"); text != "" {
		defaultTier = text
	}
	lowerQuery := strings.ToLower(query)
	bestScore := -1
	bestPriority := -1
	bestRule := jsonObject{}
	matches := []string{}
	for _, rawRule := range rawRules {
		rule, ok := rawRule.(map[string]any)
		if !ok {
			continue
		}
		keywords := stringSlice(rule, "keywords")
		score := 0
		ruleMatches := []string{}
		for _, keyword := range keywords {
			if keyword != "" && strings.Contains(lowerQuery, strings.ToLower(keyword)) {
				score++
				ruleMatches = append(ruleMatches, keyword)
			}
		}
		priority, _ := intFromAny(rule["priority"])
		if score > bestScore || (score == bestScore && priority > bestPriority) {
			bestScore = score
			bestPriority = priority
			bestRule = jsonObject{
				"id":          objectString(rule, "id"),
				"name":        objectString(rule, "name"),
				"provider_id": objectString(rule, "provider_id"),
				"model":       objectString(rule, "model"),
				"priority":    priority,
			}
			matches = ruleMatches
		}
	}
	if bestScore < 0 {
		fmt.Fprintln(os.Stderr, "no routing rule found")
		return 1
	}
	complexityTier := defaultTier
	reasoningEffort := "low"
	tierReason := "default_low_cost_route"
	if bestScore >= 6 || bestPriority >= 90 {
		complexityTier = "high"
		reasoningEffort = "high"
		tierReason = "dense_or_high_priority_task"
	} else if bestScore >= 3 || bestPriority >= 80 {
		complexityTier = "standard"
		reasoningEffort = "medium"
		tierReason = "multi_signal_task"
	} else {
		complexityTier = "low"
		reasoningEffort = "low"
		tierReason = "simple_or_low_risk_task"
	}
	selectedProfile := jsonObject{}
	if profileRaw, ok := modelProfiles[complexityTier]; ok {
		if profile, ok := profileRaw.(map[string]any); ok {
			selectedProfile = jsonObject(profile)
			if profileEffort := objectString(profile, "reasoning_effort"); profileEffort != "" {
				reasoningEffort = profileEffort
			}
		}
	}
	report := jsonObject{
		"query":               query,
		"matched_route":       bestRule,
		"matched_count":       bestScore,
		"matched_terms":       matches,
		"recommended_tier":    complexityTier,
		"recommended_profile": selectedProfile,
		"reasoning_effort":    reasoningEffort,
		"tier_reason":         tierReason,
	}
	if hasReport {
		if err := writeJSON(reportPath, report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	fmt.Printf("GO route-task\n- route=%s\n- tier=%s\n- reasoning=%s\n", bestRule["id"], complexityTier, reasoningEffort)
	return 0
}

func contextPackCommand(args []string) int {
	configPath, ok := argValue(args, "--config")
	if !ok {
		usage()
		return 2
	}
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	query, ok := argValue(args, "--query")
	if !ok {
		usage()
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	artifacts := argValues(args, "--artifact")
	rawConfig, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	rawConfig = bytes.TrimPrefix(rawConfig, []byte{0xef, 0xbb, 0xbf})
	var config map[string]any
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	tmpReport := filepath.Join(os.TempDir(), fmt.Sprintf("wuji-route-%d.json", time.Now().UnixNano()))
	defer os.Remove(tmpReport)
	routeCode := routeTaskCommand([]string{"--config", configPath, "--query", query, "--report", tmpReport})
	if routeCode != 0 {
		return routeCode
	}
	rawRoute, err := os.ReadFile(tmpReport)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	var routeReport map[string]any
	if err := json.Unmarshal(rawRoute, &routeReport); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	routeInfo, _ := objectMap(routeReport, "matched_route")
	cacheConfig, _ := objectMap(config, "cache_config")
	stablePrefix := jsonObject{
		"iron_rules_version": objectString(config, "iron_rules_version"),
		"route_id":           objectString(routeInfo, "id"),
		"provider_id":        objectString(routeInfo, "provider_id"),
		"model_tier":         objectString(routeReport, "recommended_tier"),
		"reasoning_effort":   objectString(routeReport, "reasoning_effort"),
		"target_hit_rate":    cacheConfig["target_hit_rate"],
		"flatten_threshold":  cacheConfig["flatten_threshold"],
	}
	dynamicContext := jsonObject{
		"query":     query,
		"artifacts": artifacts,
		"workspace": absClean(workspace),
	}
	stableBytes, _ := json.Marshal(stablePrefix)
	cacheHash := sha256.Sum256(stableBytes)
	contextPack := jsonObject{
		"stable_prefix":   stablePrefix,
		"dynamic_context": dynamicContext,
		"cache_key":       hex.EncodeToString(cacheHash[:]),
		"cache_strategy":  "stable-prefix-first",
		"route_report":    routeReport,
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(workspace, "context-pack.json")
	}
	if err := writeJSON(outputPath, contextPack); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("GO context-pack\n- report=%s\n", outputPath)
	return 0
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	args := os.Args[2:]
	code := 2
	switch os.Args[1] {
	case "reference-guard":
		code = referenceGuard(args)
	case "workflow-guard":
		code = workflowGuard(args)
	case "claim-guard":
		code = claimGuard(args)
	case "time-guard":
		code = timeGuard(args)
	case "task":
		code = taskCommand(args)
	case "sync":
		code = syncCommand(args)
	case "audit":
		code = auditCommand(args)
	case "bench":
		code = benchCommand(args)
	case "bench-report":
		code = benchReportCommand(args)
	case "preview":
		code = previewCommand(args)
	case "asset-map":
		code = assetMapCommand(args)
	case "pptx-audit":
		code = pptxAuditCommand(args)
	case "pptx-preflight":
		code = pptxPreflight(args)
	case "pptx-batch-gate":
		code = pptxBatchGate(args)
	case "mcp-guard":
		code = mcpGuard(args)
	case "mcp-distill":
		code = mcpDistill(args)
	case "route-task":
		code = routeTaskCommand(args)
	case "context-pack":
		code = contextPackCommand(args)
	case "feedback-log":
		code = feedbackLogCommand(args)
	case "feedback-dataset":
		code = feedbackDatasetCommand(args)
	case "prompt-candidate-audit":
		code = promptCandidateAudit(args)
	case "prompt-eval":
		code = promptEvalCommand(args)
	case "prompt-distill":
		code = promptDistillCommand(args)
	default:
		usage()
		code = 2
	}
	os.Exit(code)
}
