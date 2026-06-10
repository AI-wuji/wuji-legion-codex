package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type jsonObject map[string]any

type slideSummary struct {
	Name            string   `json:"name"`
	TextCount       int      `json:"text_count"`
	PicCount        int      `json:"pic_count"`
	ShapeCount      int      `json:"shape_count"`
	TextChars       int      `json:"text_chars"`
	PlaceholderHits []string `json:"placeholder_hits,omitempty"`
	TextSnippets    []string `json:"-"`
}

type pptxSummary struct {
	PPTXPath string         `json:"pptx_path"`
	Slides   []slideSummary `json:"slides"`
	Media    []string       `json:"media"`
	Layouts  []string       `json:"layouts"`
	Themes   []string       `json:"themes"`
}

type previewStats struct {
	SampleCount int     `json:"sample_count"`
	AverageLuma float64 `json:"average_luma"`
	StdDevLuma  float64 `json:"stddev_luma"`
	LightRatio  float64 `json:"light_ratio"`
	DarkRatio   float64 `json:"dark_ratio"`
}

type modelProfile struct {
	ProviderID      string `json:"provider_id"`
	Model           string `json:"model"`
	ReasoningEffort string `json:"reasoning_effort"`
}

type routeRule struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Keywords   []string `json:"keywords"`
	ProviderID string   `json:"provider_id"`
	Model      string   `json:"model"`
	Priority   int      `json:"priority"`
}

type pluginBinding struct {
	Plugin  string   `json:"plugin"`
	Owners  []string `json:"owners"`
	Purpose string   `json:"purpose"`
}

type canonDecision struct {
	Scope    string `json:"scope"`
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type distilledAtomDef struct {
	Name      string
	Residency string
	Owner     string
}

const expectedDistilledAtomCount = 21

var distilledAtomRegistry = []distilledAtomDef{
	{Name: "assumption-ledger", Residency: "resident-light", Owner: "white-hat+quality-inspection+audit"},
	{Name: "claim-fact-check", Residency: "resident-light", Owner: "white-hat+audit+intelligence-profile"},
	{Name: "reversible-evidence-handle", Residency: "resident-light", Owner: "go-execution-base+audit+quality-inspection"},
	{Name: "content-type-compression-router", Residency: "resident-light", Owner: "go-execution-base+performance-benchmark-on-demand"},
	{Name: "version-doc-mcp", Residency: "on-demand", Owner: "development-profile+intelligence-profile+guard-office"},
	{Name: "guarded-realtime-source-search", Residency: "on-demand", Owner: "intelligence-profile+guard-office+audit"},
	{Name: "research-evidence-pack", Residency: "on-demand", Owner: "intelligence-profile+audit+content-profile"},
	{Name: "skill-stocktake-daily-library", Residency: "on-demand", Owner: "evolution-profile+audit+white-hat"},
	{Name: "verified-learning-loop", Residency: "on-demand", Owner: "evolution-profile+quality-inspection+go-execution-base+performance-benchmark-on-demand"},
	{Name: "disciplined-debug-loop", Residency: "on-demand", Owner: "development-profile+quality-inspection"},
	{Name: "prior-art-solution-search", Residency: "on-demand", Owner: "intelligence-profile+owner-profile+guard-office"},
	{Name: "root-cause-radar", Residency: "on-demand", Owner: "root-cause-officer+development-profile+quality-inspection+white-hat"},
	{Name: "parallel-hypothesis-fanout", Residency: "on-demand", Owner: "staff-runtime+quality-inspection"},
	{Name: "patch-debt-root-cure", Residency: "on-demand", Owner: "evolution-profile+root-cause-officer+audit+performance-benchmark-on-demand"},
	{Name: "terminal-real-run-verification", Residency: "on-demand", Owner: "quality-inspection+audit+go-execution-base"},
	{Name: "html-native-design-canvas", Residency: "on-demand", Owner: "visual-profile+quality-inspection"},
	{Name: "brand-asset-protocol", Residency: "on-demand", Owner: "visual-profile+intelligence-profile+guard-office"},
	{Name: "anti-ai-slop-visual-rules", Residency: "on-demand", Owner: "visual-profile+quality-inspection+white-hat"},
	{Name: "design-direction-triad", Residency: "on-demand", Owner: "staff-runtime+visual-profile+nuwa-preflight"},
	{Name: "html-deck-to-editable-pptx", Residency: "on-demand", Owner: "visual-profile+go-execution-base+quality-inspection"},
	{Name: "motion-stage-sprite-engine", Residency: "on-demand", Owner: "visual-profile+performance-benchmark-on-demand"},
}

var slideTextPattern = regexp.MustCompile(`(?is)<a:t[^>]*>(.*?)</a:t>`)
var unfinishedLinePattern = regexp.MustCompile(`(?m)(^\s*(?:[-*]\s*)?(?:` + "待" + `开发|` + "后续" + `路线)(?:\s|$))|((?:` + "待" + `开发|` + "后续" + `路线)\s*[:：])`)
var incompleteLinePattern = regexp.MustCompile(`(?im)(^\s*(?:[-*]|\[\s?\])\s*(?:todo|tbd)\b)|(^\s*(?:#|//|/\*|\*)\s*(?:todo|tbd)\b)|(\b(?:todo|tbd)\s*[:：])`)

var executionPrecheckMarkers = []string{
	"先看看能不能做",
	"先确认能不能做",
	"先查环境",
	"先跑环境",
	"先试工具",
	"先试接口",
	"先探接口",
	"先扫一遍",
	"先全仓扫描",
	"先读 skill",
	"先读规则",
	"先检查一下",
	"先预检",
	"开工前预检",
	"check environment first",
	"check the environment first",
	"test the tool first",
	"probe the tool first",
	"preflight first",
	"scan the repo first",
	"inspect the repo first",
}

var executionExploratoryPhases = map[string]bool{
	"explore":   true,
	"research":  true,
	"probe":     true,
	"prototype": true,
	"preflight": true,
}

var closeoutLeakMarkers = []string{
	"下" + "一步",
	"还有可优化",
	"还可以优化",
	"继续优化",
	"要不要" + "继续",
	"是否" + "继续",
	"next step",
	"could continue",
	"can continue",
	"more to optimize",
	"further optimization",
}

var managementCeremonyMarkers = []string{
	"参谋" + "本部已接管",
	"第一阶段",
	"第二阶段",
	"第三阶段",
	"阶段一",
	"阶段二",
	"阶段三",
	"分五个阶段",
	"多角色协作",
	"多角色会诊",
	"角色分工如下",
	"committee",
	"phase 1",
	"phase 2",
	"phase 3",
	"five phases",
	"multi-agent coordination",
	"role breakdown",
}

var managementPauseMarkers = []string{
	"等你继续",
	"等你确认",
	"回复继续",
	"你回复继续后",
	"确认后继续",
	"收到继续后",
	"wait for your confirmation",
	"wait for user confirmation",
	"wait for confirmation",
	"reply continue",
	"continue to proceed",
	"after your confirmation",
	"continue to the next phase",
	"next phase after confirmation",
}

var uncertaintyMarkers = []string{
	"推测",
	"猜测",
	"未验证",
	"待查",
	"不确定",
	"可能",
	"大概",
	"maybe",
	"might",
	"possibly",
	"unverified",
	"uncertain",
	"needs verification",
}

func auditMarkerText(text string) string {
	lines := strings.Split(text, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lowerTrimmed := strings.ToLower(trimmed)
		hasMarkerToken := strings.Contains(lowerTrimmed, "todo") ||
			strings.Contains(lowerTrimmed, "tbd") ||
			strings.Contains(trimmed, "待"+"开发") ||
			strings.Contains(trimmed, "后续"+"路线") ||
			strings.Contains(trimmed, "A/B") ||
			strings.Contains(lowerTrimmed, "a/b")
		if unfinishedLinePattern.MatchString(trimmed) ||
			incompleteLinePattern.MatchString(trimmed) ||
			(hasMarkerToken && (strings.HasPrefix(trimmed, "#") ||
				strings.HasPrefix(trimmed, "//") ||
				strings.HasPrefix(trimmed, "/*") ||
				strings.HasPrefix(trimmed, "*") ||
				strings.Contains(trimmed, "A/B:") ||
				strings.Contains(lowerTrimmed, "a/b:"))) {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

func isAuditMarkerFixture(rel string) bool {
	cleanRel := filepath.ToSlash(rel)
	return cleanRel == "tools/wuji_cli.go" || cleanRel == "scripts/test-wuji-cli.ps1"
}

func normalizedLower(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isExecutionMetaArtifact(path string) bool {
	base := normalizedLower(filepath.Base(path))
	if base == "" {
		return true
	}
	metaHints := []string{
		"audit", "report", "scan", "inspect", "preflight", "check", "log", "manifest",
		"sarif", "evidence", "bench", "context-pack", "route-report", "canon-report",
	}
	for _, hint := range metaHints {
		if strings.Contains(base, hint) {
			return true
		}
	}
	return false
}

func hasPrimaryArtifact(artifacts []string) bool {
	for _, artifact := range artifacts {
		if nonEmpty(artifact) && !isExecutionMetaArtifact(artifact) {
			return true
		}
	}
	return false
}

func executionRhythmFailures(record map[string]any) []string {
	event := normalizedLower(objectString(record, "event"))
	if event != "start" && event != "heartbeat" {
		return nil
	}
	phase := normalizedLower(objectString(record, "phase"))
	note := objectString(record, "note")
	hits := markerHits(strings.Join([]string{phase, note}, "\n"), executionPrecheckMarkers)
	if note == "" {
		hits = stringSlice(record, "execution_precheck_hits")
	}
	exploratory := executionExploratoryPhases[phase]
	artifacts := stringSlice(record, "artifacts")
	primaryArtifactPresent := hasPrimaryArtifact(artifacts)
	if len(artifacts) == 0 {
		if value, ok := objectBool(record, "primary_artifact_present"); ok {
			primaryArtifactPresent = value
		}
	}
	if (exploratory || len(hits) > 0) && !primaryArtifactPresent {
		parts := []string{}
		if phase != "" {
			parts = append(parts, "phase="+phase)
		}
		if len(hits) > 0 {
			parts = append(parts, "markers="+strings.Join(hits, "|"))
		}
		if len(artifacts) > 0 {
			parts = append(parts, "artifact_keys="+privacyHash(strings.Join(artifacts, "|")))
		} else {
			parts = append(parts, "artifacts=none")
		}
		return []string{"execution_precheck_loop_detected " + strings.Join(parts, " ")}
	}
	return nil
}

func taskLogExecutionRhythmFailures(records []map[string]any) []string {
	failures := []string{}
	for idx, record := range records {
		for _, failure := range executionRhythmFailures(record) {
			failures = append(failures, fmt.Sprintf("task_log_record_%02d_%s", idx+1, strings.ReplaceAll(failure, " ", "_")))
		}
	}
	return failures
}

func taskLogCloseoutLeakFailures(records []map[string]any) []string {
	failures := []string{}
	for idx, record := range records {
		event := normalizedLower(objectString(record, "event"))
		status := normalizedLower(objectString(record, "status"))
		note := objectString(record, "note")
		if event == "end" && status == "done" {
			hits := markerHits(note, closeoutLeakMarkers)
			if note == "" {
				hits = stringSlice(record, "closeout_leak_hits")
			}
			if len(hits) > 0 {
				failures = append(failures, fmt.Sprintf("task_log_record_%02d_done_note_reopens_work=%s", idx+1, strings.Join(hits, "|")))
			}
		}
	}
	return failures
}

func taskLogBlockedWaitFailures(records []map[string]any) []string {
	failures := []string{}
	for idx, record := range records {
		event := normalizedLower(objectString(record, "event"))
		status := normalizedLower(objectString(record, "status"))
		note := objectString(record, "note")
		hits := markerHits(note, managementPauseMarkers)
		if note == "" {
			hits = stringSlice(record, "management_pause_hits")
		}
		if (event == "blocked" || event == "heartbeat" || status == "blocked" || status == "needs_decision" || status == "running") && len(hits) > 0 {
			failures = append(failures, fmt.Sprintf("task_log_record_%02d_note_waits_for_continue=%s", idx+1, strings.Join(hits, "|")))
		}
	}
	return failures
}

var placeholderMarkers = []string{
	"输入标题内容",
	"输入小标题",
	"输入内容",
	"点击添加",
	"单击此处添加",
	"单击此处输入",
	"click to add",
	"click to add title",
	"click to add subtitle",
	"click to add text",
	"tap to add",
	"lorem ipsum",
	"keyword",
	"placeholder",
	"在此处添加",
}

var teachingMarkers = []string{
	"教程",
	"教学",
	"步骤",
	"操作",
	"界面",
	"按钮",
	"点击",
	"导入",
	"导出",
	"剪辑",
	"时间线",
	"字幕",
	"调色",
	"音频",
	"设置",
	"新建",
	"软件",
	"手机",
	"电脑",
	"流程",
	"演示",
	"使用",
	"回顾",
	"打开",
}

var pilotApprovalPositiveMarkers = []string{
	"approved",
	"approve",
	"pass",
	"同意",
	"通过",
	"可批量",
	"可以批量",
	"允许批量",
}

var pilotApprovalNegativeMarkers = []string{
	"disapprove",
	"disapproved",
	"reject",
	"rejected",
	"no-go",
	"blocked",
	"rework",
	"fail",
	"不同意",
	"不通过",
	"驳回",
	"退回",
	"重做",
}

const builtinIronRulesVersion = "11.3"
const builtinDefaultModelTier = "low"
const maxOptimizationContextPackBytes int64 = 12 * 1024
const maxOptimizationStablePrefixFields = 16
const maxOptimizationOutputsBytes int64 = 50 * 1024 * 1024
const maxOptimizationOutputsFiles = 500
const maxOptimizationToolsBytes int64 = 300 * 1024 * 1024
const maxOptimizationToolsFiles = 16000
const maxHotpathResidentBytes int64 = 8 * 1024
const maxHotpathDynamicBytes int64 = 32 * 1024
const maxCachedPrefixBytesP95 int64 = 32 * 1024
const maxInputTokensP95 = 32000
const maxFreshInputTokensP95 = 12000
const maxOutputTokensP95 = 4000
const maxUncachedTokensP95 = 14000
const maxActivatedOfficers = 3
const maxActivatedSkills = 1
const maxLoadedFileBytes int64 = 32 * 1024
const maxLargestContextSegmentBytes int64 = 32 * 1024
const minMeasuredCacheObservations = 2
const maxRuntimeSkillBytes int64 = 8 * 1024
const maxRuntimeRepoInstructionBytes int64 = 2 * 1024
const maxRuntimeMirrorBytes int64 = 24 * 1024
const minRuntimeUsageObservations = 2

var successClaimMarkers = []string{
	"\u5b8c\u6210", "\u6210\u529f", "\u901a\u8fc7", "\u5df2\u878d\u5408", "\u5df2\u751f\u6210", "\u5df2\u4fee\u590d", "\u5df2\u5b9e\u73b0",
	"verified", "passed", "complete", "completed", "success", "fixed", "resolved", "implemented",
}

var builtinTopLevelRoles = []string{
	"阿极",
	"参谋运行时",
	"女娲",
	"白帽",
	"保卫科",
	"root-cause-officer",
	"审计",
	"质检",
	"进化主帅",
}

var builtinModelProfiles = map[string]modelProfile{
	"low": {
		ProviderID:      "deepseek-web",
		Model:           "deepseek-chat",
		ReasoningEffort: "low",
	},
	"standard": {
		ProviderID:      "deepseek-web",
		Model:           "deepseek-chat",
		ReasoningEffort: "medium",
	},
	"high": {
		ProviderID:      "deepseek-web",
		Model:           "deepseek-chat",
		ReasoningEffort: "high",
	},
}

func isQualityInspectionRoute(routeID string) bool {
	return strings.EqualFold(routeID, "quality-inspection") || strings.EqualFold(routeID, "qa")
}

var builtinPluginBindings = []pluginBinding{
	{
		Plugin:  "Browser",
		Owners:  []string{"视觉主帅", "开发主帅", "情报主帅", "质检官"},
		Purpose: "网页打开、检查、交互测试、截图",
	},
	{
		Plugin:  "Documents",
		Owners:  []string{"内容主帅", "情报主帅", "审计官"},
		Purpose: "Word/文档生成、整理、归档",
	},
	{
		Plugin:  "Spreadsheets",
		Owners:  []string{"数据主帅", "情报主帅", "内容主帅", "性能基准官"},
		Purpose: "表格、结构化数据、分析交付",
	},
	{
		Plugin:  "Presentations",
		Owners:  []string{"视觉主帅", "质检官"},
		Purpose: "PPTX 生成、修改、导出预览",
	},
}

var builtinMCPPolicies = []canonDecision{
	{
		Scope:    "local-low-permission",
		Decision: "absorb",
		Reason:   "本地、低权限、边界清楚的工具可以直接进入候选层",
	},
	{
		Scope:    "network-or-account",
		Decision: "defer",
		Reason:   "联网、账号、OAuth、外部写操作需要任务级授权",
	},
	{
		Scope:    "high-risk-or-secrets",
		Decision: "reject/defer",
		Reason:   "高权限、带 secrets、来源或许可证不清的工具不进默认主链",
	},
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  wuji-cli reference-guard --reference <file>... --output <path>...")
	fmt.Fprintln(os.Stderr, "  wuji-cli claim-guard --claim <text> [--workspace <dir>] [--evidence <file>]...")
	fmt.Fprintln(os.Stderr, "  wuji-cli time-guard --kind <non-code|general> --elapsed-minutes <n> [--artifact <file>] [--phase <name>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli audit --path <dir> [--report <file>] [--sarif <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli code-map --workspace <dir> --goal <text> --entry <text> [--dependency <text>]... [--risk <text>]... [--verify <text>]... [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli root-cause-radar --workspace <dir> --symptom <text> --repro <text> --hypothesis <text>... --root-cause <text> --same-class-scan <text> --same-class-evidence <file>... --fix-strategy <text> --regression-evidence <file>... [--eliminated-cause <text>]... [--patch-debt-action <text>] [--artifact <file>]... [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli bugfix-guard --workspace <dir> --goal <text> --repro <text> --root-cause-report <file> [--artifact <file>]... [--verify <text>]... [--self-test <file>]... [--independent-check <file>]... [--browser-check <file>]... [--still-failing <text>]... [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli quality-guard --workspace <dir> --goal <text> [--artifact <file>]... [--verify <text>]... [--browser-check <file>]... [--program-check <file>]... [--command-check <file>]... [--mcp-check <file>]... [--still-failing <text>]... [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli migration-guard --workspace <dir> --goal <text> --feature-map <file> [--artifact <file>]... [--verify <text>]... [--run-evidence <file>]... [--preview-evidence <file>]... [--missing-feature <text>]... [--fake-page <text>]... [--placeholder-page <text>]... [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli closeout-check --workspace <dir> --goal <text> [--audit-workspace <dir>] [--root-cause-required true|false] [--root-cause-report <file>] [--artifact <file>]... [--verify <text>]... [--next-gap <text>]... [--needs-user-decision true|false] [--blocked-reason <text>] [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli finish-or-block [--workspace <dir>] --goal <text> [--audit-workspace <dir>] [--root-cause-required true|false] [--root-cause-report <file>] [--remaining-step <text>]... [--needs-user-decision true|false] [--blocked-reason <text>] [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli evidence-grade [--workspace <dir>] --status <candidate|checked|verified|shipped> --summary <text> [--artifact <file>]... [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli truth-state --text <text> --state <fact|inference|todo> [--evidence <file>]... [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli asset-map --pptx <file> --workspace <dir>")
	fmt.Fprintln(os.Stderr, "  wuji-cli pptx-audit --pptx <file> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli pptx-preflight --workspace <dir> [--generator <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli pptx-batch-gate --workspace <dir> [--generator <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli mcp-guard --manifest <file> [--workspace <dir>] [--allow-network true|false]")
	fmt.Fprintln(os.Stderr, "  wuji-cli supply-chain --manifest <file> [--workspace <dir>] [--allow-network true|false]")
	fmt.Fprintln(os.Stderr, "  wuji-cli fusion-audit --workspace <dir> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli optimization-audit --workspace <dir> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli context-bloat-audit --workspace <dir> [--bench-report <file>] [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli runtime-context-audit --workspace <dir> [--usage-log <file>] [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli bench --workspace <measured-dir> --name <run> [--log-dir <dir>] [--input-tokens <n>] [--output-tokens <n>] [--cached-tokens <n>] [--fresh-input-tokens <n>] [--duration-ms <n>] [--tool-calls <n>] [--retries <n>] [--quality-pass <true|false>] [--cache-hit <true|false>] [--reused-prefix-bytes <n>] [--activated-officers <n>] [--activated-skills <n>] [--loaded-file-bytes <n>] [--largest-context-segment-bytes <n>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli bench-report --workspace <measured-dir> [--log-dir <dir>] [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli help all")
}

func usageAll() {
	usage()
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Specialized / offline commands:")
	fmt.Fprintln(os.Stderr, "  wuji-cli workflow-guard --workspace <dir> [--stage scaffold|final]")
	fmt.Fprintln(os.Stderr, "  wuji-cli task --workspace <dir> --event <start|heartbeat|blocked|end> [--status <running|blocked|needs_decision|done>] [--artifact <file>]... [--note <text>] [--phase <name>] [--audit-workspace <dir>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli sync --source <dir> --dest <dir>")
	fmt.Fprintln(os.Stderr, "  wuji-cli bench --workspace <measured-dir> --name <run> [--log-dir <dir>] [--input-tokens <n>] [--output-tokens <n>] [--cached-tokens <n>] [--fresh-input-tokens <n>] [--duration-ms <n>] [--tool-calls <n>] [--retries <n>] [--quality-pass <true|false>] [--cache-hit <true|false>] [--reused-prefix-bytes <n>] [--activated-officers <n>] [--activated-skills <n>] [--loaded-file-bytes <n>] [--largest-context-segment-bytes <n>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli bench-report --workspace <measured-dir> [--log-dir <dir>] [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli repeat-candidates --log <file> [--min-occurrences <n>] [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli preview --command <exe> [--arg <value>]... --output <file>")
	fmt.Fprintln(os.Stderr, "  wuji-cli ppt-template-inspect --workspace <dir> --pptx <file> [--out-dir <dir>] [--scale <n>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli ppt-template-starter --workspace <dir> --pptx <file> --map <file> --out <file> [--preview-dir <dir>] [--layout-dir <dir>] [--inspect <file>] [--contact-sheet <file>] [--scale <n>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli ppt-template-edit --workspace <dir> --starter-pptx <file> --map <file> --out <file> [--preview-dir <dir>] [--layout-dir <dir>] [--report <file>] [--scale <n>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli ppt-template-fidelity --workspace <dir> --final-pptx <file> [--map <file>] [--starter-pptx <file>] [--starter-layout-dir <dir>] [--final-layout-dir <dir>] [--edit-dir <dir>] [--agent-log <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli ppt-htmlfirst --workspace <dir> --html <file> --out <file> [--title <text>] [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli ppt-com-refine --pptx <file> --out <file> [--instructions <file>] [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli ppt-pipeline --workspace <dir> --route <html-first|template-following> --out <file> [--html <file>] [--pptx <file>] [--map <file>] [--report <file>] [--auto-approve true|false] [--pilot-approval <file>] [--com-refine true|false] [--refine-instructions <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli mcp-distill --catalog <file> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli canon-report [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli fusion-audit --workspace <dir> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli optimization-audit --workspace <dir> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli context-bloat-audit --workspace <dir> [--bench-report <file>] [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli runtime-context-audit --workspace <dir> [--usage-log <file>] [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli route-task --config <file> --query <text> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli context-pack --config <file> --workspace <dir> --query <text> [--artifact <file>]... [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli feedback-log --workspace <dir> --task <text> [--prefer <term>]... [--avoid <term>]... [--note <text>] [--source <user|quality_inspection|audit>] [--report <file>]")
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

func pathPrivacyRef(workspace string, path string) string {
	cleanPath := absClean(path)
	if strings.TrimSpace(workspace) != "" && sameOrDescendant(cleanPath, workspace) {
		rel, err := filepath.Rel(absClean(workspace), cleanPath)
		if err == nil {
			rel = filepath.ToSlash(rel)
			if rel == "." {
				return "."
			}
			return rel
		}
	}
	return "external:" + privacyHash(cleanPath)
}

func nonEmpty(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Size() >= 20
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func directoryStats(root string) (int, int64, error) {
	fileCount := 0
	var totalBytes int64
	if !dirExists(root) {
		return 0, 0, nil
	}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		fileCount++
		totalBytes += info.Size()
		return nil
	})
	return fileCount, totalBytes, err
}

func workspaceFileInventory(root string) ([]string, error) {
	files := []string{}
	root = absClean(root)
	skipDirs := map[string]bool{
		".git":          true,
		".wuji-tools":   true,
		"output":        true,
		"outputs":       true,
		"__pycache__":   true,
		".wuji-errors":  true,
		".wuji-backups": true,
		"node_modules":  true,
	}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name := entry.Name()
		if entry.IsDir() {
			if path != root && skipDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		lower := strings.ToLower(rel)
		if strings.HasSuffix(lower, ".tmp") || strings.HasSuffix(lower, ".log") || strings.HasSuffix(lower, ".pyc") {
			return nil
		}
		files = append(files, rel)
		return nil
	})
	sort.Strings(files)
	return files, err
}

func workspaceEntryPattern(workspace string, entryPath string) (string, bool) {
	entryPath = strings.TrimSpace(entryPath)
	if entryPath == "" {
		return "", false
	}
	workspaceSlash := strings.TrimRight(filepath.ToSlash(absClean(workspace)), "/")
	entrySlash := filepath.ToSlash(entryPath)
	entrySlash = strings.TrimPrefix(entrySlash, "./")
	entryLower := strings.ToLower(entrySlash)
	workspaceLower := strings.ToLower(workspaceSlash)
	if strings.HasPrefix(entryLower, workspaceLower+"/") {
		return strings.TrimPrefix(entrySlash, workspaceSlash+"/"), true
	}
	if strings.EqualFold(entrySlash, workspaceSlash) {
		return "", false
	}
	if filepath.IsAbs(entryPath) || strings.Contains(entrySlash, ":/") {
		return "", false
	}
	return entrySlash, true
}

func slashPatternMatch(pattern string, rel string) bool {
	pattern = filepath.ToSlash(strings.TrimSpace(pattern))
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	if pattern == rel {
		return true
	}
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(rel, pattern)
	}
	if strings.ContainsAny(pattern, "*?[") {
		ok, err := filepath.Match(pattern, rel)
		return err == nil && ok
	}
	return false
}

func isRetiredResidualStatus(status string) bool {
	switch status {
	case "retired-deleted", "retire-and-label", "delete-now", "delete-now-except-latest-evidence", "delete-when-not-building", "excluded":
		return true
	default:
		return false
	}
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return 0
	}
	return info.Size()
}

func repoRootCandidates() []string {
	seen := map[string]bool{}
	candidates := []string{}
	addWithParents := func(start string) {
		if strings.TrimSpace(start) == "" {
			return
		}
		dir := absClean(start)
		for {
			if !seen[dir] {
				seen[dir] = true
				candidates = append(candidates, dir)
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		addWithParents(cwd)
	}
	if exe, err := os.Executable(); err == nil {
		addWithParents(filepath.Dir(exe))
	}
	return candidates
}

func resolveRepoScript(scriptName string) (string, error) {
	for _, root := range repoRootCandidates() {
		scriptPath := filepath.Join(root, "scripts", scriptName)
		if fileExists(scriptPath) && fileExists(filepath.Join(root, "tools", "wuji_cli.go")) {
			return scriptPath, nil
		}
	}
	return "", fmt.Errorf("wuji script not found: %s", scriptName)
}

func toPowerShellArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		if !strings.HasPrefix(arg, "--") {
			out = append(out, arg)
			continue
		}
		parts := strings.Split(strings.TrimPrefix(arg, "--"), "-")
		for index, part := range parts {
			if part == "" {
				continue
			}
			parts[index] = strings.ToUpper(part[:1]) + part[1:]
		}
		out = append(out, "-"+strings.Join(parts, ""))
	}
	return out
}

func runPowerShellScript(scriptName string, args []string) error {
	scriptPath, err := resolveRepoScript(scriptName)
	if err != nil {
		return err
	}
	commandArgs := append([]string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath}, toPowerShellArgs(args)...)
	cmd := exec.Command("powershell.exe", commandArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runPowerShellScriptCommand(scriptName string, args []string) int {
	if err := runPowerShellScript(scriptName, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
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

func auditInputHashes(workspace string, relPaths []string) jsonObject {
	hashes := jsonObject{}
	for _, rel := range relPaths {
		rel = filepath.ToSlash(strings.TrimSpace(rel))
		if rel == "" {
			continue
		}
		path := filepath.Join(workspace, filepath.FromSlash(rel))
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			hashes[rel] = "__missing__"
			continue
		}
		hash, err := fileSHA256(path)
		if err != nil {
			hashes[rel] = "__unreadable__"
			continue
		}
		hashes[rel] = hash
	}
	return hashes
}

func auditExternalInputHashes(paths []string) jsonObject {
	hashes := jsonObject{}
	for _, path := range paths {
		key := privacyHash(absClean(path))
		if !fileExists(path) {
			hashes[key] = "__missing__"
			continue
		}
		hash, err := fileSHA256(path)
		if err != nil {
			hashes[key] = "__unreadable__"
			continue
		}
		hashes[key] = hash
	}
	return hashes
}

func auditManifest(workspace string, command string, relPaths []string, externalPaths ...string) jsonObject {
	toolHash := ""
	if hash, err := fileSHA256(filepath.Join(workspace, "tools", "wuji_cli.go")); err == nil {
		toolHash = hash
	}
	manifest := jsonObject{
		"generated_at":     time.Now().UTC().Format(time.RFC3339),
		"command":          command,
		"wuji_version":     builtinIronRulesVersion,
		"tool_source_hash": toolHash,
		"input_hashes":     auditInputHashes(workspace, relPaths),
	}
	if len(externalPaths) > 0 {
		manifest["external_input_hashes"] = auditExternalInputHashes(externalPaths)
	}
	return manifest
}

func auditReportFreshnessFailures(report map[string]any, workspace string, gate string) []string {
	failures := []string{}
	generatedAt := objectString(report, "generated_at")
	if generatedAt == "" {
		failures = append(failures, gate+"_report_missing_generated_at")
	} else if _, err := time.Parse(time.RFC3339, generatedAt); err != nil {
		failures = append(failures, gate+"_report_bad_generated_at")
	}
	if objectString(report, "command") != gate {
		failures = append(failures, gate+"_report_command_mismatch")
	}
	if objectString(report, "wuji_version") != builtinIronRulesVersion {
		failures = append(failures, gate+"_report_version_mismatch")
	}
	if toolHash := objectString(report, "tool_source_hash"); toolHash == "" {
		failures = append(failures, gate+"_report_missing_tool_source_hash")
	} else if current, err := fileSHA256(filepath.Join(workspace, "tools", "wuji_cli.go")); err != nil || current != toolHash {
		failures = append(failures, gate+"_report_tool_source_hash_mismatch")
	}
	inputs, ok := objectMap(report, "input_hashes")
	if !ok || len(inputs) == 0 {
		failures = append(failures, gate+"_report_missing_input_hashes")
		return failures
	}
	requiredInputs := map[string][]string{
		"fusion-audit": {
			"kernel-source.json",
			"config.json",
			"fusion-matrix.json",
			"residual-entrypoints.json",
			"acceptance-checklists.json",
			"purification-charter.json",
			"hotpath-manifest.json",
			"README.md",
			"tools/wuji_cli.go",
		},
		"optimization-audit": {
			"config.json",
			"acceptance-checklists.json",
			"outputs/context-pack-rich.json",
			"hotpath-manifest.json",
			"tools/wuji_cli.go",
		},
		"context-bloat-audit": {
			"hotpath-manifest.json",
			"outputs/context-pack-rich.json",
			"outputs/bench-report.json",
			"tools/wuji_cli.go",
		},
	}
	for _, rel := range requiredInputs[gate] {
		if _, ok := inputs[rel]; !ok {
			failures = append(failures, gate+"_report_missing_required_input_hash="+rel)
		}
	}
	for rel, rawExpected := range inputs {
		expected, ok := rawExpected.(string)
		if !ok || expected == "" || strings.HasPrefix(expected, "__") {
			failures = append(failures, gate+"_report_bad_input_hash="+rel)
			continue
		}
		current, err := fileSHA256(filepath.Join(workspace, filepath.FromSlash(rel)))
		if err != nil {
			failures = append(failures, gate+"_report_input_unreadable="+rel)
			continue
		}
		if current != expected {
			failures = append(failures, gate+"_report_input_hash_mismatch="+rel)
		}
	}
	if gate == "fusion-audit" {
		externalInputs, ok := objectMap(report, "external_input_hashes")
		if !ok || len(externalInputs) == 0 {
			failures = append(failures, gate+"_report_missing_external_input_hashes")
		} else {
			for _, path := range []string{`C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md`} {
				key := privacyHash(absClean(path))
				expected, ok := externalInputs[key].(string)
				if !ok || expected == "" || strings.HasPrefix(expected, "__") {
					failures = append(failures, gate+"_report_bad_external_input_hash="+key)
					continue
				}
				current, err := fileSHA256(path)
				if err != nil {
					failures = append(failures, gate+"_report_external_input_unreadable="+key)
					continue
				}
				if current != expected {
					failures = append(failures, gate+"_report_external_input_hash_mismatch="+key)
				}
			}
		}
	}
	return uniqueStrings(failures)
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

func writeCompactJSON(path string, value any) error {
	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}
	bytes, err := json.Marshal(value)
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

func objectBoolValue(obj map[string]any, key string) bool {
	value, ok := objectBool(obj, key)
	return ok && value
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
	if typed, ok := obj[key].([]string); ok {
		result := []string{}
		for _, value := range typed {
			if strings.TrimSpace(value) != "" {
				result = append(result, strings.TrimSpace(value))
			}
		}
		return result
	}
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
	case int64:
		return int(typed), true
	default:
		return 0, false
	}
}

func int64FromAny(value any) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), true
	case int:
		return int64(typed), true
	case int64:
		return typed, true
	default:
		return 0, false
	}
}

func intFromKeys(obj map[string]any, keys ...string) (int, bool) {
	for _, key := range keys {
		if value, ok := intFromAny(obj[key]); ok {
			return value, true
		}
	}
	return 0, false
}

func sortedIntCopy(values []int) []int {
	copied := append([]int{}, values...)
	sort.Ints(copied)
	return copied
}

func percentileInt(values []int, percentile float64) int {
	if len(values) == 0 {
		return 0
	}
	sorted := sortedIntCopy(values)
	if percentile <= 0 {
		return sorted[0]
	}
	if percentile >= 1 {
		return sorted[len(sorted)-1]
	}
	index := int(math.Ceil(percentile*float64(len(sorted)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

func maxInt(values []int) int {
	max := 0
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}

func boolFromAny(value any) (bool, bool) {
	typed, ok := value.(bool)
	return typed, ok
}

func evidenceLevelFromDecision(decision string) string {
	switch strings.ToLower(strings.TrimSpace(decision)) {
	case "absorb":
		return "verified"
	case "defer":
		return "checked"
	case "reject":
		return "checked"
	default:
		return "candidate"
	}
}

func stringFromAny(value any) (string, bool) {
	typed, ok := value.(string)
	if !ok {
		return "", false
	}
	trimmed := strings.TrimSpace(typed)
	if trimmed == "" {
		return "", false
	}
	return trimmed, true
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

func ternaryStatus(ok bool, pass string, fail string) string {
	if ok {
		return pass
	}
	return fail
}

func summarizeFailureLines(text string, limit int) []string {
	if limit <= 0 {
		limit = 8
	}
	lines := strings.Split(text, "\n")
	hits := []string{}
	markers := []string{
		"error", "failed", "failure", "exception", "panic", "traceback", "fatal",
		"warn", "warning", "blocked", "mismatch", "missing",
		"错误", "失败", "异常", "崩溃", "警告", "阻塞", "缺失", "未通过",
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if containsSecretLikeContent(trimmed) {
			continue
		}
		lower := strings.ToLower(trimmed)
		for _, marker := range markers {
			if strings.Contains(lower, marker) || strings.Contains(trimmed, marker) {
				hits = append(hits, trimmed)
				break
			}
		}
		if len(hits) >= limit {
			break
		}
	}
	return uniqueStrings(hits)
}

func summarizeKeyLines(text string, limit int) []string {
	if limit <= 0 {
		limit = 8
	}
	lines := strings.Split(text, "\n")
	snippets := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if containsSecretLikeContent(trimmed) {
			continue
		}
		snippets = append(snippets, trimmed)
		if len(snippets) >= limit {
			break
		}
	}
	return snippets
}

func orderedStablePrefix(stablePrefix jsonObject) []jsonObject {
	order := []string{
		"iron_rules_version",
		"route_id",
		"task_state",
		"owner_profile",
		"provider_id",
		"model_tier",
		"reasoning_effort",
		"target_hit_rate",
		"flatten_threshold",
		"stable_prefix_policy",
		"mount_policy",
		"tool_output_policy",
		"concise_execution_policy",
		"optimization_objective",
		"canon_source",
	}
	result := []jsonObject{}
	seen := map[string]bool{}
	for _, key := range order {
		if value, ok := stablePrefix[key]; ok {
			result = append(result, jsonObject{"key": key, "value": value})
			seen[key] = true
		}
	}
	extras := []string{}
	for key := range stablePrefix {
		if !seen[key] {
			extras = append(extras, key)
		}
	}
	sort.Strings(extras)
	for _, key := range extras {
		result = append(result, jsonObject{"key": key, "value": stablePrefix[key]})
	}
	return result
}

func stablePrefixCanon(stablePrefix jsonObject) jsonObject {
	ordered := orderedStablePrefix(stablePrefix)
	canonLines := []string{}
	seenValues := map[string]string{}
	duplicateKeys := []string{}
	for _, item := range ordered {
		key := fmt.Sprint(item["key"])
		value := normalizeSpace(fmt.Sprint(item["value"]))
		if prior, ok := seenValues[key]; ok && prior == value {
			duplicateKeys = append(duplicateKeys, key)
			continue
		}
		seenValues[key] = value
		canonLines = append(canonLines, key+"="+value)
	}
	canonText := strings.Join(canonLines, "\n")
	hash := sha256.Sum256([]byte(canonText))
	return jsonObject{
		"canon_hash":       hex.EncodeToString(hash[:]),
		"duplicate_count":  len(uniqueStrings(duplicateKeys)),
		"field_count":      len(ordered),
		"canon_line_count": len(canonLines),
	}
}

func assemblyStrategyForKind(kind string) string {
	switch kind {
	case "log":
		return "failure-lines-first"
	case "structured":
		return "schema-and-salient-fields"
	case "text":
		return "key-lines-first"
	default:
		return "handle-only"
	}
}

func artifactKind(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".log", ".txt":
		return "log"
	case ".json", ".jsonl":
		return "structured"
	case ".md", ".go", ".py", ".js", ".ts", ".tsx", ".jsx", ".ps1", ".yaml", ".yml":
		return "text"
	default:
		return "binary"
	}
}

func summarizeArtifact(workspace string, path string) jsonObject {
	pathRef := pathPrivacyRef(workspace, path)
	summary := jsonObject{
		"path_ref":        pathRef,
		"kind":            artifactKind(path),
		"evidence_handle": pathRef,
	}
	summary["assembly_strategy"] = assemblyStrategyForKind(summary["kind"].(string))
	if hash, err := fileSHA256(path); err == nil {
		summary["artifact_hash"] = hash
	}
	info, err := os.Stat(path)
	if err != nil {
		summary["status"] = "missing"
		summary["summary_mode"] = "missing"
		summary["failures"] = []string{"artifact_missing"}
		return summary
	}
	summary["bytes"] = info.Size()
	kind := summary["kind"].(string)
	if kind == "binary" {
		summary["status"] = "ok"
		summary["summary_mode"] = "handle-only"
		return summary
	}
	data, err := os.ReadFile(path)
	if err != nil {
		summary["status"] = "unreadable"
		summary["summary_mode"] = "handle-only"
		summary["failures"] = []string{"artifact_unreadable"}
		return summary
	}
	text := string(data)
	failureLines := []string{}
	if kind == "log" {
		failureLines = summarizeFailureLines(text, 8)
	}
	keyLines := summarizeKeyLines(text, 6)
	if containsSecretLikeContent(text) {
		summary["status"] = "redacted"
		summary["failures"] = []string{"secret_like_lines_redacted"}
	} else {
		summary["status"] = "ok"
	}
	if len(failureLines) > 0 {
		summary["failure_lines"] = failureLines
	}
	if len(keyLines) > 0 {
		summary["key_lines"] = keyLines
	}
	if len(failureLines) > 0 {
		summary["summary_mode"] = "failure-lines"
	} else if len(keyLines) > 0 {
		summary["summary_mode"] = "key-lines"
	} else {
		summary["summary_mode"] = "redacted-lines"
	}
	return summary
}

func summarizeArtifactSafe(workspace string, path string) jsonObject {
	pathRef := pathPrivacyRef(workspace, path)
	summary := jsonObject{
		"path_ref":          pathRef,
		"kind":              artifactKind(path),
		"evidence_handle":   pathRef,
		"assembly_strategy": "handle-only",
		"summary_mode":      "handle-only",
	}
	if privateEvidencePathDenied(path) {
		summary["status"] = "denied"
		summary["failures"] = []string{"artifact_private_denied"}
		return summary
	}
	if workspace != "" && !sameOrDescendant(path, workspace) {
		summary["status"] = "denied"
		summary["failures"] = []string{"artifact_outside_workspace_denied"}
		return summary
	}
	return summarizeArtifact(workspace, path)
}

func splitArtifactSummaries(summaries []jsonObject) ([]jsonObject, []jsonObject) {
	execution := []jsonObject{}
	audit := []jsonObject{}
	for _, item := range summaries {
		record := jsonObject{
			"path_ref":          item["path_ref"],
			"kind":              item["kind"],
			"summary_mode":      item["summary_mode"],
			"evidence_handle":   item["evidence_handle"],
			"assembly_strategy": item["assembly_strategy"],
		}
		if hash, ok := item["artifact_hash"]; ok {
			record["artifact_hash"] = hash
		}
		if failures, ok := item["failure_lines"]; ok {
			if typed, ok := failures.([]string); ok {
				record["failure_line_count"] = len(typed)
			}
		}
		if keys, ok := item["key_lines"]; ok {
			if typed, ok := keys.([]string); ok {
				record["key_line_count"] = len(typed)
			}
		}
		mode := fmt.Sprint(item["summary_mode"])
		kind := fmt.Sprint(item["kind"])
		if mode == "failure-lines" || kind == "log" {
			execution = append(execution, record)
		} else {
			audit = append(audit, record)
		}
	}
	return execution, audit
}

func reviewOptimizationAssembly(stablePrefix jsonObject, summaries []jsonObject, execution []jsonObject, audit []jsonObject) jsonObject {
	failures := []string{}
	warnings := []string{}
	if len(summaries) > 0 && len(execution) == 0 && len(audit) == 0 {
		failures = append(failures, "artifact_summaries_missing")
	}
	for _, item := range summaries {
		if fmt.Sprint(item["evidence_handle"]) == "" {
			failures = append(failures, "summary_missing_evidence_handle")
		}
		mode := fmt.Sprint(item["summary_mode"])
		kind := fmt.Sprint(item["kind"])
		if mode == "handle-only" && kind != "binary" {
			failures = append(failures, "text_artifact_not_summarized="+fmt.Sprint(item["path_ref"]))
		}
		if mode == "key-lines" {
			if failureLines, ok := item["failure_lines"].([]string); ok && len(failureLines) > 0 {
				warnings = append(warnings, "failure_lines_not_prioritized="+fmt.Sprint(item["path_ref"]))
			}
		}
	}
	if len(stablePrefix) == 0 {
		failures = append(failures, "stable_prefix_missing")
	}
	return jsonObject{
		"gate":     "anti-token-overoptimization",
		"status":   ternaryStatus(len(failures) == 0, "pass", "fail"),
		"failures": failures,
		"warnings": uniqueStrings(warnings),
		"checks": []string{
			"evidence-handle-preserved",
			"failure-lines-prioritized",
			"execution-and-audit-separated",
			"stable-prefix-kept-intact",
			"assembly-strategy-explicit",
		},
	}
}

func evolutionDistillReport(classifications []jsonObject) jsonObject {
	resident := []jsonObject{}
	mountOnDemand := []jsonObject{}
	retire := []jsonObject{}
	for _, item := range classifications {
		switch fmt.Sprint(item["classification"]) {
		case "resident":
			resident = append(resident, item)
		case "retire":
			retire = append(retire, item)
		default:
			mountOnDemand = append(mountOnDemand, item)
		}
	}
	return jsonObject{
		"resident":        resident,
		"mount_on_demand": mountOnDemand,
		"retire":          retire,
		"decision_rule":   "prefer stable high-signal repeats, defer cold utility, retire noisy negative-signal patterns",
		"evidence_policy": "classification must preserve supporting prefer/avoid counts",
	}
}

func classifyStrategyResidency(task string, occurrences int, preferCount int, avoidCount int) (string, string) {
	switch {
	case avoidCount > preferCount && avoidCount >= 2:
		return "retire", "negative evidence outweighs positive reuse value"
	case occurrences >= 5 && preferCount >= 5:
		return "resident", "high repeat value, strong positive signal, and still requires benchmark/quality-inspection before resident admission"
	default:
		return "mount-on-demand", "useful but should stay cold until the task needs it"
	}
}

func normalizeSpace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func extractSlideTexts(xmlText string) []string {
	matches := slideTextPattern.FindAllStringSubmatch(xmlText, -1)
	texts := []string{}
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		text := normalizeSpace(html.UnescapeString(match[1]))
		if text != "" {
			texts = append(texts, text)
		}
	}
	return texts
}

func countTextChars(texts []string) int {
	total := 0
	for _, text := range texts {
		total += len([]rune(strings.Join(strings.Fields(text), "")))
	}
	return total
}

func uniqueSorted(values []string) []string {
	if len(values) == 0 {
		return values
	}
	set := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		set[value] = struct{}{}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func containsAny(haystack string, markers []string) bool {
	for _, marker := range markers {
		if strings.Contains(haystack, strings.ToLower(marker)) {
			return true
		}
	}
	return false
}

func findPlaceholderHits(rawXML string, texts []string) []string {
	haystacks := []string{
		strings.ToLower(rawXML),
		strings.ToLower(strings.Join(texts, "\n")),
	}
	hits := []string{}
	for _, marker := range placeholderMarkers {
		needle := strings.ToLower(marker)
		for _, haystack := range haystacks {
			if strings.Contains(haystack, needle) {
				hits = append(hits, marker)
				break
			}
		}
	}
	return uniqueSorted(hits)
}

func slideTeachingSignals(slide slideSummary) []string {
	signals := []string{}
	lowerText := strings.ToLower(strings.Join(slide.TextSnippets, " "))
	for _, marker := range teachingMarkers {
		if strings.Contains(lowerText, strings.ToLower(marker)) {
			signals = append(signals, "tutorial-keywords")
			break
		}
	}
	if slide.TextChars >= 110 {
		signals = append(signals, "high-text-density")
	}
	if slide.TextCount >= 4 {
		signals = append(signals, "multi-step-content")
	}
	if slide.TextChars >= 60 && slide.TextCount >= 2 {
		signals = append(signals, "teaching-content")
	}
	return uniqueSorted(signals)
}

func slideNeedsIllustration(slide slideSummary) bool {
	if slide.PicCount > 0 {
		return false
	}
	signals := slideTeachingSignals(slide)
	return len(signals) > 0 || (slide.TextCount >= 5 && slide.TextChars >= 60)
}

func slideJoinedText(slide slideSummary) string {
	return strings.ToLower(strings.Join(slide.TextSnippets, " "))
}

func detectFixedPageRole(slide slideSummary, slideIndex int, totalSlides int) string {
	lower := slideJoinedText(slide)
	if slideIndex == 0 {
		return "cover"
	}
	if containsAny(lower, []string{"目录", "课程目录", "agenda", "table of contents", "table_of_contents"}) {
		return "agenda"
	}
	if containsAny(lower, []string{"致谢", "谢谢观看", "thanks", "thank you", "ending", "outro", "结尾页"}) {
		return "ending"
	}
	if containsAny(lower, []string{"总结", "结尾总结", "summary", "recap", "wrap up", "wrap-up", "conclusion"}) {
		return "summary"
	}
	if containsAny(lower, []string{"第一部分", "第二部分", "第三部分", "第四部分", "章节", "单元", "chapter", "section", "module", "part "}) && slide.TextCount <= 6 {
		return "section"
	}
	if slideIndex == totalSlides-1 && slide.TextCount <= 4 && slide.TextChars <= 70 {
		return "ending"
	}
	return "content"
}

func fixedPageTypeLabel(role string) string {
	switch role {
	case "cover":
		return "固定首页"
	case "agenda":
		return "固定目录页"
	case "section":
		return "固定单元页"
	case "summary":
		return "固定总结页"
	case "ending":
		return "固定结尾页"
	default:
		return "普通内容页"
	}
}

func styleLockPresetForDeck(pptxPath string, slides []slideSummary) jsonObject {
	combined := strings.ToLower(filepath.Base(pptxPath))
	for _, slide := range slides {
		combined += " " + slideJoinedText(slide)
	}
	if containsAny(combined, []string{"悦蓝", "yuelan", "霓虹", "neon", "赛博", "cyber"}) {
		return jsonObject{
			"style_name":      "neon-cyber-cartoon",
			"style_brief":     "霓虹赛博卡通风",
			"background":      "深紫蓝暗色底，禁止发白洗底。",
			"highlights":      "粉紫蓝霓虹高光、未来感描边与光效。",
			"illustrations":   "卡通化 UI / 教学插图 / 同风格界面示意。",
			"fixed_page_rule": "首页/目录页/单元页/总结页/结尾页如果在参考 deck 中存在，必须继续沿用同角色页型，不得改作普通内容页。",
			"prompt_rule":     "如果用户或模板已点名风格名，必须原样写进 image2 / 配图提示，不得自行改风格。",
			"keep_dark":       true,
			"forbid": []string{
				"白底",
				"写实照片",
				"电商宣传风",
				"随机扁平方框",
				"与模板不一致的浅色清新风",
			},
		}
	}
	return jsonObject{
		"style_name":      "inherit-reference-visual-system",
		"style_brief":     "继承参考 deck 的整体视觉系统",
		"background":      "保持参考 deck 的深浅极性；如果参考偏暗，不得切成白底。",
		"highlights":      "沿用参考 deck 的主色、强调色、边框和装饰层级。",
		"illustrations":   "保持参考 deck 同等级插图密度、用途与信息承载方式。",
		"fixed_page_rule": "首页/目录页/单元页/总结页/结尾页如果在参考 deck 中存在，必须继续沿用同角色页型，不得改作普通内容页。",
		"prompt_rule":     "用户或模板已点名的风格名必须原样写入配图提示，不得自由发挥改风格。",
		"keep_dark":       false,
		"forbid": []string{
			"无依据改白底",
			"只借颜色不借结构",
			"与参考不匹配的写实素材图",
			"临时自创丑方框",
		},
	}
}

func containsIllustrationStrategy(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	markers := []string{
		"截图",
		"示意图",
		"插图",
		"复用参考图",
		"复用现有",
		"换图",
		"image2",
		"教学配图",
	}
	for _, marker := range markers {
		if strings.Contains(lower, strings.ToLower(marker)) {
			return true
		}
	}
	return false
}

func illustrationPlanFailures(workspace string) []string {
	path, ok := existingPlanFile(workspace, "illustration-plan")
	if !ok || !nonEmpty(path) {
		return nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("illustration_plan_unreadable=%s", path)}
	}
	failures := []string{}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".json" {
		payload := map[string]any{}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return []string{fmt.Sprintf("illustration_plan_invalid_json=%s", path)}
		}
		slides, ok := objectSlice(payload, "slides")
		if !ok {
			return nil
		}
		for _, item := range slides {
			entry, ok := item.(map[string]any)
			if !ok {
				continue
			}
			slideLabel := "slide"
			if number, ok := intFromAny(entry["slide"]); ok && number > 0 {
				slideLabel = fmt.Sprintf("slide-%02d", number)
			}
			strategy, _ := stringFromAny(entry["strategy"])
			requiresVisual := false
			if value, ok := boolFromAny(entry["requires_visual"]); ok {
				requiresVisual = value
			} else if value, ok := boolFromAny(entry["requiresVisual"]); ok {
				requiresVisual = value
			}
			signals := []string{}
			if arr, ok := entry["signals"].([]any); ok {
				for _, value := range arr {
					if text, ok := stringFromAny(value); ok {
						signals = append(signals, text)
					}
				}
			}
			needsVisualBySignals := false
			for _, signal := range signals {
				if signal == "tutorial-keywords" || signal == "multi-step-content" || signal == "teaching-content" {
					needsVisualBySignals = true
					break
				}
			}
			if (requiresVisual || needsVisualBySignals) && !containsIllustrationStrategy(strategy) {
				failures = append(failures, fmt.Sprintf("illustration_plan_missing_visual_strategy=%s", slideLabel))
			}
		}
		return failures
	}

	lines := strings.Split(string(raw), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- slide-") {
			continue
		}
		lower := strings.ToLower(trimmed)
		needsVisual := strings.Contains(lower, "requires_visual=true") ||
			strings.Contains(lower, "tutorial-keywords") ||
			strings.Contains(lower, "multi-step-content") ||
			strings.Contains(lower, "teaching-content")
		if needsVisual && !containsIllustrationStrategy(trimmed) {
			label := strings.TrimPrefix(strings.SplitN(trimmed, ":", 2)[0], "- ")
			failures = append(failures, fmt.Sprintf("illustration_plan_missing_visual_strategy=%s", label))
		}
	}
	return uniqueSorted(failures)
}

func planFileText(path string) string {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func hasTeachingSignalsInPlan(path string) bool {
	lower := strings.ToLower(planFileText(path))
	return strings.Contains(lower, "tutorial-keywords") ||
		strings.Contains(lower, "multi-step-content") ||
		strings.Contains(lower, "teaching-content") ||
		strings.Contains(lower, "requires_visual=true")
}

func styleLockRequiresDarkBackground(workspace string) bool {
	path, ok := existingPlanFile(workspace, "style-lock")
	if !ok || !nonEmpty(path) {
		return false
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	if strings.EqualFold(filepath.Ext(path), ".json") {
		payload := map[string]any{}
		if err := json.Unmarshal(raw, &payload); err == nil {
			if keepDark, ok := boolFromAny(payload["keep_dark"]); ok {
				return keepDark
			}
		}
	}
	lower := strings.ToLower(string(raw))
	return strings.Contains(lower, "keep_dark_background: true")
}

func styleLockFailures(workspace string) []string {
	path, ok := existingPlanFile(workspace, "style-lock")
	if !ok || !nonEmpty(path) {
		return nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("style_lock_unreadable=%s", path)}
	}
	failures := []string{}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".json" {
		payload := map[string]any{}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return []string{fmt.Sprintf("style_lock_invalid_json=%s", path)}
		}
		requiredKeys := []string{"style_name", "style_brief", "background", "highlights", "illustrations", "fixed_page_rule", "prompt_rule"}
		for _, key := range requiredKeys {
			if objectString(payload, key) == "" {
				failures = append(failures, fmt.Sprintf("style_lock_missing_%s=%s", key, path))
			}
		}
		return uniqueSorted(failures)
	}
	lower := strings.ToLower(string(raw))
	for _, marker := range []string{"visual_system:", "background_policy:", "highlight_policy:", "illustration_policy:", "fixed_page_rule:", "prompt_rule:"} {
		if !strings.Contains(lower, marker) {
			failures = append(failures, fmt.Sprintf("style_lock_missing_marker=%s", marker))
		}
	}
	return uniqueSorted(failures)
}

func pageRolePolicyFailures(workspace string) []string {
	path, ok := existingPlanFile(workspace, "page-role-policy")
	if !ok || !nonEmpty(path) {
		return nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("page_role_policy_unreadable=%s", path)}
	}
	failures := []string{}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".json" {
		payload := map[string]any{}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return []string{fmt.Sprintf("page_role_policy_invalid_json=%s", path)}
		}
		lockedRoles, ok := objectSlice(payload, "locked_roles")
		if !ok || len(lockedRoles) == 0 {
			return nil
		}
		for _, rawItem := range lockedRoles {
			item, ok := rawItem.(map[string]any)
			if !ok {
				continue
			}
			role, _ := stringFromAny(item["role"])
			slideLabel := "slide"
			if number, ok := intFromAny(item["slide"]); ok && number > 0 {
				slideLabel = fmt.Sprintf("slide-%02d", number)
			}
			fixedPage, _ := boolFromAny(item["fixed_page"])
			doNotRepurpose, _ := boolFromAny(item["do_not_repurpose"])
			if role != "" && role != "content" && (!fixedPage || !doNotRepurpose) {
				failures = append(failures, fmt.Sprintf("page_role_policy_unlocked_fixed_role=%s[%s]", slideLabel, role))
			}
		}
		return uniqueSorted(failures)
	}
	lines := strings.Split(string(raw), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- slide-") {
			continue
		}
		lower := strings.ToLower(trimmed)
		if strings.Contains(lower, "[cover]") || strings.Contains(lower, "[agenda]") || strings.Contains(lower, "[section]") || strings.Contains(lower, "[summary]") || strings.Contains(lower, "[ending]") {
			if !strings.Contains(lower, "fixed_page=true") || !strings.Contains(lower, "do_not_repurpose=true") {
				label := strings.TrimPrefix(strings.SplitN(trimmed, ":", 2)[0], "- ")
				failures = append(failures, fmt.Sprintf("page_role_policy_unlocked_fixed_role=%s", label))
			}
		}
	}
	return uniqueSorted(failures)
}

func motionPlanRequired(workspace string) bool {
	path, ok := existingPlanFile(workspace, "motion-plan")
	if !ok || !nonEmpty(path) {
		return false
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	if strings.EqualFold(filepath.Ext(path), ".json") {
		payload := map[string]any{}
		if err := json.Unmarshal(raw, &payload); err == nil {
			if required, ok := boolFromAny(payload["required"]); ok {
				return required
			}
		}
	}
	lower := strings.ToLower(string(raw))
	return strings.Contains(lower, "required: true")
}

func motionPlanFailures(workspace string) []string {
	path, ok := existingPlanFile(workspace, "motion-plan")
	if !ok || !nonEmpty(path) {
		return nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("motion_plan_unreadable=%s", path)}
	}
	failures := []string{}
	required := false
	if strings.EqualFold(filepath.Ext(path), ".json") {
		payload := map[string]any{}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return []string{fmt.Sprintf("motion_plan_invalid_json=%s", path)}
		}
		if value, ok := boolFromAny(payload["required"]); ok {
			required = value
		}
		requiredKeys := []string{"dynamic_source", "motion_intent", "static_fallback", "gate_note"}
		for _, key := range requiredKeys {
			if objectString(payload, key) == "" {
				failures = append(failures, fmt.Sprintf("motion_plan_missing_%s=%s", key, path))
			}
		}
		if required && len(objectStringSlice(payload, "motion_roles")) == 0 {
			failures = append(failures, fmt.Sprintf("motion_plan_missing_motion_roles=%s", path))
		}
	} else {
		lower := strings.ToLower(string(raw))
		required = strings.Contains(lower, "required: true")
		for _, marker := range []string{"dynamic_source:", "motion_intent:", "static_fallback:", "gate_note:"} {
			if !strings.Contains(lower, marker) {
				failures = append(failures, fmt.Sprintf("motion_plan_missing_marker=%s", marker))
			}
		}
		if required && !strings.Contains(lower, "motion_roles:") {
			failures = append(failures, "motion_plan_missing_marker=motion_roles:")
		}
	}
	if required {
		if sourceArtifact, ok := motionPlanSourceArtifact(workspace); ok {
			if _, err := os.Stat(sourceArtifact); err != nil {
				failures = append(failures, fmt.Sprintf("motion_plan_missing_source_artifact=%s", sourceArtifact))
			}
		} else if !hasWorkspaceHTMLArtifact(workspace) {
			failures = append(failures, "motion_plan_required_but_no_live_html_demo_artifact")
		}
	}
	return uniqueSorted(failures)
}

func objectStringSlice(payload map[string]any, key string) []string {
	values := []string{}
	raw, ok := payload[key]
	if !ok {
		return values
	}
	switch typed := raw.(type) {
	case []any:
		for _, item := range typed {
			if text, ok := stringFromAny(item); ok && strings.TrimSpace(text) != "" {
				values = append(values, strings.TrimSpace(text))
			}
		}
	case []string:
		for _, item := range typed {
			if strings.TrimSpace(item) != "" {
				values = append(values, strings.TrimSpace(item))
			}
		}
	}
	return values
}

func motionPlanSourceArtifact(workspace string) (string, bool) {
	path, ok := existingPlanFile(workspace, "motion-plan")
	if !ok || !nonEmpty(path) {
		return "", false
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	if strings.EqualFold(filepath.Ext(path), ".json") {
		payload := map[string]any{}
		if err := json.Unmarshal(raw, &payload); err == nil {
			if value := strings.TrimSpace(objectString(payload, "source_artifact")); value != "" {
				if filepath.IsAbs(value) {
					return value, true
				}
				return filepath.Join(workspace, value), true
			}
		}
	}
	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToLower(trimmed), "- source_artifact:") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- source_artifact:"))
		if value == "" {
			return "", false
		}
		if filepath.IsAbs(value) {
			return value, true
		}
		return filepath.Join(workspace, value), true
	}
	return "", false
}

func hasWorkspaceHTMLArtifact(workspace string) bool {
	entries, err := os.ReadDir(workspace)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".html") {
			return true
		}
	}
	return false
}

func pilotPreviewLayoutFailures(path string) []string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("pilot_preview_layout_unreadable=%s", path)}
	}
	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return []string{fmt.Sprintf("pilot_preview_layout_invalid_json=%s", path)}
	}
	failures := []string{}
	if overflowCount, ok := intFromAny(payload["overflow_count"]); ok && overflowCount > 0 {
		failures = append(failures, fmt.Sprintf("pilot_preview_layout_overflow=%s count=%d", path, overflowCount))
	}
	if unsafeCount, ok := intFromAny(payload["unsafe_count"]); ok && unsafeCount > 0 {
		failures = append(failures, fmt.Sprintf("pilot_preview_layout_unsafe_area=%s count=%d", path, unsafeCount))
	}
	return uniqueSorted(failures)
}

func decodeImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	return img, err
}

func analyzePreviewImage(path string) (previewStats, error) {
	stats := previewStats{}
	img, err := decodeImage(path)
	if err != nil {
		return stats, err
	}
	bounds := img.Bounds()
	if bounds.Empty() {
		return stats, fmt.Errorf("image has empty bounds: %s", path)
	}
	totalPixels := bounds.Dx() * bounds.Dy()
	step := 1
	if totalPixels > 200000 {
		step = int(math.Sqrt(float64(totalPixels) / 200000.0))
		if step < 1 {
			step = 1
		}
	}
	var mean float64
	var m2 float64
	light := 0
	dark := 0
	samples := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			r16, g16, b16, _ := img.At(x, y).RGBA()
			r := float64(r16) / 257.0
			g := float64(g16) / 257.0
			b := float64(b16) / 257.0
			luma := 0.2126*r + 0.7152*g + 0.0722*b
			samples++
			delta := luma - mean
			mean += delta / float64(samples)
			m2 += delta * (luma - mean)
			if luma >= 235 {
				light++
			}
			if luma <= 45 {
				dark++
			}
		}
	}
	if samples == 0 {
		return stats, fmt.Errorf("image has no samples: %s", path)
	}
	variance := 0.0
	if samples > 1 {
		variance = m2 / float64(samples-1)
	}
	stats.SampleCount = samples
	stats.AverageLuma = mean / 255.0
	stats.StdDevLuma = math.Sqrt(variance) / 255.0
	stats.LightRatio = float64(light) / float64(samples)
	stats.DarkRatio = float64(dark) / float64(samples)
	return stats, nil
}

func pilotPreviewFailures(path string) []string {
	stats, err := analyzePreviewImage(path)
	if err != nil {
		return []string{fmt.Sprintf("pilot_preview_unreadable=%s", path)}
	}
	failures := []string{}
	if stats.AverageLuma >= 0.93 && stats.StdDevLuma <= 0.05 {
		failures = append(failures, fmt.Sprintf("pilot_preview_whitewashed=%s avg=%.2f std=%.2f", path, stats.AverageLuma, stats.StdDevLuma))
	}
	if stats.StdDevLuma <= 0.035 {
		failures = append(failures, fmt.Sprintf("pilot_preview_low_contrast=%s std=%.2f", path, stats.StdDevLuma))
	}
	if stats.LightRatio >= 0.92 && stats.DarkRatio <= 0.01 {
		failures = append(failures, fmt.Sprintf("pilot_preview_near_blank=%s light=%.2f dark=%.2f", path, stats.LightRatio, stats.DarkRatio))
	}
	return failures
}

func styleLockedPilotPreviewFailures(path string, requireDark bool) []string {
	if !requireDark {
		return nil
	}
	stats, err := analyzePreviewImage(path)
	if err != nil {
		return nil
	}
	failures := []string{}
	if stats.AverageLuma >= 0.72 {
		failures = append(failures, fmt.Sprintf("pilot_preview_breaks_dark_style_lock=%s avg=%.2f", path, stats.AverageLuma))
	}
	if stats.LightRatio >= 0.55 && stats.DarkRatio <= 0.12 {
		failures = append(failures, fmt.Sprintf("pilot_preview_too_light_for_dark_style_lock=%s light=%.2f dark=%.2f", path, stats.LightRatio, stats.DarkRatio))
	}
	return failures
}

func pilotApprovalGranted(path string) (bool, string) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Sprintf("pilot_approval_unreadable=%s", path)
	}
	lower := strings.ToLower(strings.TrimSpace(string(bytes)))
	if lower == "" {
		return false, fmt.Sprintf("pilot_approval_too_small=%s", path)
	}
	if strings.EqualFold(filepath.Ext(path), ".json") {
		payload := map[string]any{}
		if err := json.Unmarshal(bytes, &payload); err == nil {
			if approved, ok := boolFromAny(payload["approved"]); ok {
				if approved {
					return true, ""
				}
				return false, fmt.Sprintf("pilot_approval_rejected=%s", path)
			}
			for _, key := range []string{"status", "decision", "result"} {
				if value, ok := stringFromAny(payload[key]); ok {
					lower = strings.ToLower(value)
					break
				}
			}
		}
	}
	if containsAny(lower, pilotApprovalNegativeMarkers) {
		return false, fmt.Sprintf("pilot_approval_rejected=%s", path)
	}
	if containsAny(lower, pilotApprovalPositiveMarkers) {
		return true, ""
	}
	return false, fmt.Sprintf("pilot_approval_missing_explicit_approve=%s", path)
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
			texts := extractSlideTexts(text)
			summary.Slides = append(summary.Slides, slideSummary{
				Name:            filepath.Base(name),
				TextCount:       len(texts),
				PicCount:        strings.Count(text, "<p:pic"),
				ShapeCount:      strings.Count(text, "<p:sp"),
				TextChars:       countTextChars(texts),
				PlaceholderHits: findPlaceholderHits(text, texts),
				TextSnippets:    texts,
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
		taskLogPath := filepath.Join(workspace, "task-log.jsonl")
		if nonEmpty(taskLogPath) {
			records, err := loadJSONLines(taskLogPath)
			if err != nil || len(records) == 0 {
				failures = append(failures, "workflow_task_log_unreadable")
			} else {
				failures = append(failures, taskLogExecutionRhythmFailures(records)...)
				failures = append(failures, taskLogCloseoutLeakFailures(records)...)
				failures = append(failures, taskLogBlockedWaitFailures(records)...)
				last := records[len(records)-1]
				if objectString(last, "event") != "end" || objectString(last, "status") != "done" {
					failures = append(failures, "workflow_task_log_not_closed")
				}
				if path := objectString(last, "closeout_report"); (path == "" || !nonEmpty(resolveWorkspaceRef(workspace, path))) && !objectBoolValue(last, "closeout_report_verified") {
					failures = append(failures, "workflow_task_log_missing_closeout_report")
				}
				if path := objectString(last, "evidence_report"); (path == "" || !nonEmpty(resolveWorkspaceRef(workspace, path))) && !objectBoolValue(last, "evidence_report_verified") {
					failures = append(failures, "workflow_task_log_missing_evidence_report")
				} else if evidenceObj, err := loadJSONObject(resolveWorkspaceRef(workspace, path)); err != nil {
					if !objectBoolValue(last, "evidence_report_verified") {
						failures = append(failures, "workflow_evidence_report_unreadable")
					}
				} else {
					status := objectString(evidenceObj, "status")
					if status != "verified" && status != "shipped" {
						failures = append(failures, "workflow_evidence_report_not_verified")
					}
				}
			}
		}
	}
	return printGate("workflow-guard", failures)
}

func claimTextHasSuccessMarker(text string) bool {
	lower := strings.ToLower(text)
	for _, word := range successClaimMarkers {
		if strings.Contains(text, word) || strings.Contains(lower, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

func runtimeTokenAuditRequiredForTexts(values ...string) bool {
	markers := []string{
		"token", "tokens", "cached token", "cached tokens", "cache hit", "cache-hit", "prompt cache", "prefix cache",
		"context volume", "runtime context", "outer context", "api usage",
		"cost", "costs", "spend", "billing", "usage", "input tokens", "output tokens", "uncached", "fresh input",
		"\u8d39\u7528", "\u82b1\u8d39", "\u8d26\u5355", "\u540e\u53f0", "\u6d88\u8017", "\u547d\u4e2d",
		"\u7f13\u5b58", "\u4e0a\u4e0b\u6587", "\u5916\u5c42", "\u8bf7\u6c42\u4f53\u91cf", "\u771f\u5b9e\u6d88\u8017",
		"\u7701token", "\u8f93\u5165token", "\u8f93\u51fatoken", "\u672a\u547d\u4e2d", "\u65b0\u9c9c\u8f93\u5165",
		"费用", "花费", "账单", "后台", "消耗", "命中", "缓存", "上下文", "外层", "请求体量", "真实消耗", "省token",
	}
	for _, value := range values {
		if len(markerHits(value, markers)) > 0 {
			return true
		}
	}
	return false
}

func runtimeContextAuditReportPassFailures(workspace string, reportPath string) []string {
	if privateEvidencePathDenied(reportPath) {
		return []string{"runtime_context_audit_private_path_denied=" + pathPrivacyRef(workspace, reportPath)}
	}
	if strings.TrimSpace(workspace) != "" && !sameOrDescendant(reportPath, workspace) {
		return []string{"runtime_context_audit_outside_workspace=" + pathPrivacyRef(workspace, reportPath)}
	}
	if !nonEmpty(reportPath) {
		return []string{"runtime_context_audit_missing_or_too_small=" + pathPrivacyRef(workspace, reportPath)}
	}
	report, err := loadJSONObject(reportPath)
	if err != nil {
		return []string{"runtime_context_audit_unreadable=" + pathPrivacyRef(workspace, reportPath)}
	}
	failures := []string{}
	if objectString(report, "status") != "pass" {
		failures = append(failures, "runtime_context_audit_not_pass="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "command") != "runtime-context-audit" {
		failures = append(failures, "runtime_context_audit_command_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "schema_version") != "runtime-context-audit.v1" {
		failures = append(failures, "runtime_context_audit_schema_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "workspace_key") != privacyHash(absClean(workspace)) {
		failures = append(failures, "runtime_context_audit_workspace_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	if generatedAt := objectString(report, "generated_at"); generatedAt == "" {
		failures = append(failures, "runtime_context_audit_missing_generated_at="+pathPrivacyRef(workspace, reportPath))
	} else if _, err := time.Parse(time.RFC3339, generatedAt); err != nil {
		failures = append(failures, "runtime_context_audit_bad_generated_at="+pathPrivacyRef(workspace, reportPath))
	}
	if value, ok := intFromAny(report["usage_observations"]); !ok || value < minRuntimeUsageObservations {
		failures = append(failures, fmt.Sprintf("runtime_context_audit_usage_observations_below_floor=%d<%d", value, minRuntimeUsageObservations))
	}
	if objectString(report, "volume_gate") != "pass" {
		failures = append(failures, "runtime_context_audit_volume_gate_not_pass="+pathPrivacyRef(workspace, reportPath))
	}
	for _, key := range []string{"cached_tokens_p95", "input_tokens_p95", "fresh_input_tokens_p95", "output_tokens_p95", "uncached_tokens_p95"} {
		if _, ok := intFromAny(report[key]); !ok {
			failures = append(failures, "runtime_context_audit_missing_metric="+key)
		}
	}
	if inputs, ok := objectMap(report, "input_hashes"); !ok || len(inputs) == 0 {
		failures = append(failures, "runtime_context_audit_missing_input_hashes="+pathPrivacyRef(workspace, reportPath))
	} else {
		for rel, rawExpected := range inputs {
			expected, ok := rawExpected.(string)
			if !ok || expected == "" || strings.HasPrefix(expected, "__") {
				failures = append(failures, "runtime_context_audit_bad_input_hash="+rel)
				continue
			}
			current, err := fileSHA256(filepath.Join(workspace, filepath.FromSlash(rel)))
			if err != nil {
				failures = append(failures, "runtime_context_audit_input_unreadable="+rel)
				continue
			}
			if current != expected {
				failures = append(failures, "runtime_context_audit_input_hash_mismatch="+rel)
			}
		}
	}
	return uniqueStrings(failures)
}

func auditReportPassFailures(workspace string) []string {
	if strings.TrimSpace(workspace) == "" {
		return []string{"success_claim_requires_workspace_for_current_audits"}
	}
	workspace = absClean(workspace)
	requiredReports := map[string]string{
		"fusion-audit":        filepath.Join(workspace, "outputs", "fusion-audit-report.json"),
		"optimization-audit":  filepath.Join(workspace, "outputs", "optimization-audit-report.json"),
		"context-bloat-audit": filepath.Join(workspace, "outputs", "context-bloat-audit-report.json"),
	}
	failures := []string{}
	for gate, path := range requiredReports {
		if !sameOrDescendant(path, workspace) {
			failures = append(failures, gate+"_report_outside_workspace="+pathPrivacyRef(workspace, path))
			continue
		}
		report, err := loadJSONObject(path)
		if err != nil {
			failures = append(failures, gate+"_report_unreadable="+pathPrivacyRef(workspace, path))
			continue
		}
		if objectString(report, "status") != "pass" {
			failures = append(failures, gate+"_report_not_pass="+pathPrivacyRef(workspace, path))
		}
		if reportWorkspace := objectString(report, "workspace"); reportWorkspace != "" && absClean(reportWorkspace) != workspace {
			failures = append(failures, gate+"_report_workspace_mismatch="+privacyHash(reportWorkspace))
		}
		if reportWorkspaceKey := objectString(report, "workspace_key"); reportWorkspaceKey != "" && reportWorkspaceKey != privacyHash(workspace) {
			failures = append(failures, gate+"_report_workspace_key_mismatch="+pathPrivacyRef(workspace, path))
		}
		failures = append(failures, auditReportFreshnessFailures(report, workspace, gate)...)
	}
	return failures
}

func tokenOptimizationAuditPassFailures(workspace string) []string {
	if strings.TrimSpace(workspace) == "" {
		return []string{"token_optimization_claim_requires_workspace_for_runtime_context_audit"}
	}
	return runtimeContextAuditReportPassFailures(absClean(workspace), filepath.Join(absClean(workspace), "outputs", "runtime-context-audit-report.json"))
}

func claimGuard(args []string) int {
	claim, ok := argValue(args, "--claim")
	if !ok {
		usage()
		return 2
	}
	workspace, _ := argValue(args, "--workspace")
	evidence := argValues(args, "--evidence")
	makesSuccessClaim := claimTextHasSuccessMarker(claim)
	failures := []string{}
	if makesSuccessClaim {
		if len(evidence) == 0 {
			failures = append(failures, "success_claim_requires_evidence")
		}
		failures = append(failures, auditReportPassFailures(workspace)...)
		if runtimeTokenAuditRequiredForTexts(claim) {
			failures = append(failures, tokenOptimizationAuditPassFailures(workspace)...)
		}
	}
	for _, path := range evidence {
		if privateEvidencePathDenied(path) {
			failures = append(failures, "evidence_private_path_denied="+pathPrivacyRef(workspace, path))
		} else if !nonEmpty(path) {
			failures = append(failures, "evidence_missing_or_too_small="+pathPrivacyRef(workspace, path))
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
		if minutes >= 30 && !hasArtifact {
			failures = append(failures, fmt.Sprintf("no_verifiable_artifact_after_30_minutes phase=%s", phase))
		} else if minutes >= 15 && !hasArtifact && executionExploratoryPhases[phase] {
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
	phase, _ := argValue(args, "--phase")
	artifacts := argValues(args, "--artifact")
	closeoutReport, _ := argValue(args, "--closeout-report")
	evidenceReport, _ := argValue(args, "--evidence-report")
	auditWorkspace, _ := argValue(args, "--audit-workspace")
	closeoutVerified := false
	evidenceVerified := false
	allowedStatus := map[string]bool{
		"":               true,
		"running":        true,
		"blocked":        true,
		"needs_decision": true,
		"done":           true,
	}
	if !allowedStatus[status] {
		fmt.Fprintln(os.Stderr, "status must be running, blocked, needs_decision, or done")
		return 2
	}
	expectedStatusByEvent := map[string][]string{
		"start":     {"running"},
		"heartbeat": {"running", "blocked", "needs_decision"},
		"blocked":   {"blocked", "needs_decision"},
		"end":       {"done"},
	}
	if expected, ok := expectedStatusByEvent[event]; ok {
		valid := false
		for _, item := range expected {
			if status == item {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Fprintf(os.Stderr, "event %s requires status %s\n", event, strings.Join(expected, " or "))
			return 2
		}
	}
	if event == "end" && status == "done" {
		if !nonEmpty(closeoutReport) {
			fmt.Fprintln(os.Stderr, "event end with status done requires --closeout-report")
			return 2
		}
		if !nonEmpty(evidenceReport) {
			fmt.Fprintln(os.Stderr, "event end with status done requires --evidence-report")
			return 2
		}
		if failures := closeoutReportPassFailures(workspace, closeoutReport); len(failures) > 0 {
			fmt.Fprintln(os.Stderr, strings.Join(failures, "; "))
			return 2
		}
		if failures := evidenceReportPassFailures(workspace, evidenceReport); len(failures) > 0 {
			fmt.Fprintln(os.Stderr, strings.Join(failures, "; "))
			return 2
		}
		closeoutVerified = true
		evidenceVerified = true
		if strings.TrimSpace(auditWorkspace) == "" {
			auditWorkspace = absClean(filepath.Dir(workspace))
		}
		if failures := auditReportPassFailures(auditWorkspace); len(failures) > 0 {
			return printGate("task", failures)
		}
	}
	if (status == "blocked" || status == "needs_decision" || event == "blocked") && strings.TrimSpace(note) == "" {
		fmt.Fprintln(os.Stderr, "blocked or needs_decision status requires a concrete --note reason")
		return 2
	}
	entry := jsonObject{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"event":     event,
		"status":    status,
		"note":      note,
		"artifacts": artifacts,
	}
	if phase != "" {
		entry["phase"] = phase
	}
	if failures := executionRhythmFailures(entry); len(failures) > 0 {
		return printGate("task", failures)
	}
	if (status == "running" || event == "heartbeat") && len(markerHits(note, managementPauseMarkers)) > 0 {
		return printGate("task", []string{"running_note_waits_for_continue=" + strings.Join(markerHits(note, managementPauseMarkers), "|")})
	}
	if event == "end" && status == "done" {
		if hits := markerHits(note, closeoutLeakMarkers); len(hits) > 0 {
			return printGate("task", []string{"done_note_reopens_work=" + strings.Join(hits, "|")})
		}
	}
	if err := ensureDir(workspace); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	safeEntry := jsonObject{
		"timestamp":                entry["timestamp"],
		"event":                    event,
		"status":                   status,
		"note_key":                 privacyHash(note),
		"note_length":              len([]rune(strings.TrimSpace(note))),
		"artifact_refs":            pathPrivacyRefs(workspace, artifacts),
		"primary_artifact_present": hasPrimaryArtifact(artifacts),
		"execution_precheck_hits":  markerHits(strings.Join([]string{phase, note}, "\n"), executionPrecheckMarkers),
		"closeout_leak_hits":       markerHits(note, closeoutLeakMarkers),
		"management_pause_hits":    markerHits(note, managementPauseMarkers),
	}
	if phase != "" {
		safeEntry["phase"] = phase
	}
	if closeoutReport != "" {
		safeEntry["closeout_report"] = pathPrivacyRef(workspace, closeoutReport)
		safeEntry["closeout_report_verified"] = closeoutVerified || (event != "end" && nonEmpty(closeoutReport))
	}
	if evidenceReport != "" {
		safeEntry["evidence_report"] = pathPrivacyRef(workspace, evidenceReport)
		safeEntry["evidence_report_verified"] = evidenceVerified || (event != "end" && nonEmpty(evidenceReport))
	}
	logPath := filepath.Join(workspace, "task-log.jsonl")
	if err := appendJSONLine(logPath, safeEntry); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("GO task\n- log=%s\n", pathPrivacyRef(workspace, logPath))
	return 0
}

func codeMapCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	goal, ok := argValue(args, "--goal")
	if !ok || strings.TrimSpace(goal) == "" {
		usage()
		return 2
	}
	entry, ok := argValue(args, "--entry")
	if !ok || strings.TrimSpace(entry) == "" {
		usage()
		return 2
	}
	dependencies := uniqueStrings(argValues(args, "--dependency"))
	risks := uniqueStrings(argValues(args, "--risk"))
	verifications := uniqueStrings(argValues(args, "--verify"))
	if len(verifications) == 0 {
		fmt.Fprintln(os.Stderr, "code-map requires at least one --verify item")
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	if err := ensureDir(workspace); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	report := jsonObject{
		"goal":          strings.TrimSpace(goal),
		"entry":         strings.TrimSpace(entry),
		"dependencies":  dependencies,
		"risks":         risks,
		"verifications": verifications,
		"generated_at":  time.Now().UTC().Format(time.RFC3339),
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(workspace, "code-map.json")
	}
	if err := writeJSON(outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("GO code-map\n- report=%s\n", absClean(outputPath))
	return 0
}

func textEvidenceSummary(text string) jsonObject {
	trimmed := strings.TrimSpace(text)
	return jsonObject{
		"key":    privacyHash(trimmed),
		"length": len([]rune(trimmed)),
	}
}

func textEvidenceSummaries(values []string) []jsonObject {
	result := []jsonObject{}
	for _, value := range uniqueStrings(values) {
		result = append(result, textEvidenceSummary(value))
	}
	return result
}

func pathEvidenceRefs(workspace string, paths []string) []jsonObject {
	result := []jsonObject{}
	for _, path := range uniqueStrings(paths) {
		ref := jsonObject{
			"path_ref": pathPrivacyRef(workspace, path),
		}
		if privateEvidencePathDenied(path) {
			ref["status"] = "denied-private-path"
			result = append(result, ref)
			continue
		}
		ref["bytes"] = fileSize(path)
		if hash, err := fileSHA256(path); err == nil {
			ref["sha256"] = hash
		}
		result = append(result, ref)
	}
	return result
}

func pathPrivacyRefs(workspace string, paths []string) []string {
	refs := []string{}
	for _, path := range uniqueStrings(paths) {
		refs = append(refs, pathPrivacyRef(workspace, path))
	}
	return refs
}

func resolveWorkspaceRef(workspace string, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" || strings.HasPrefix(ref, "external:") {
		return ref
	}
	if filepath.IsAbs(ref) || strings.Contains(filepath.ToSlash(ref), ":/") {
		return ref
	}
	return filepath.Join(workspace, filepath.FromSlash(ref))
}

func privateEvidencePathDenied(path string) bool {
	normalized := strings.ToLower(filepath.ToSlash(absClean(path)))
	markers := []string{
		"/.opensquilla/", "/.codex/auth", "/.codex/session", "/.codex/sessions", "/.codex/attachments",
		"/cookies", "/cookie", "/tokens", "/token", "/secrets", "/secret", "/credentials", "/credential", "/keychain",
	}
	for _, marker := range markers {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	base := strings.ToLower(filepath.Base(path))
	baseMarkers := []string{"cookie", "token", "secret", "credential", "apikey", "api_key", "private_key"}
	for _, marker := range baseMarkers {
		if strings.Contains(base, marker) {
			return true
		}
	}
	return false
}

func containsExactString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func checklistContainsMarker(obj map[string]any, marker string) bool {
	needle := strings.ToLower(marker)
	for _, rawGroup := range obj {
		items, ok := rawGroup.([]any)
		if !ok {
			continue
		}
		for _, rawItem := range items {
			text, ok := rawItem.(string)
			if ok && strings.Contains(strings.ToLower(text), needle) {
				return true
			}
		}
	}
	return false
}

func hotpathColdLedgerHandleOnly(obj map[string]any, relPath string) bool {
	items, ok := objectSlice(obj, "cold_ledger")
	if !ok {
		return false
	}
	needle := filepath.ToSlash(relPath)
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if filepath.ToSlash(objectString(item, "path")) == needle && objectString(item, "default_mode") == "handle-only" {
			return true
		}
	}
	return false
}

func hotpathForbiddenContains(obj map[string]any, marker string) bool {
	items, ok := objectSlice(obj, "forbidden_resident")
	if !ok {
		return false
	}
	needle := strings.ToLower(marker)
	for _, raw := range items {
		text, ok := raw.(string)
		if ok && strings.Contains(strings.ToLower(filepath.ToSlash(text)), needle) {
			return true
		}
	}
	return false
}

func hasSecretLikeText(values ...string) bool {
	for _, value := range values {
		if containsSecretLikeContent(value) {
			return true
		}
	}
	return false
}

func patchDebtMentioned(values ...string) bool {
	markers := []string{
		"temporary patch", "hotfix", "workaround", "quick fix", "bypass", "patch debt", "technical debt", "rule debt",
		"补丁", "临时", "绕过", "权宜", "债务", "热修", "先这样",
	}
	for _, value := range values {
		if len(markerHits(value, markers)) > 0 {
			return true
		}
	}
	return false
}

func weakPatchDebtAction(action string) bool {
	normalized := normalizeSpace(strings.ToLower(action))
	if normalized == "" {
		return true
	}
	weak := map[string]bool{
		"none": true, "no": true, "n/a": true, "na": true, "not needed": true, "skip": true,
		"无": true, "没有": true, "不需要": true, "跳过": true,
	}
	if weak[normalized] {
		return true
	}
	return len([]rune(strings.TrimSpace(action))) < 16
}

func weakRootCauseText(rootCause string, symptom string, repro string) bool {
	normalized := normalizeSpace(strings.ToLower(rootCause))
	if normalized == "" {
		return true
	}
	if normalized == normalizeSpace(strings.ToLower(symptom)) || normalized == normalizeSpace(strings.ToLower(repro)) {
		return true
	}
	weakExact := map[string]bool{
		"bug": true, "failure": true, "error": true, "not working": true, "unknown": true, "not sure": true, "symptom": true,
		"报错": true, "失败": true, "不工作": true, "未知": true, "原因不明": true, "还没确定": true,
	}
	if weakExact[normalized] {
		return true
	}
	if len([]rune(strings.TrimSpace(rootCause))) < 24 {
		return true
	}
	symptomMarkers := []string{"not working", "still failing", "error", "bug", "failure", "broken", "报错", "失败", "不工作", "坏了", "卡住"}
	causeMarkers := []string{"because", "caused by", "due to", "missing", "mismatch", "race", "timeout", "nil", "null", "root", "因为", "导致", "缺失", "不一致", "竞争", "超时", "根因"}
	return len(markerHits(rootCause, symptomMarkers)) > 0 && len(markerHits(rootCause, causeMarkers)) == 0
}

func weakSameClassScanText(scan string) bool {
	normalized := normalizeSpace(strings.ToLower(scan))
	if normalized == "" {
		return true
	}
	weakExact := map[string]bool{
		"none": true, "no": true, "n/a": true, "na": true, "not needed": true, "no scan": true, "skip": true,
		"无": true, "没有": true, "不需要": true, "未扫描": true, "跳过": true,
	}
	if weakExact[normalized] {
		return true
	}
	return len([]rune(strings.TrimSpace(scan))) < 20
}

func rootCauseEvidenceFailures(workspace string, symptom string, repro string, hypotheses []string, rootCause string, sameClassScan string, sameClassEvidence []string, fixStrategy string, regressionEvidence []string, patchDebtAction string, artifacts []string) []string {
	failures := []string{}
	if strings.TrimSpace(workspace) == "" {
		failures = append(failures, "root_cause_requires_workspace")
	}
	if strings.TrimSpace(symptom) == "" {
		failures = append(failures, "root_cause_requires_symptom")
	}
	if strings.TrimSpace(repro) == "" {
		failures = append(failures, "root_cause_requires_repro")
	}
	if len(uniqueStrings(hypotheses)) == 0 {
		failures = append(failures, "root_cause_requires_hypothesis")
	}
	if strings.TrimSpace(rootCause) == "" {
		failures = append(failures, "root_cause_requires_root_cause")
	} else if weakRootCauseText(rootCause, symptom, repro) {
		failures = append(failures, "root_cause_looks_symptom_only")
	}
	if weakSameClassScanText(sameClassScan) {
		failures = append(failures, "root_cause_requires_same_class_scan")
	}
	if len(sameClassEvidence) == 0 {
		failures = append(failures, "root_cause_requires_same_class_evidence")
	}
	for _, path := range sameClassEvidence {
		if privateEvidencePathDenied(path) {
			failures = append(failures, "root_cause_same_class_evidence_private_path_denied="+pathPrivacyRef(workspace, path))
		} else if strings.TrimSpace(workspace) != "" && !sameOrDescendant(path, workspace) {
			failures = append(failures, "root_cause_same_class_evidence_outside_workspace="+pathPrivacyRef(workspace, path))
		} else if !nonEmpty(path) {
			failures = append(failures, "root_cause_same_class_evidence_missing_or_too_small="+pathPrivacyRef(workspace, path))
		}
	}
	if weakSameClassScanText(fixStrategy) {
		failures = append(failures, "root_cause_requires_fix_strategy")
	}
	if len(regressionEvidence) == 0 {
		failures = append(failures, "root_cause_requires_regression_evidence")
	}
	for _, path := range regressionEvidence {
		if privateEvidencePathDenied(path) {
			failures = append(failures, "root_cause_regression_evidence_private_path_denied="+pathPrivacyRef(workspace, path))
		} else if strings.TrimSpace(workspace) != "" && !sameOrDescendant(path, workspace) {
			failures = append(failures, "root_cause_regression_evidence_outside_workspace="+pathPrivacyRef(workspace, path))
		} else if !nonEmpty(path) {
			failures = append(failures, "root_cause_regression_evidence_missing_or_too_small="+pathPrivacyRef(workspace, path))
		}
	}
	for _, path := range artifacts {
		if privateEvidencePathDenied(path) {
			failures = append(failures, "root_cause_artifact_private_path_denied="+pathPrivacyRef(workspace, path))
		} else if strings.TrimSpace(workspace) != "" && !sameOrDescendant(path, workspace) {
			failures = append(failures, "root_cause_artifact_outside_workspace="+pathPrivacyRef(workspace, path))
		} else if !nonEmpty(path) {
			failures = append(failures, "root_cause_artifact_missing_or_too_small="+pathPrivacyRef(workspace, path))
		}
	}
	secretTexts := append([]string{symptom, repro, rootCause, sameClassScan, fixStrategy, patchDebtAction}, hypotheses...)
	if hasSecretLikeText(secretTexts...) {
		failures = append(failures, "root_cause_secret_like_text_rejected")
	}
	patchTexts := append([]string{symptom, rootCause, sameClassScan, fixStrategy}, hypotheses...)
	if patchDebtMentioned(patchTexts...) && weakPatchDebtAction(patchDebtAction) {
		failures = append(failures, "root_cause_patch_debt_action_required")
	}
	return uniqueStrings(failures)
}

func reportTextSummaryFailures(report map[string]any, key string, reportPath string, workspace string) []string {
	summary, ok := objectMap(report, key)
	if !ok {
		return []string{"root_cause_report_missing_" + key + "=" + pathPrivacyRef(workspace, reportPath)}
	}
	failures := []string{}
	if objectString(summary, "key") == "" {
		failures = append(failures, "root_cause_report_"+key+"_missing_key="+pathPrivacyRef(workspace, reportPath))
	}
	length, ok := intFromAny(summary["length"])
	if !ok || length <= 0 {
		failures = append(failures, "root_cause_report_"+key+"_missing_length="+pathPrivacyRef(workspace, reportPath))
	}
	if _, hasRawText := summary["text"]; hasRawText {
		failures = append(failures, "root_cause_report_"+key+"_raw_text_forbidden="+pathPrivacyRef(workspace, reportPath))
	}
	return failures
}

func reportTextSummaryListFailures(report map[string]any, key string, reportPath string, workspace string) []string {
	items, ok := objectSlice(report, key)
	if !ok || len(items) == 0 {
		return []string{"root_cause_report_missing_" + key + "=" + pathPrivacyRef(workspace, reportPath)}
	}
	failures := []string{}
	for index, rawItem := range items {
		item, ok := rawItem.(map[string]any)
		if !ok {
			failures = append(failures, fmt.Sprintf("root_cause_report_%s_%d_invalid=%s", key, index+1, pathPrivacyRef(workspace, reportPath)))
			continue
		}
		if objectString(item, "key") == "" {
			failures = append(failures, fmt.Sprintf("root_cause_report_%s_%d_missing_key=%s", key, index+1, pathPrivacyRef(workspace, reportPath)))
		}
		length, ok := intFromAny(item["length"])
		if !ok || length <= 0 {
			failures = append(failures, fmt.Sprintf("root_cause_report_%s_%d_missing_length=%s", key, index+1, pathPrivacyRef(workspace, reportPath)))
		}
		if _, hasRawText := item["text"]; hasRawText {
			failures = append(failures, fmt.Sprintf("root_cause_report_%s_%d_raw_text_forbidden=%s", key, index+1, pathPrivacyRef(workspace, reportPath)))
		}
	}
	return failures
}

func reportEvidenceRefFailures(workspace string, reportPath string, report map[string]any, key string) []string {
	refs, ok := objectSlice(report, key)
	if !ok || len(refs) == 0 {
		return []string{"root_cause_report_missing_" + key + "=" + pathPrivacyRef(workspace, reportPath)}
	}
	failures := []string{}
	for index, rawRef := range refs {
		ref, ok := rawRef.(map[string]any)
		if !ok {
			failures = append(failures, fmt.Sprintf("root_cause_report_%s_%d_invalid=%s", key, index+1, pathPrivacyRef(workspace, reportPath)))
			continue
		}
		pathRef := objectString(ref, "path_ref")
		if pathRef == "" {
			failures = append(failures, fmt.Sprintf("root_cause_report_%s_%d_missing_path_ref=%s", key, index+1, pathPrivacyRef(workspace, reportPath)))
		}
		bytesValue, ok := intFromAny(ref["bytes"])
		if !ok || bytesValue < 20 {
			failures = append(failures, fmt.Sprintf("root_cause_report_%s_%d_bad_bytes=%s", key, index+1, pathPrivacyRef(workspace, reportPath)))
		}
		sha := objectString(ref, "sha256")
		if len(sha) != 64 {
			failures = append(failures, fmt.Sprintf("root_cause_report_%s_%d_bad_sha256=%s", key, index+1, pathPrivacyRef(workspace, reportPath)))
		}
		if pathRef != "" {
			if strings.HasPrefix(pathRef, "external:") || pathRef == "." || filepath.IsAbs(pathRef) || strings.Contains(pathRef, ":") {
				failures = append(failures, fmt.Sprintf("root_cause_report_%s_%d_unsafe_path_ref=%s", key, index+1, pathPrivacyRef(workspace, reportPath)))
				continue
			}
			cleanRef := filepath.Clean(filepath.FromSlash(pathRef))
			if cleanRef == ".." || strings.HasPrefix(cleanRef, ".."+string(os.PathSeparator)) {
				failures = append(failures, fmt.Sprintf("root_cause_report_%s_%d_parent_path_ref_forbidden=%s", key, index+1, pathPrivacyRef(workspace, reportPath)))
				continue
			}
			localPath := filepath.Join(workspace, cleanRef)
			if !sameOrDescendant(localPath, workspace) || privateEvidencePathDenied(localPath) {
				failures = append(failures, fmt.Sprintf("root_cause_report_%s_%d_local_evidence_private_or_outside=%s", key, index+1, pathPrivacyRef(workspace, localPath)))
				continue
			}
			if !nonEmpty(localPath) {
				failures = append(failures, fmt.Sprintf("root_cause_report_%s_%d_local_evidence_missing=%s", key, index+1, pathPrivacyRef(workspace, localPath)))
			} else if sha != "" {
				if current, err := fileSHA256(localPath); err != nil || current != sha {
					failures = append(failures, fmt.Sprintf("root_cause_report_%s_%d_local_evidence_hash_mismatch=%s", key, index+1, pathPrivacyRef(workspace, localPath)))
				}
			}
		}
	}
	return failures
}

func rootCauseReportPassFailures(workspace string, reportPath string) []string {
	if privateEvidencePathDenied(reportPath) {
		return []string{"root_cause_report_private_path_denied=" + pathPrivacyRef(workspace, reportPath)}
	}
	if strings.TrimSpace(workspace) != "" && !sameOrDescendant(reportPath, workspace) {
		return []string{"root_cause_report_outside_workspace=" + pathPrivacyRef(workspace, reportPath)}
	}
	if !nonEmpty(reportPath) {
		return []string{"root_cause_report_missing_or_too_small=" + pathPrivacyRef(workspace, reportPath)}
	}
	report, err := loadJSONObject(reportPath)
	if err != nil {
		return []string{"root_cause_report_unreadable=" + pathPrivacyRef(workspace, reportPath)}
	}
	failures := []string{}
	if objectString(report, "status") != "pass" {
		failures = append(failures, "root_cause_report_not_pass="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "officer") != "root-cause-officer" {
		failures = append(failures, "root_cause_report_wrong_officer="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "command") != "root-cause-radar" {
		failures = append(failures, "root_cause_report_command_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "schema_version") != "root-cause-radar.v1" {
		failures = append(failures, "root_cause_report_schema_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "verdict") != "root-cause-repair-ready" {
		failures = append(failures, "root_cause_report_verdict_not_ready="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "privacy_mode") != "hash-length-and-evidence-ref-only" {
		failures = append(failures, "root_cause_report_privacy_mode_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	if generatedAt := objectString(report, "generated_at"); generatedAt == "" {
		failures = append(failures, "root_cause_report_missing_generated_at="+pathPrivacyRef(workspace, reportPath))
	} else if _, err := time.Parse(time.RFC3339, generatedAt); err != nil {
		failures = append(failures, "root_cause_report_bad_generated_at="+pathPrivacyRef(workspace, reportPath))
	}
	if expected := privacyHash(absClean(workspace)); objectString(report, "workspace_key") != expected {
		failures = append(failures, "root_cause_report_workspace_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	for _, key := range []string{"symptom", "repro", "root_cause", "same_class_scan", "fix_strategy"} {
		failures = append(failures, reportTextSummaryFailures(report, key, reportPath, workspace)...)
	}
	failures = append(failures, reportTextSummaryListFailures(report, "hypotheses", reportPath, workspace)...)
	patchRequired, hasPatchRequired := objectBool(report, "patch_debt_required")
	if !hasPatchRequired {
		failures = append(failures, "root_cause_report_missing_patch_debt_required="+pathPrivacyRef(workspace, reportPath))
	} else if patchRequired {
		failures = append(failures, reportTextSummaryFailures(report, "patch_debt_action", reportPath, workspace)...)
	}
	failures = append(failures, reportEvidenceRefFailures(workspace, reportPath, report, "same_class_evidence_refs")...)
	failures = append(failures, reportEvidenceRefFailures(workspace, reportPath, report, "regression_evidence_refs")...)
	return failures
}

func closeoutReportPassFailures(workspace string, reportPath string) []string {
	if privateEvidencePathDenied(reportPath) {
		return []string{"closeout_report_private_path_denied=" + pathPrivacyRef(workspace, reportPath)}
	}
	if strings.TrimSpace(workspace) != "" && !sameOrDescendant(reportPath, workspace) {
		return []string{"closeout_report_outside_workspace=" + pathPrivacyRef(workspace, reportPath)}
	}
	if !nonEmpty(reportPath) {
		return []string{"closeout_report_missing_or_too_small=" + pathPrivacyRef(workspace, reportPath)}
	}
	report, err := loadJSONObject(reportPath)
	if err != nil {
		return []string{"closeout_report_unreadable=" + pathPrivacyRef(workspace, reportPath)}
	}
	failures := []string{}
	if objectString(report, "status") != "pass" {
		failures = append(failures, "closeout_report_not_pass="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "command") != "closeout-check" {
		failures = append(failures, "closeout_report_command_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "schema_version") != "closeout-check.v1" {
		failures = append(failures, "closeout_report_schema_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "privacy_mode") != "hash-length-and-evidence-ref-only" {
		failures = append(failures, "closeout_report_privacy_mode_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "workspace_key") != privacyHash(absClean(workspace)) {
		failures = append(failures, "closeout_report_workspace_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "audit_workspace_key") == "" {
		failures = append(failures, "closeout_report_missing_audit_workspace_key="+pathPrivacyRef(workspace, reportPath))
	}
	if needsDecision, ok := objectBool(report, "needs_user_decision"); !ok {
		failures = append(failures, "closeout_report_missing_needs_user_decision="+pathPrivacyRef(workspace, reportPath))
	} else if needsDecision {
		failures = append(failures, "closeout_report_needs_user_decision="+pathPrivacyRef(workspace, reportPath))
	}
	if gaps, ok := objectSlice(report, "remaining_gaps"); ok && len(gaps) > 0 {
		failures = append(failures, "closeout_report_has_remaining_gaps="+pathPrivacyRef(workspace, reportPath))
	}
	rootCauseRequired, hasRootCauseRequired := objectBool(report, "root_cause_required")
	if !hasRootCauseRequired {
		failures = append(failures, "closeout_report_missing_root_cause_required="+pathPrivacyRef(workspace, reportPath))
	}
	rootCauseVerified, hasRootCauseVerified := objectBool(report, "root_cause_report_verified")
	if !hasRootCauseVerified {
		failures = append(failures, "closeout_report_missing_root_cause_report_verified="+pathPrivacyRef(workspace, reportPath))
	} else if rootCauseRequired && !rootCauseVerified {
		failures = append(failures, "closeout_report_root_cause_unverified="+pathPrivacyRef(workspace, reportPath))
	}
	failures = append(failures, reportEvidenceRefFailures(workspace, reportPath, report, "artifact_refs")...)
	failures = append(failures, reportTextSummaryListFailures(report, "verifications", reportPath, workspace)...)
	return uniqueStrings(failures)
}

func evidenceReportPassFailures(workspace string, reportPath string) []string {
	if privateEvidencePathDenied(reportPath) {
		return []string{"evidence_report_private_path_denied=" + pathPrivacyRef(workspace, reportPath)}
	}
	if strings.TrimSpace(workspace) != "" && !sameOrDescendant(reportPath, workspace) {
		return []string{"evidence_report_outside_workspace=" + pathPrivacyRef(workspace, reportPath)}
	}
	if !nonEmpty(reportPath) {
		return []string{"evidence_report_missing_or_too_small=" + pathPrivacyRef(workspace, reportPath)}
	}
	report, err := loadJSONObject(reportPath)
	if err != nil {
		return []string{"evidence_report_unreadable=" + pathPrivacyRef(workspace, reportPath)}
	}
	failures := []string{}
	if objectString(report, "command") != "evidence-grade" {
		failures = append(failures, "evidence_report_command_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "schema_version") != "evidence-grade.v1" {
		failures = append(failures, "evidence_report_schema_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "privacy_mode") != "hash-length-and-evidence-ref-only" {
		failures = append(failures, "evidence_report_privacy_mode_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	if objectString(report, "workspace_key") != privacyHash(absClean(workspace)) {
		failures = append(failures, "evidence_report_workspace_mismatch="+pathPrivacyRef(workspace, reportPath))
	}
	status := objectString(report, "status")
	if status != "verified" && status != "shipped" {
		failures = append(failures, "evidence_report_status_not_verified="+pathPrivacyRef(workspace, reportPath))
	}
	failures = append(failures, reportTextSummaryFailures(report, "summary", reportPath, workspace)...)
	failures = append(failures, reportEvidenceRefFailures(workspace, reportPath, report, "artifact_refs")...)
	if _, hasRawSummary := report["summary_text"]; hasRawSummary {
		failures = append(failures, "evidence_report_raw_summary_forbidden="+pathPrivacyRef(workspace, reportPath))
	}
	if _, hasRawArtifacts := report["artifacts"]; hasRawArtifacts {
		failures = append(failures, "evidence_report_raw_artifacts_forbidden="+pathPrivacyRef(workspace, reportPath))
	}
	return uniqueStrings(failures)
}

func reportHasRootCauseSignal(report map[string]any) bool {
	if value, ok := objectBool(report, "root_cause_required"); ok && value {
		return true
	}
	if value, ok := objectBool(report, "root_cause_report_verified"); ok && value {
		return true
	}
	switch objectString(report, "command") {
	case "bugfix-guard", "root-cause-radar":
		return true
	}
	if atoms, ok := objectMap(report, "capability_mounts"); ok {
		if containsExactString(stringSlice(atoms, "distilled_atoms"), "root-cause-radar") {
			return true
		}
	}
	if taskRoute, ok := objectMap(report, "task_route"); ok {
		if containsExactString(stringSlice(taskRoute, "oversight_chain"), "root-cause-officer") {
			return true
		}
	}
	if deterministic, ok := objectMap(report, "deterministic_execution"); ok {
		if containsExactString(stringSlice(deterministic, "command_candidates"), "root-cause-radar") {
			return true
		}
	}
	return false
}

func artifactReportsRequireRootCause(workspace string, artifacts []string) bool {
	for _, artifact := range artifacts {
		if privateEvidencePathDenied(artifact) || !nonEmpty(artifact) {
			continue
		}
		if strings.TrimSpace(workspace) != "" && !sameOrDescendant(artifact, workspace) {
			continue
		}
		report, err := loadJSONObject(artifact)
		if err == nil && reportHasRootCauseSignal(report) {
			return true
		}
	}
	return false
}

func rootCauseRadarCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	symptom, ok := argValue(args, "--symptom")
	if !ok {
		usage()
		return 2
	}
	repro, ok := argValue(args, "--repro")
	if !ok {
		usage()
		return 2
	}
	rootCause, ok := argValue(args, "--root-cause")
	if !ok {
		usage()
		return 2
	}
	sameClassScan, ok := argValue(args, "--same-class-scan")
	if !ok {
		usage()
		return 2
	}
	fixStrategy, ok := argValue(args, "--fix-strategy")
	if !ok {
		usage()
		return 2
	}
	hypotheses := uniqueStrings(argValues(args, "--hypothesis"))
	eliminatedCauses := uniqueStrings(argValues(args, "--eliminated-cause"))
	sameClassEvidence := uniqueStrings(argValues(args, "--same-class-evidence"))
	regressionEvidence := uniqueStrings(argValues(args, "--regression-evidence"))
	artifacts := uniqueStrings(argValues(args, "--artifact"))
	patchDebtAction, _ := argValue(args, "--patch-debt-action")
	reportPath, hasReport := argValue(args, "--report")

	allHypotheses := append([]string{}, hypotheses...)
	allHypotheses = append(allHypotheses, eliminatedCauses...)
	failures := rootCauseEvidenceFailures(workspace, symptom, repro, hypotheses, rootCause, sameClassScan, sameClassEvidence, fixStrategy, regressionEvidence, patchDebtAction, artifacts)
	report := jsonObject{
		"officer":                  "root-cause-officer",
		"command":                  "root-cause-radar",
		"schema_version":           "root-cause-radar.v1",
		"generated_at":             time.Now().UTC().Format(time.RFC3339),
		"workspace_key":            privacyHash(absClean(workspace)),
		"privacy_mode":             "hash-length-and-evidence-ref-only",
		"symptom":                  textEvidenceSummary(symptom),
		"repro":                    textEvidenceSummary(repro),
		"hypotheses":               textEvidenceSummaries(hypotheses),
		"eliminated_causes":        textEvidenceSummaries(eliminatedCauses),
		"root_cause":               textEvidenceSummary(rootCause),
		"same_class_scan":          textEvidenceSummary(sameClassScan),
		"same_class_evidence_refs": pathEvidenceRefs(workspace, sameClassEvidence),
		"fix_strategy":             textEvidenceSummary(fixStrategy),
		"patch_debt_required":      patchDebtMentioned(append([]string{symptom, rootCause, sameClassScan, fixStrategy}, allHypotheses...)...),
		"patch_debt_action":        textEvidenceSummary(patchDebtAction),
		"regression_evidence_refs": pathEvidenceRefs(workspace, regressionEvidence),
		"artifact_refs":            pathEvidenceRefs(workspace, artifacts),
		"verdict":                  "root-cause-repair-ready",
		"status":                   "pass",
	}
	if len(failures) > 0 {
		report["status"] = "fail"
		report["verdict"] = "root-cause-repair-blocked"
		report["failures"] = failures
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(workspace, "root-cause-report.json")
	}
	if err := ensureDir(workspace); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := writeJSON(outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return printGate("root-cause-radar", failures)
}

func bugfixGuardCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	goal, ok := argValue(args, "--goal")
	if !ok || strings.TrimSpace(goal) == "" {
		usage()
		return 2
	}
	repro, ok := argValue(args, "--repro")
	if !ok || strings.TrimSpace(repro) == "" {
		usage()
		return 2
	}
	artifacts := uniqueStrings(argValues(args, "--artifact"))
	verifications := uniqueStrings(argValues(args, "--verify"))
	selfTests := uniqueStrings(argValues(args, "--self-test"))
	independentChecks := uniqueStrings(argValues(args, "--independent-check"))
	browserChecks := uniqueStrings(argValues(args, "--browser-check"))
	stillFailing := uniqueStrings(argValues(args, "--still-failing"))
	rootCauseReport, _ := argValue(args, "--root-cause-report")
	reportPath, hasReport := argValue(args, "--report")

	failures := []string{}
	if containsSecretLikeContent(goal) || containsSecretLikeContent(repro) || hasSecretLikeText(verifications...) || hasSecretLikeText(stillFailing...) {
		failures = append(failures, "bugfix_secret_like_text_rejected")
	}
	rootCauseReportVerified := false
	if rootCauseReport == "" {
		failures = append(failures, "bugfix_requires_root_cause_report")
	} else {
		rootCauseFailures := rootCauseReportPassFailures(workspace, rootCauseReport)
		if len(rootCauseFailures) == 0 {
			rootCauseReportVerified = true
		}
		failures = append(failures, rootCauseFailures...)
	}
	if len(artifacts) == 0 {
		failures = append(failures, "bugfix_requires_primary_artifact")
	}
	for _, artifact := range artifacts {
		if privateEvidencePathDenied(artifact) {
			failures = append(failures, "artifact_private_path_denied="+pathPrivacyRef(workspace, artifact))
		} else if strings.TrimSpace(workspace) != "" && !sameOrDescendant(artifact, workspace) {
			failures = append(failures, "artifact_outside_workspace="+pathPrivacyRef(workspace, artifact))
		} else if !nonEmpty(artifact) {
			failures = append(failures, "artifact_missing_or_too_small="+pathPrivacyRef(workspace, artifact))
		}
	}
	if len(verifications) == 0 {
		failures = append(failures, "bugfix_requires_verification")
	}
	if len(selfTests) == 0 {
		failures = append(failures, "bugfix_requires_self_test")
	}
	for _, path := range selfTests {
		if privateEvidencePathDenied(path) {
			failures = append(failures, "self_test_private_path_denied="+pathPrivacyRef(workspace, path))
		} else if !nonEmpty(path) {
			failures = append(failures, "self_test_missing_or_too_small="+pathPrivacyRef(workspace, path))
		}
	}
	if len(independentChecks) == 0 && len(browserChecks) == 0 {
		failures = append(failures, "bugfix_requires_independent_or_browser_check")
	}
	for _, path := range independentChecks {
		if privateEvidencePathDenied(path) {
			failures = append(failures, "independent_check_private_path_denied="+pathPrivacyRef(workspace, path))
		} else if !nonEmpty(path) {
			failures = append(failures, "independent_check_missing_or_too_small="+pathPrivacyRef(workspace, path))
		}
	}
	for _, path := range browserChecks {
		if privateEvidencePathDenied(path) {
			failures = append(failures, "browser_check_private_path_denied="+pathPrivacyRef(workspace, path))
		} else if !nonEmpty(path) {
			failures = append(failures, "browser_check_missing_or_too_small="+pathPrivacyRef(workspace, path))
		}
	}
	for _, item := range stillFailing {
		if strings.TrimSpace(item) != "" {
			failures = append(failures, "still_failing_key="+privacyHash(item))
		}
	}

	report := jsonObject{
		"command":                    "bugfix-guard",
		"schema_version":             "bugfix-guard.v1",
		"workspace_key":              privacyHash(absClean(workspace)),
		"privacy_mode":               "hash-length-and-evidence-ref-only",
		"goal":                       textEvidenceSummary(goal),
		"repro":                      textEvidenceSummary(repro),
		"artifact_refs":              pathEvidenceRefs(workspace, artifacts),
		"verifications":              textEvidenceSummaries(verifications),
		"self_test_refs":             pathEvidenceRefs(workspace, selfTests),
		"independent_check_refs":     pathEvidenceRefs(workspace, independentChecks),
		"browser_check_refs":         pathEvidenceRefs(workspace, browserChecks),
		"still_failing":              textEvidenceSummaries(stillFailing),
		"root_cause_report_ref":      pathPrivacyRef(workspace, rootCauseReport),
		"root_cause_report_verified": rootCauseReportVerified,
		"status":                     "pass",
	}
	if len(failures) > 0 {
		report["status"] = "fail"
		report["failures"] = failures
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(workspace, "bugfix-guard.json")
	}
	if err := ensureDir(workspace); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := writeJSON(outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return printGate("bugfix-guard", failures)
}

func qualityGuardCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	goal, ok := argValue(args, "--goal")
	if !ok || strings.TrimSpace(goal) == "" {
		usage()
		return 2
	}
	artifacts := uniqueStrings(argValues(args, "--artifact"))
	verifications := uniqueStrings(argValues(args, "--verify"))
	browserChecks := uniqueStrings(argValues(args, "--browser-check"))
	programChecks := uniqueStrings(argValues(args, "--program-check"))
	commandChecks := uniqueStrings(argValues(args, "--command-check"))
	mcpChecks := uniqueStrings(argValues(args, "--mcp-check"))
	stillFailing := uniqueStrings(argValues(args, "--still-failing"))
	reportPath, hasReport := argValue(args, "--report")

	failures := []string{}
	if len(artifacts) == 0 {
		failures = append(failures, "quality_requires_artifact")
	}
	for _, artifact := range artifacts {
		if privateEvidencePathDenied(artifact) {
			failures = append(failures, "artifact_private_path_denied="+pathPrivacyRef(workspace, artifact))
		} else if strings.TrimSpace(workspace) != "" && !sameOrDescendant(artifact, workspace) {
			failures = append(failures, "artifact_outside_workspace="+pathPrivacyRef(workspace, artifact))
		} else if !nonEmpty(artifact) {
			failures = append(failures, "artifact_missing_or_too_small="+pathPrivacyRef(workspace, artifact))
		}
	}
	if len(verifications) == 0 {
		failures = append(failures, "quality_requires_verification")
	}
	totalChecks := len(browserChecks) + len(programChecks) + len(commandChecks) + len(mcpChecks)
	if totalChecks == 0 {
		failures = append(failures, "quality_requires_independent_check")
	}
	for _, path := range browserChecks {
		if privateEvidencePathDenied(path) {
			failures = append(failures, "browser_check_private_path_denied="+pathPrivacyRef(workspace, path))
		} else if !nonEmpty(path) {
			failures = append(failures, "browser_check_missing_or_too_small="+pathPrivacyRef(workspace, path))
		}
	}
	for _, path := range programChecks {
		if privateEvidencePathDenied(path) {
			failures = append(failures, "program_check_private_path_denied="+pathPrivacyRef(workspace, path))
		} else if !nonEmpty(path) {
			failures = append(failures, "program_check_missing_or_too_small="+pathPrivacyRef(workspace, path))
		}
	}
	for _, path := range commandChecks {
		if privateEvidencePathDenied(path) {
			failures = append(failures, "command_check_private_path_denied="+pathPrivacyRef(workspace, path))
		} else if !nonEmpty(path) {
			failures = append(failures, "command_check_missing_or_too_small="+pathPrivacyRef(workspace, path))
		}
	}
	for _, path := range mcpChecks {
		if privateEvidencePathDenied(path) {
			failures = append(failures, "mcp_check_private_path_denied="+pathPrivacyRef(workspace, path))
		} else if !nonEmpty(path) {
			failures = append(failures, "mcp_check_missing_or_too_small="+pathPrivacyRef(workspace, path))
		}
	}
	for _, item := range stillFailing {
		if strings.TrimSpace(item) != "" {
			failures = append(failures, "still_failing_key="+privacyHash(item))
		}
	}

	report := jsonObject{
		"command":            "quality-guard",
		"schema_version":     "quality-guard.v1",
		"workspace_key":      privacyHash(absClean(workspace)),
		"privacy_mode":       "hash-length-and-evidence-ref-only",
		"goal":               textEvidenceSummary(goal),
		"artifact_refs":      pathEvidenceRefs(workspace, artifacts),
		"verifications":      textEvidenceSummaries(verifications),
		"browser_check_refs": pathEvidenceRefs(workspace, browserChecks),
		"program_check_refs": pathEvidenceRefs(workspace, programChecks),
		"command_check_refs": pathEvidenceRefs(workspace, commandChecks),
		"mcp_check_refs":     pathEvidenceRefs(workspace, mcpChecks),
		"still_failing":      textEvidenceSummaries(stillFailing),
		"status":             "pass",
	}
	if len(failures) > 0 {
		report["status"] = "fail"
		report["failures"] = failures
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(workspace, "quality-guard.json")
	}
	if err := ensureDir(workspace); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := writeJSON(outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return printGate("quality-guard", failures)
}

func migrationGuardCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	goal, ok := argValue(args, "--goal")
	if !ok || strings.TrimSpace(goal) == "" {
		usage()
		return 2
	}
	featureMap, ok := argValue(args, "--feature-map")
	if !ok || strings.TrimSpace(featureMap) == "" {
		usage()
		return 2
	}
	artifacts := uniqueStrings(argValues(args, "--artifact"))
	verifications := uniqueStrings(argValues(args, "--verify"))
	runEvidence := uniqueStrings(argValues(args, "--run-evidence"))
	previewEvidence := uniqueStrings(argValues(args, "--preview-evidence"))
	missingFeatures := uniqueStrings(argValues(args, "--missing-feature"))
	fakePages := uniqueStrings(argValues(args, "--fake-page"))
	placeholderPages := uniqueStrings(argValues(args, "--placeholder-page"))
	reportPath, hasReport := argValue(args, "--report")

	failures := []string{}
	if privateEvidencePathDenied(featureMap) {
		failures = append(failures, "feature_map_private_path_denied="+pathPrivacyRef(workspace, featureMap))
	} else if !nonEmpty(featureMap) {
		failures = append(failures, "feature_map_missing_or_too_small="+pathPrivacyRef(workspace, featureMap))
	}
	if len(artifacts) == 0 {
		failures = append(failures, "migration_requires_primary_artifact")
	}
	for _, artifact := range artifacts {
		if privateEvidencePathDenied(artifact) {
			failures = append(failures, "artifact_private_path_denied="+pathPrivacyRef(workspace, artifact))
		} else if strings.TrimSpace(workspace) != "" && !sameOrDescendant(artifact, workspace) {
			failures = append(failures, "artifact_outside_workspace="+pathPrivacyRef(workspace, artifact))
		} else if !nonEmpty(artifact) {
			failures = append(failures, "artifact_missing_or_too_small="+pathPrivacyRef(workspace, artifact))
		}
	}
	if len(verifications) == 0 {
		failures = append(failures, "migration_requires_verification")
	}
	if len(runEvidence) == 0 {
		failures = append(failures, "migration_requires_run_evidence")
	}
	for _, path := range runEvidence {
		if privateEvidencePathDenied(path) {
			failures = append(failures, "run_evidence_private_path_denied="+pathPrivacyRef(workspace, path))
		} else if !nonEmpty(path) {
			failures = append(failures, "run_evidence_missing_or_too_small="+pathPrivacyRef(workspace, path))
		}
	}
	for _, path := range previewEvidence {
		if privateEvidencePathDenied(path) {
			failures = append(failures, "preview_evidence_private_path_denied="+pathPrivacyRef(workspace, path))
		} else if !nonEmpty(path) {
			failures = append(failures, "preview_evidence_missing_or_too_small="+pathPrivacyRef(workspace, path))
		}
	}
	for _, item := range missingFeatures {
		if strings.TrimSpace(item) != "" {
			failures = append(failures, "missing_feature_key="+privacyHash(item))
		}
	}
	for _, item := range fakePages {
		if strings.TrimSpace(item) != "" {
			failures = append(failures, "fake_page_key="+privacyHash(item))
		}
	}
	for _, item := range placeholderPages {
		if strings.TrimSpace(item) != "" {
			failures = append(failures, "placeholder_page_key="+privacyHash(item))
		}
	}

	report := jsonObject{
		"command":               "migration-guard",
		"schema_version":        "migration-guard.v1",
		"workspace_key":         privacyHash(absClean(workspace)),
		"privacy_mode":          "hash-length-and-evidence-ref-only",
		"goal":                  textEvidenceSummary(goal),
		"feature_map_ref":       pathPrivacyRef(workspace, featureMap),
		"artifact_refs":         pathEvidenceRefs(workspace, artifacts),
		"verifications":         textEvidenceSummaries(verifications),
		"run_evidence_refs":     pathEvidenceRefs(workspace, runEvidence),
		"preview_evidence_refs": pathEvidenceRefs(workspace, previewEvidence),
		"missing_features":      textEvidenceSummaries(missingFeatures),
		"fake_pages":            textEvidenceSummaries(fakePages),
		"placeholder_pages":     textEvidenceSummaries(placeholderPages),
		"status":                "pass",
	}
	if len(failures) > 0 {
		report["status"] = "fail"
		report["failures"] = failures
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(workspace, "migration-guard.json")
	}
	if err := ensureDir(workspace); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := writeJSON(outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return printGate("migration-guard", failures)
}

func rootCauseRequiredForCloseout(values ...string) bool {
	markers := []string{
		"bug", "debug bug", "fix bug", "bugfix", "bug fix", "root cause", "root-cause", "failure", "failing", "regression", "rca", "symptom", "reproduce", "repro",
		"patch debt", "temporary patch", "hotfix", "workaround", "bypass", "still failing",
		"根因", "复现", "回归", "失败", "故障", "补丁债", "临时补丁", "绕过", "仍然失败", "返工",
	}
	for _, value := range values {
		if len(markerHits(value, markers)) > 0 {
			return true
		}
	}
	return false
}

func closeoutCheckCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	goal, ok := argValue(args, "--goal")
	if !ok || strings.TrimSpace(goal) == "" {
		usage()
		return 2
	}
	artifacts := uniqueStrings(argValues(args, "--artifact"))
	verifications := uniqueStrings(argValues(args, "--verify"))
	nextGaps := uniqueStrings(argValues(args, "--next-gap"))
	auditWorkspace, _ := argValue(args, "--audit-workspace")
	rootCauseReport, _ := argValue(args, "--root-cause-report")
	rootCauseRequiredText, _ := argValue(args, "--root-cause-required")
	reportPath, hasReport := argValue(args, "--report")
	failures := []string{}
	if len(artifacts) == 0 {
		failures = append(failures, "closeout_requires_artifact")
	}
	for _, artifact := range artifacts {
		if privateEvidencePathDenied(artifact) {
			failures = append(failures, "artifact_private_path_denied="+pathPrivacyRef(workspace, artifact))
		} else if strings.TrimSpace(workspace) != "" && !sameOrDescendant(artifact, workspace) {
			failures = append(failures, "artifact_outside_workspace="+pathPrivacyRef(workspace, artifact))
		} else if !nonEmpty(artifact) {
			failures = append(failures, "artifact_missing_or_too_small="+pathPrivacyRef(workspace, artifact))
		}
	}
	if len(verifications) == 0 {
		failures = append(failures, "closeout_requires_verification")
	}
	needsDecisionText, _ := argValue(args, "--needs-user-decision")
	blockedReason, _ := argValue(args, "--blocked-reason")
	needsDecision := strings.EqualFold(strings.TrimSpace(needsDecisionText), "true")
	hasResolvedGap := false
	for _, gap := range nextGaps {
		trimmedGap := strings.TrimSpace(gap)
		if trimmedGap == "" {
			continue
		}
		if needsDecision || strings.TrimSpace(blockedReason) != "" {
			hasResolvedGap = true
			continue
		}
		failures = append(failures, "unfinished_goal_has_remaining_step_key="+privacyHash(trimmedGap))
	}
	if needsDecision && strings.TrimSpace(blockedReason) == "" {
		failures = append(failures, "needs_user_decision_requires_blocked_reason")
	}
	rootCauseTexts := append([]string{goal}, nextGaps...)
	rootCauseRequired := rootCauseRequiredForCloseout(rootCauseTexts...) || artifactReportsRequireRootCause(workspace, artifacts)
	if strings.TrimSpace(rootCauseRequiredText) != "" {
		forcedRootCause, err := parseBoolValue(rootCauseRequiredText)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		if forcedRootCause {
			rootCauseRequired = true
		}
	}
	rootCauseReportVerified := false
	if rootCauseRequired {
		if strings.TrimSpace(rootCauseReport) == "" {
			failures = append(failures, "closeout_requires_root_cause_report")
		} else {
			rootCauseFailures := rootCauseReportPassFailures(workspace, rootCauseReport)
			if len(rootCauseFailures) == 0 {
				rootCauseReportVerified = true
			}
			failures = append(failures, rootCauseFailures...)
		}
	}
	if !needsDecision {
		if strings.TrimSpace(auditWorkspace) == "" {
			failures = append(failures, "closeout_requires_audit_workspace")
		} else {
			failures = append(failures, auditReportPassFailures(auditWorkspace)...)
		}
	}
	runtimeContextAuditRequired := runtimeTokenAuditRequiredForTexts(append(append(append([]string{goal, blockedReason}, verifications...), nextGaps...), rootCauseTexts...)...)
	runtimeContextAuditVerified := false
	if !needsDecision && runtimeContextAuditRequired {
		auditTarget := auditWorkspace
		if strings.TrimSpace(auditTarget) == "" {
			auditTarget = workspace
		}
		runtimeFailures := tokenOptimizationAuditPassFailures(auditTarget)
		if len(runtimeFailures) == 0 {
			runtimeContextAuditVerified = true
		}
		failures = append(failures, runtimeFailures...)
	}
	report := jsonObject{
		"command":                        "closeout-check",
		"schema_version":                 "closeout-check.v1",
		"workspace_key":                  privacyHash(absClean(workspace)),
		"privacy_mode":                   "hash-length-and-evidence-ref-only",
		"goal":                           textEvidenceSummary(goal),
		"artifact_refs":                  pathEvidenceRefs(workspace, artifacts),
		"verifications":                  textEvidenceSummaries(verifications),
		"remaining_gaps":                 textEvidenceSummaries(nextGaps),
		"audit_workspace_key":            privacyHash(absClean(auditWorkspace)),
		"root_cause_required":            rootCauseRequired,
		"root_cause_report":              pathPrivacyRef(workspace, rootCauseReport),
		"root_cause_report_verified":     rootCauseReportVerified,
		"needs_user_decision":            needsDecision,
		"blocked_reason":                 textEvidenceSummary(blockedReason),
		"resolved_gap_mode":              hasResolvedGap,
		"runtime_context_audit_required": runtimeContextAuditRequired,
		"runtime_context_audit_verified": runtimeContextAuditVerified,
		"status":                         "pass",
	}
	if len(failures) > 0 {
		report["status"] = "fail"
		report["failures"] = failures
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(workspace, "closeout-check.json")
	}
	if err := writeJSON(outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return printGate("closeout-check", failures)
}

func finishOrBlockCommand(args []string) int {
	workspace, _ := argValue(args, "--workspace")
	goal, ok := argValue(args, "--goal")
	if !ok || strings.TrimSpace(goal) == "" {
		usage()
		return 2
	}
	remainingSteps := uniqueStrings(argValues(args, "--remaining-step"))
	needsDecisionText, _ := argValue(args, "--needs-user-decision")
	blockedReason, _ := argValue(args, "--blocked-reason")
	auditWorkspace, _ := argValue(args, "--audit-workspace")
	rootCauseReport, _ := argValue(args, "--root-cause-report")
	rootCauseRequiredText, _ := argValue(args, "--root-cause-required")
	reportPath, hasReport := argValue(args, "--report")
	needsDecision := strings.EqualFold(strings.TrimSpace(needsDecisionText), "true")
	failures := []string{}
	for _, step := range remainingSteps {
		if strings.TrimSpace(step) == "" {
			continue
		}
		if !needsDecision && strings.TrimSpace(blockedReason) == "" {
			failures = append(failures, "unfinished_goal_has_remaining_step_key="+privacyHash(step))
		}
	}
	if needsDecision && strings.TrimSpace(blockedReason) == "" {
		failures = append(failures, "needs_user_decision_requires_blocked_reason")
	}
	rootCauseTexts := append([]string{goal}, remainingSteps...)
	rootCauseRequired := rootCauseRequiredForCloseout(rootCauseTexts...)
	if strings.TrimSpace(rootCauseRequiredText) != "" {
		forcedRootCause, err := parseBoolValue(rootCauseRequiredText)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		if forcedRootCause {
			rootCauseRequired = true
		}
	}
	rootCauseReportVerified := false
	if rootCauseRequired {
		if strings.TrimSpace(rootCauseReport) == "" {
			failures = append(failures, "finish_requires_root_cause_report")
		} else if strings.TrimSpace(workspace) == "" {
			failures = append(failures, "finish_requires_workspace_for_root_cause_report")
		} else {
			rootCauseFailures := rootCauseReportPassFailures(workspace, rootCauseReport)
			if len(rootCauseFailures) == 0 {
				rootCauseReportVerified = true
			}
			failures = append(failures, rootCauseFailures...)
		}
	}
	if len(remainingSteps) == 0 && !needsDecision {
		if strings.TrimSpace(auditWorkspace) == "" {
			failures = append(failures, "finish_requires_audit_workspace")
		} else {
			failures = append(failures, auditReportPassFailures(auditWorkspace)...)
		}
	}
	runtimeContextAuditRequired := runtimeTokenAuditRequiredForTexts(append([]string{goal, blockedReason}, remainingSteps...)...)
	runtimeContextAuditVerified := false
	if len(remainingSteps) == 0 && !needsDecision && runtimeContextAuditRequired {
		auditTarget := auditWorkspace
		if strings.TrimSpace(auditTarget) == "" {
			auditTarget = workspace
		}
		runtimeFailures := tokenOptimizationAuditPassFailures(auditTarget)
		if len(runtimeFailures) == 0 {
			runtimeContextAuditVerified = true
		}
		failures = append(failures, runtimeFailures...)
	}
	report := jsonObject{
		"command":                        "finish-or-block",
		"schema_version":                 "finish-or-block.v1",
		"workspace_key":                  privacyHash(absClean(workspace)),
		"goal":                           textEvidenceSummary(goal),
		"remaining_steps":                textEvidenceSummaries(remainingSteps),
		"audit_workspace_key":            privacyHash(absClean(auditWorkspace)),
		"root_cause_required":            rootCauseRequired,
		"root_cause_report":              pathPrivacyRef(workspace, rootCauseReport),
		"root_cause_report_verified":     rootCauseReportVerified,
		"needs_user_decision":            needsDecision,
		"blocked_reason":                 textEvidenceSummary(blockedReason),
		"runtime_context_audit_required": runtimeContextAuditRequired,
		"runtime_context_audit_verified": runtimeContextAuditVerified,
		"status":                         "pass",
	}
	if len(failures) > 0 {
		report["status"] = "fail"
		report["failures"] = failures
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(os.TempDir(), fmt.Sprintf("wuji-finish-or-block-%d.json", time.Now().UnixNano()))
	}
	if err := writeJSON(outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return printGate("finish-or-block", failures)
}

func repeatCandidatesCommand(args []string) int {
	logPath, ok := argValue(args, "--log")
	if !ok {
		usage()
		return 2
	}
	minOccurrences := 2
	if value, ok, err := parseIntArg(args, "--min-occurrences"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	} else if ok {
		minOccurrences = value
	}
	if minOccurrences < 2 {
		fmt.Fprintln(os.Stderr, "min-occurrences must be >= 2")
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	records, err := loadJSONLines(logPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	counts := map[string]int{}
	examples := map[string]string{}
	preferCounts := map[string]int{}
	avoidCounts := map[string]int{}
	for _, record := range records {
		task := objectString(record, "task_key")
		if task == "" {
			task = objectString(record, "task")
			if task == "" {
				continue
			}
		}
		counts[task]++
		if examples[task] == "" {
			if notePresent, ok := boolFromAny(record["note_present"]); ok && notePresent {
				examples[task] = "present"
			} else {
				examples[task] = "none"
			}
		}
		preferCounts[task] += len(uniqueStrings(stringSlice(record, "prefer_signal_keys")))
		avoidCounts[task] += len(uniqueStrings(stringSlice(record, "avoid_signal_keys")))
	}
	candidates := []jsonObject{}
	for task, count := range counts {
		if count >= minOccurrences {
			classification, reason := classifyStrategyResidency(task, count, preferCounts[task], avoidCounts[task])
			candidates = append(candidates, jsonObject{
				"task_key":              task,
				"occurrences":           count,
				"example_note_label":    examples[task],
				"recommended_sink":      "cli-or-skill",
				"classification":        classification,
				"classification_reason": reason,
			})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		left := candidates[i]["occurrences"].(int)
		right := candidates[j]["occurrences"].(int)
		if left == right {
			return fmt.Sprint(candidates[i]["task_key"]) < fmt.Sprint(candidates[j]["task_key"])
		}
		return left > right
	})
	distillQueue := []jsonObject{}
	for _, candidate := range candidates {
		distillQueue = append(distillQueue, jsonObject{
			"task_key":       candidate["task_key"],
			"occurrences":    candidate["occurrences"],
			"action":         "evaluate_for_distillation",
			"target":         candidate["recommended_sink"],
			"classification": candidate["classification"],
			"evidence_level": "checked",
		})
	}
	report := jsonObject{
		"log":             absClean(logPath),
		"min_occurrences": minOccurrences,
		"candidates":      candidates,
		"distill_queue":   distillQueue,
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(filepath.Dir(logPath), "repeat-candidates.json")
	}
	if err := writeJSON(outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	distillQueuePath := filepath.Join(filepath.Dir(logPath), "distill-queue.json")
	if err := writeJSON(distillQueuePath, jsonObject{
		"source_log":     absClean(logPath),
		"generated_at":   time.Now().UTC().Format(time.RFC3339),
		"evidence_level": "checked",
		"queue":          distillQueue,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("GO repeat-candidates\n- report=%s\n", absClean(outputPath))
	return 0
}

func evidenceGradeCommand(args []string) int {
	status, ok := argValue(args, "--status")
	if !ok {
		usage()
		return 2
	}
	summary, ok := argValue(args, "--summary")
	if !ok || strings.TrimSpace(summary) == "" {
		usage()
		return 2
	}
	allowed := map[string]bool{
		"candidate": true,
		"checked":   true,
		"verified":  true,
		"shipped":   true,
	}
	if !allowed[status] {
		fmt.Fprintln(os.Stderr, "status must be candidate, checked, verified, or shipped")
		return 2
	}
	artifacts := uniqueStrings(argValues(args, "--artifact"))
	workspace, _ := argValue(args, "--workspace")
	reportPath, hasReport := argValue(args, "--report")
	if strings.TrimSpace(workspace) == "" {
		if hasReport {
			workspace = filepath.Dir(absClean(reportPath))
		} else {
			workspace = "."
		}
	}
	failures := []string{}
	if status == "verified" || status == "shipped" {
		if len(artifacts) == 0 {
			failures = append(failures, "verified_or_shipped_requires_artifact")
		}
	}
	for _, artifact := range artifacts {
		if privateEvidencePathDenied(artifact) {
			failures = append(failures, "artifact_private_path_denied="+pathPrivacyRef(workspace, artifact))
		} else if strings.TrimSpace(workspace) != "" && !sameOrDescendant(artifact, workspace) {
			failures = append(failures, "artifact_outside_workspace="+pathPrivacyRef(workspace, artifact))
		} else if !nonEmpty(artifact) {
			failures = append(failures, "artifact_missing_or_too_small="+pathPrivacyRef(workspace, artifact))
		}
	}
	report := jsonObject{
		"command":        "evidence-grade",
		"schema_version": "evidence-grade.v1",
		"workspace_key":  privacyHash(absClean(workspace)),
		"privacy_mode":   "hash-length-and-evidence-ref-only",
		"generated_at":   time.Now().UTC().Format(time.RFC3339),
		"status":         status,
		"summary":        textEvidenceSummary(summary),
		"artifact_refs":  pathEvidenceRefs(workspace, artifacts),
	}
	if len(failures) > 0 {
		report["failures"] = failures
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(".", "evidence-grade.json")
	}
	if err := writeJSON(outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return printGate("evidence-grade", failures)
}

func truthStateCommand(args []string) int {
	text, ok := argValue(args, "--text")
	if !ok || strings.TrimSpace(text) == "" {
		usage()
		return 2
	}
	state, ok := argValue(args, "--state")
	if !ok {
		usage()
		return 2
	}
	state = strings.ToLower(strings.TrimSpace(state))
	allowedStates := map[string]bool{
		"fact":      true,
		"inference": true,
		"todo":      true,
	}
	if !allowedStates[state] {
		fmt.Fprintln(os.Stderr, "state must be fact, inference, or todo")
		return 2
	}
	evidence := uniqueStrings(argValues(args, "--evidence"))
	reportPath, hasReport := argValue(args, "--report")
	failures := []string{}
	makesSuccessClaim := claimTextHasSuccessMarker(text)
	hasUncertainty := len(markerHits(text, uncertaintyMarkers)) > 0
	switch state {
	case "fact":
		if len(evidence) == 0 {
			failures = append(failures, "fact_requires_evidence")
		}
		if hasUncertainty {
			failures = append(failures, "fact_must_not_include_uncertainty_markers")
		}
	case "inference":
		if makesSuccessClaim {
			failures = append(failures, "inference_must_not_make_success_claim")
		}
		if !hasUncertainty {
			failures = append(failures, "inference_requires_uncertainty_marker")
		}
	case "todo":
		if makesSuccessClaim {
			failures = append(failures, "todo_must_not_make_success_claim")
		}
	}
	for _, path := range evidence {
		if !nonEmpty(path) {
			failures = append(failures, "evidence_missing_or_too_small="+path)
		}
	}
	report := jsonObject{
		"text":            strings.TrimSpace(text),
		"state":           state,
		"evidence":        evidence,
		"success_claim":   makesSuccessClaim,
		"has_uncertainty": hasUncertainty,
		"status":          "pass",
	}
	if len(failures) > 0 {
		report["status"] = "fail"
		report["failures"] = failures
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(os.TempDir(), fmt.Sprintf("wuji-truth-state-%d.json", time.Now().UnixNano()))
	}
	if err := writeJSON(outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return printGate("truth-state", failures)
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
		".md": true, ".ps1": true, ".py": true, ".go": true, ".json": true, ".jsonl": true, ".toml": true, ".yaml": true, ".yml": true, ".txt": true,
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
		markerText := auditMarkerText(text)
		rel, _ := filepath.Rel(root, path)
		allowMarkerChecks := !isAuditMarkerFixture(rel)
		check := func(kind string, pattern string) {
			findings = append(findings, jsonObject{"file": rel, "kind": kind, "pattern": pattern})
			failures = append(failures, fmt.Sprintf("%s=%s", kind, rel))
		}
		if strings.EqualFold(filepath.Base(path), "task-log.jsonl") {
			records, err := loadJSONLines(path)
			if err != nil {
				check("task_log_unreadable", "task log unreadable")
				return nil
			}
			for _, failure := range taskLogExecutionRhythmFailures(records) {
				check("execution_rhythm_violation", failure)
			}
			for _, failure := range taskLogCloseoutLeakFailures(records) {
				check("closeout_leak_violation", failure)
			}
			for _, failure := range taskLogBlockedWaitFailures(records) {
				check("blocked_wait_violation", failure)
			}
		}
		replacementChar := string(rune(0xfffd))
		if strings.Contains(text, replacementChar) {
			check("encoding_replacement_char", "replacement-char")
		}
		staleRefs := []string{"units/" + "rust" + ".md", "experts/" + "rust/", "Rust" + "师", "Rust" + "主帅"}
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
		if allowMarkerChecks {
			unfinishedRefs := []string{"待" + "开发", "后续" + "路线"}
			for _, unfinishedRef := range unfinishedRefs {
				if strings.Contains(markerText, unfinishedRef) {
					check("unfinished_marker", "unfinished marker")
					break
				}
			}
			stackingRefs := []string{"A" + "/B", "a" + "/b", "并行" + "主线"}
			for _, stackingRef := range stackingRefs {
				if strings.Contains(markerText, stackingRef) {
					check("stacking_marker", "stacking marker")
					break
				}
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
		if allowMarkerChecks {
			incompleteRefs := []string{"to" + "do", "t" + "bd"}
			lowerText := strings.ToLower(markerText)
			for _, incompleteRef := range incompleteRefs {
				if strings.Contains(lowerText, incompleteRef) {
					check("incomplete_marker", "incomplete marker")
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if hasReport {
		_ = writeJSON(report, jsonObject{"path_ref": pathPrivacyRef(root, root), "workspace_key": privacyHash(absClean(root)), "findings": findings})
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
	logDir := workspace
	if value, ok := argValue(args, "--log-dir"); ok {
		logDir = value
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
	cachedTokens, _, err := parseIntArg(args, "--cached-tokens")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	freshInputTokens, _, err := parseIntArg(args, "--fresh-input-tokens")
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
	var qualityPass any = nil
	qualityValue, hasQualityValue := argValue(args, "--quality-pass")
	if !hasQualityValue {
		qualityValue, hasQualityValue = argValue(args, "--qa-pass")
	}
	if hasQualityValue {
		parsed, err := parseBoolValue(qualityValue)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		qualityPass = parsed
	}
	cacheHit, hasCacheHit := false, false
	if cacheValue, ok := argValue(args, "--cache-hit"); ok {
		parsed, err := parseBoolValue(cacheValue)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		cacheHit, hasCacheHit = parsed, true
	}
	reusedPrefixBytes, _, err := parseIntArg(args, "--reused-prefix-bytes")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	firstTokenMs, _, err := parseIntArg(args, "--first-token-ms")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	parallelFanout, _, err := parseIntArg(args, "--parallel-fanout")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	joinWaitMs, _, err := parseIntArg(args, "--join-wait-ms")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	mergeCostMs, _, err := parseIntArg(args, "--merge-cost-ms")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	activatedOfficers, _, err := parseIntArg(args, "--activated-officers")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	activatedSkills, _, err := parseIntArg(args, "--activated-skills")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	loadedFileBytes, _, err := parseIntArg(args, "--loaded-file-bytes")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	largestContextSegmentBytes, _, err := parseIntArg(args, "--largest-context-segment-bytes")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	prefixHash, _ := argValue(args, "--prefix-hash")
	routeID, _ := argValue(args, "--route-id")
	providerID, _ := argValue(args, "--provider-id")
	model, _ := argValue(args, "--model")
	if containsSecretLikeContent(name) || containsSecretLikeContent(routeID) || containsSecretLikeContent(providerID) || containsSecretLikeContent(model) {
		fmt.Fprintln(os.Stderr, "bench contains secret-like content")
		return 1
	}
	workspaceKey := privacyHash(absClean(workspace))
	entry := jsonObject{
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"workspace_key": workspaceKey,
		"name_key":      privacyHash(name),
		"input_tokens":  inputTokens,
		"output_tokens": outputTokens,
		"duration_ms":   durationMs,
		"tool_calls":    toolCalls,
		"retries":       retries,
		"quality_pass":  qualityPass,
	}
	if hasCacheHit {
		entry["cache_hit"] = cacheHit
	}
	if reusedPrefixBytes > 0 {
		entry["reused_prefix_bytes"] = reusedPrefixBytes
	}
	if cachedTokens > 0 {
		entry["cached_tokens"] = cachedTokens
	}
	if freshInputTokens > 0 {
		entry["fresh_input_tokens"] = freshInputTokens
	}
	if firstTokenMs > 0 {
		entry["first_token_ms"] = firstTokenMs
	}
	if parallelFanout > 0 {
		entry["parallel_fanout"] = parallelFanout
	}
	if joinWaitMs > 0 {
		entry["join_wait_ms"] = joinWaitMs
	}
	if mergeCostMs > 0 {
		entry["merge_cost_ms"] = mergeCostMs
	}
	if activatedOfficers > 0 {
		entry["activated_officers"] = activatedOfficers
	}
	if activatedSkills > 0 {
		entry["activated_skills"] = activatedSkills
	}
	if loadedFileBytes > 0 {
		entry["loaded_file_bytes"] = loadedFileBytes
	}
	if largestContextSegmentBytes > 0 {
		entry["largest_context_segment_bytes"] = largestContextSegmentBytes
	}
	if prefixHash != "" {
		entry["prefix_hash"] = prefixHash
	}
	if routeID != "" {
		entry["route_id"] = routeID
	}
	if providerID != "" {
		entry["provider_id"] = providerID
	}
	if model != "" {
		entry["model"] = model
	}
	logPath := filepath.Join(logDir, "bench.jsonl")
	if err := appendJSONLine(logPath, entry); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("GO bench\n- log=%s\n", pathPrivacyRef(workspace, logPath))
	return 0
}

func benchReportCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	logDir := workspace
	if value, ok := argValue(args, "--log-dir"); ok {
		logDir = value
	}
	reportPath, hasReport := argValue(args, "--report")
	logPath := filepath.Join(logDir, "bench.jsonl")
	bytes, err := os.ReadFile(logPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	logHash, err := fileSHA256(logPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	lines := strings.Split(strings.TrimSpace(string(bytes)), "\n")
	workspaceKey := privacyHash(absClean(workspace))
	totalRuns := 0
	totalInput := 0
	totalOutput := 0
	totalDuration := 0
	totalToolCalls := 0
	totalRetries := 0
	qualityPassCount := 0
	qualitySeen := 0
	cacheSeen := 0
	cacheHitCount := 0
	totalReusedPrefixBytes := 0
	totalFirstTokenMs := 0
	firstTokenSeen := 0
	totalParallelFanout := 0
	totalJoinWaitMs := 0
	totalMergeCostMs := 0
	totalActivatedOfficers := 0
	totalActivatedSkills := 0
	totalLoadedFileBytes := 0
	inputTokenValues := []int{}
	outputTokenValues := []int{}
	cachedTokenValues := []int{}
	freshInputTokenValues := []int{}
	uncachedTokenValues := []int{}
	reusedPrefixByteValues := []int{}
	activatedOfficerValues := []int{}
	activatedSkillValues := []int{}
	loadedFileByteValues := []int{}
	largestContextSegmentValues := []int{}
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
		if rowWorkspaceKey := objectString(row, "workspace_key"); rowWorkspaceKey != "" && rowWorkspaceKey != workspaceKey {
			fmt.Fprintln(os.Stderr, "bench log workspace_key mismatch")
			return 1
		}
		totalRuns++
		rowInput, hasRowInput := intFromAny(row["input_tokens"])
		rowOutput, hasRowOutput := intFromAny(row["output_tokens"])
		rowCached, hasRowCached := intFromAny(row["cached_tokens"])
		rowFresh, hasRowFresh := intFromAny(row["fresh_input_tokens"])
		if hasRowInput {
			totalInput += rowInput
			inputTokenValues = append(inputTokenValues, rowInput)
		}
		if hasRowOutput {
			totalOutput += rowOutput
			outputTokenValues = append(outputTokenValues, rowOutput)
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
		if value, ok := boolFromAny(row["quality_pass"]); ok {
			qualitySeen++
			if value {
				qualityPassCount++
			}
		} else if value, ok := boolFromAny(row["qa_pass"]); ok {
			qualitySeen++
			if value {
				qualityPassCount++
			}
		}
		if value, ok := boolFromAny(row["cache_hit"]); ok {
			cacheSeen++
			if value {
				cacheHitCount++
			}
		}
		if value, ok := intFromAny(row["reused_prefix_bytes"]); ok {
			totalReusedPrefixBytes += value
			reusedPrefixByteValues = append(reusedPrefixByteValues, value)
		}
		if hasRowCached {
			cachedTokenValues = append(cachedTokenValues, rowCached)
		}
		if hasRowFresh {
			freshInputTokenValues = append(freshInputTokenValues, rowFresh)
			uncachedTokenValues = append(uncachedTokenValues, rowFresh)
		} else if hasRowInput && hasRowCached {
			uncachedTokenValues = append(uncachedTokenValues, maxInt([]int{rowInput - rowCached, 0}))
		} else if hasRowInput {
			uncachedTokenValues = append(uncachedTokenValues, rowInput)
		}
		if value, ok := intFromAny(row["first_token_ms"]); ok {
			firstTokenSeen++
			totalFirstTokenMs += value
		}
		if value, ok := intFromAny(row["parallel_fanout"]); ok {
			totalParallelFanout += value
		}
		if value, ok := intFromAny(row["join_wait_ms"]); ok {
			totalJoinWaitMs += value
		}
		if value, ok := intFromAny(row["merge_cost_ms"]); ok {
			totalMergeCostMs += value
		}
		if value, ok := intFromAny(row["activated_officers"]); ok {
			totalActivatedOfficers += value
			activatedOfficerValues = append(activatedOfficerValues, value)
		}
		if value, ok := intFromAny(row["activated_skills"]); ok {
			totalActivatedSkills += value
			activatedSkillValues = append(activatedSkillValues, value)
		}
		if value, ok := intFromAny(row["loaded_file_bytes"]); ok {
			totalLoadedFileBytes += value
			loadedFileByteValues = append(loadedFileByteValues, value)
		}
		if value, ok := intFromAny(row["largest_context_segment_bytes"]); ok {
			largestContextSegmentValues = append(largestContextSegmentValues, value)
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
	qualityPassRate := 0.0
	if qualitySeen > 0 {
		qualityPassRate = float64(qualityPassCount) / float64(qualitySeen)
	}
	cacheHitRate := 0.0
	if cacheSeen > 0 {
		cacheHitRate = float64(cacheHitCount) / float64(cacheSeen)
	}
	avgFirstTokenMs := 0
	if firstTokenSeen > 0 {
		avgFirstTokenMs = totalFirstTokenMs / firstTokenSeen
	}
	report := jsonObject{
		"workspace_key":                     workspaceKey,
		"command":                           "bench-report",
		"generated_at":                      time.Now().UTC().Format(time.RFC3339),
		"wuji_version":                      builtinIronRulesVersion,
		"log_ref":                           pathPrivacyRef(workspace, logPath),
		"log_sha256":                        logHash,
		"input_hashes":                      jsonObject{"bench.jsonl": logHash},
		"runs":                              totalRuns,
		"total_input":                       totalInput,
		"total_output":                      totalOutput,
		"total_tokens":                      totalTokens,
		"avg_duration_ms":                   avgDuration,
		"total_tool_calls":                  totalToolCalls,
		"total_retries":                     totalRetries,
		"quality_pass_rate":                 qualityPassRate,
		"cache_hit_rate":                    cacheHitRate,
		"cache_observations":                cacheSeen,
		"total_reused_prefix_bytes":         totalReusedPrefixBytes,
		"reused_prefix_bytes_avg":           0,
		"reused_prefix_bytes_p95":           percentileInt(reusedPrefixByteValues, 0.95),
		"reused_prefix_bytes_max":           maxInt(reusedPrefixByteValues),
		"cached_tokens_p95":                 percentileInt(cachedTokenValues, 0.95),
		"cached_tokens_max":                 maxInt(cachedTokenValues),
		"input_tokens_p95":                  percentileInt(inputTokenValues, 0.95),
		"output_tokens_p95":                 percentileInt(outputTokenValues, 0.95),
		"fresh_input_tokens_p95":            percentileInt(freshInputTokenValues, 0.95),
		"uncached_tokens_p95":               percentileInt(uncachedTokenValues, 0.95),
		"tokens_per_success":                0,
		"total_activated_officers":          totalActivatedOfficers,
		"activated_officers_p95":            percentileInt(activatedOfficerValues, 0.95),
		"activated_officers_max":            maxInt(activatedOfficerValues),
		"total_activated_skills":            totalActivatedSkills,
		"activated_skills_p95":              percentileInt(activatedSkillValues, 0.95),
		"activated_skills_max":              maxInt(activatedSkillValues),
		"total_loaded_file_bytes":           totalLoadedFileBytes,
		"loaded_file_bytes_p95":             percentileInt(loadedFileByteValues, 0.95),
		"loaded_file_bytes_max":             maxInt(loadedFileByteValues),
		"largest_context_segment_bytes_p95": percentileInt(largestContextSegmentValues, 0.95),
		"largest_context_segment_bytes_max": maxInt(largestContextSegmentValues),
		"avg_first_token_ms":                avgFirstTokenMs,
		"total_parallel_fanout":             totalParallelFanout,
		"total_join_wait_ms":                totalJoinWaitMs,
		"total_merge_cost_ms":               totalMergeCostMs,
		"tokens_per_minute":                 tokensPerMinute,
	}
	if len(reusedPrefixByteValues) > 0 {
		report["reused_prefix_bytes_avg"] = totalReusedPrefixBytes / len(reusedPrefixByteValues)
	}
	if qualityPassCount > 0 {
		report["tokens_per_success"] = totalTokens / qualityPassCount
	}
	decision := "defer"
	volumeTooLarge := int64(percentileInt(cachedTokenValues, 0.95)) > maxCachedPrefixBytesP95 ||
		int64(percentileInt(reusedPrefixByteValues, 0.95)) > maxCachedPrefixBytesP95 ||
		percentileInt(inputTokenValues, 0.95) > maxInputTokensP95 ||
		percentileInt(freshInputTokenValues, 0.95) > maxFreshInputTokensP95 ||
		percentileInt(outputTokenValues, 0.95) > maxOutputTokensP95 ||
		percentileInt(uncachedTokenValues, 0.95) > maxUncachedTokensP95 ||
		percentileInt(activatedOfficerValues, 0.95) > maxActivatedOfficers ||
		percentileInt(activatedSkillValues, 0.95) > maxActivatedSkills ||
		int64(percentileInt(loadedFileByteValues, 0.95)) > maxLoadedFileBytes ||
		int64(percentileInt(largestContextSegmentValues, 0.95)) > maxLargestContextSegmentBytes
	if qualitySeen > 0 && qualityPassRate < 0.5 {
		decision = "reject"
	} else if cacheSeen > 0 && cacheHitRate < 0.9 {
		decision = "reject"
	} else if volumeTooLarge {
		decision = "reject"
	} else if qualityPassRate >= 0.9 && totalRetries <= totalRuns && avgDuration > 0 && cacheSeen >= minMeasuredCacheObservations {
		decision = "absorb"
	}
	report["decision"] = decision
	report["evidence_level"] = evidenceLevelFromDecision(decision)
	report["volume_gate"] = ternaryStatus(!volumeTooLarge, "pass", "fail")
	if hasReport {
		if err := writeJSON(reportPath, report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	fmt.Printf("GO bench-report\n- runs=%d\n- tokens_per_minute=%d\n", totalRuns, tokensPerMinute)
	return 0
}

func contextBloatAuditCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	benchPath, hasBenchPath := argValue(args, "--bench-report")
	if !hasBenchPath {
		benchPath = filepath.Join(workspace, "outputs", "bench-report.json")
	}
	failures := []string{}
	warnings := []string{}

	manifestPath := filepath.Join(workspace, "hotpath-manifest.json")
	manifestObj, manifestErr := loadJSONObject(manifestPath)
	if manifestErr != nil {
		failures = append(failures, "hotpath_manifest_unreadable="+absClean(manifestPath))
	}
	residentBytes := int64(0)
	dynamicBytes := int64(0)
	if manifestErr == nil {
		if resident, ok := objectSlice(manifestObj, "resident"); ok {
			for index, raw := range resident {
				item, ok := raw.(map[string]any)
				if !ok {
					failures = append(failures, fmt.Sprintf("hotpath_resident_%d_invalid", index+1))
					continue
				}
				path := objectString(item, "path")
				maxBytes, _ := int64FromAny(item["max_bytes"])
				if path == "" {
					failures = append(failures, fmt.Sprintf("hotpath_resident_%d_missing_path", index+1))
					continue
				}
				bytes := fileSize(filepath.Join(workspace, filepath.FromSlash(path)))
				residentBytes += bytes
				if maxBytes > 0 && bytes > maxBytes {
					failures = append(failures, fmt.Sprintf("hotpath_resident_over_budget=%s:%d>%d", path, bytes, maxBytes))
				}
			}
		} else {
			failures = append(failures, "hotpath_resident_missing")
		}
		if onDemand, ok := objectSlice(manifestObj, "on_demand"); ok {
			for _, raw := range onDemand {
				if item, ok := raw.(map[string]any); ok {
					if maxBytes, ok := int64FromAny(item["max_loaded_bytes"]); ok {
						dynamicBytes += maxBytes
					}
				}
			}
		}
		if cold, ok := objectSlice(manifestObj, "cold_ledger"); ok {
			for index, raw := range cold {
				item, ok := raw.(map[string]any)
				if !ok {
					failures = append(failures, fmt.Sprintf("hotpath_cold_ledger_%d_invalid", index+1))
					continue
				}
				if objectString(item, "default_mode") != "handle-only" {
					failures = append(failures, "hotpath_cold_ledger_not_handle_only="+objectString(item, "path"))
				}
			}
		}
		if forbidden, ok := objectSlice(manifestObj, "forbidden_resident"); ok {
			if len(forbidden) == 0 {
				failures = append(failures, "hotpath_forbidden_resident_empty")
			}
		} else {
			failures = append(failures, "hotpath_forbidden_resident_missing")
		}
	}
	if residentBytes > maxHotpathResidentBytes {
		failures = append(failures, fmt.Sprintf("hotpath_resident_total_over_budget=%d>%d", residentBytes, maxHotpathResidentBytes))
	}
	if dynamicBytes > maxHotpathDynamicBytes {
		failures = append(failures, fmt.Sprintf("hotpath_dynamic_budget_over=%d>%d", dynamicBytes, maxHotpathDynamicBytes))
	}

	benchObj, benchErr := loadJSONObject(benchPath)
	benchStatus := "missing"
	if benchErr != nil {
		warnings = append(warnings, "bench_report_missing_or_unreadable="+absClean(benchPath))
	} else {
		benchStatus = "checked"
		benchRef := pathPrivacyRef(workspace, benchPath)
		expectedWorkspaceKey := privacyHash(absClean(workspace))
		if objectString(benchObj, "workspace_key") != expectedWorkspaceKey {
			failures = append(failures, "bench_report_workspace_key_mismatch="+benchRef)
		}
		if objectString(benchObj, "command") != "bench-report" {
			failures = append(failures, "bench_report_command_mismatch="+benchRef)
		}
		generatedAt := objectString(benchObj, "generated_at")
		if generatedAt == "" {
			failures = append(failures, "bench_report_missing_generated_at="+benchRef)
		} else if _, err := time.Parse(time.RFC3339, generatedAt); err != nil {
			failures = append(failures, "bench_report_bad_generated_at="+benchRef)
		}
		inputHashes, ok := objectMap(benchObj, "input_hashes")
		if !ok {
			failures = append(failures, "bench_report_missing_input_hashes="+benchRef)
		} else {
			expectedHash, ok := inputHashes["bench.jsonl"].(string)
			if !ok || strings.TrimSpace(expectedHash) == "" || strings.HasPrefix(expectedHash, "__") {
				failures = append(failures, "bench_report_bad_input_hash=bench.jsonl")
			} else {
				logRef := objectString(benchObj, "log_ref")
				logPath := ""
				if logRef != "" && !strings.HasPrefix(logRef, "external:") {
					logPath = resolveWorkspaceRef(workspace, logRef)
				}
				if logPath == "" {
					logPath = filepath.Join(filepath.Dir(benchPath), "bench.jsonl")
				}
				currentHash, err := fileSHA256(logPath)
				if err != nil {
					failures = append(failures, "bench_report_input_unreadable=bench.jsonl")
				} else if currentHash != expectedHash {
					failures = append(failures, "bench_report_input_hash_mismatch=bench.jsonl")
				}
			}
		}
		cacheObs, _ := intFromAny(benchObj["cache_observations"])
		if cacheObs < minMeasuredCacheObservations {
			warnings = append(warnings, fmt.Sprintf("cache_observations_below_measurement_floor=%d<%d", cacheObs, minMeasuredCacheObservations))
		}
		if value, ok := intFromAny(benchObj["cached_tokens_p95"]); ok && int64(value) > maxCachedPrefixBytesP95 {
			failures = append(failures, fmt.Sprintf("cached_tokens_p95_over_budget=%d>%d", value, maxCachedPrefixBytesP95))
		}
		if value, ok := intFromAny(benchObj["reused_prefix_bytes_p95"]); ok && int64(value) > maxCachedPrefixBytesP95 {
			failures = append(failures, fmt.Sprintf("reused_prefix_bytes_p95_over_budget=%d>%d", value, maxCachedPrefixBytesP95))
		}
		if value, ok := intFromAny(benchObj["input_tokens_p95"]); ok && value > maxInputTokensP95 {
			failures = append(failures, fmt.Sprintf("input_tokens_p95_over_budget=%d>%d", value, maxInputTokensP95))
		}
		if value, ok := intFromAny(benchObj["fresh_input_tokens_p95"]); ok && value > maxFreshInputTokensP95 {
			failures = append(failures, fmt.Sprintf("fresh_input_tokens_p95_over_budget=%d>%d", value, maxFreshInputTokensP95))
		}
		if value, ok := intFromAny(benchObj["output_tokens_p95"]); ok && value > maxOutputTokensP95 {
			failures = append(failures, fmt.Sprintf("output_tokens_p95_over_budget=%d>%d", value, maxOutputTokensP95))
		}
		if value, ok := intFromAny(benchObj["uncached_tokens_p95"]); ok && value > maxUncachedTokensP95 {
			failures = append(failures, fmt.Sprintf("uncached_tokens_p95_over_budget=%d>%d", value, maxUncachedTokensP95))
		}
		if value, ok := intFromAny(benchObj["activated_officers_p95"]); ok && value > maxActivatedOfficers {
			failures = append(failures, fmt.Sprintf("activated_officers_p95_over_budget=%d>%d", value, maxActivatedOfficers))
		}
		if value, ok := intFromAny(benchObj["activated_skills_p95"]); ok && value > maxActivatedSkills {
			failures = append(failures, fmt.Sprintf("activated_skills_p95_over_budget=%d>%d", value, maxActivatedSkills))
		}
		if value, ok := intFromAny(benchObj["loaded_file_bytes_p95"]); ok && int64(value) > maxLoadedFileBytes {
			failures = append(failures, fmt.Sprintf("loaded_file_bytes_p95_over_budget=%d>%d", value, maxLoadedFileBytes))
		}
		if value, ok := intFromAny(benchObj["largest_context_segment_bytes_p95"]); ok && int64(value) > maxLargestContextSegmentBytes {
			failures = append(failures, fmt.Sprintf("largest_context_segment_bytes_p95_over_budget=%d>%d", value, maxLargestContextSegmentBytes))
		}
		if objectString(benchObj, "volume_gate") == "fail" {
			failures = append(failures, "bench_volume_gate_failed")
		}
		if objectString(benchObj, "decision") == "reject" {
			failures = append(failures, "bench_decision_reject")
		}
	}

	report := jsonObject{
		"workspace_key": privacyHash(absClean(workspace)),
		"command":       "context-bloat-audit",
		"checks": []string{
			"hotpath-manifest-present",
			"resident-budget",
			"cold-ledger-handle-only",
			"bench-cache-volume",
			"moe-activation-budget",
		},
		"hotpath_manifest": pathPrivacyRef(workspace, manifestPath),
		"bench_report":     pathPrivacyRef(workspace, benchPath),
		"bench_status":     benchStatus,
		"budgets": jsonObject{
			"resident_bytes_max":                    maxHotpathResidentBytes,
			"resident_bytes":                        residentBytes,
			"dynamic_bytes_max":                     maxHotpathDynamicBytes,
			"dynamic_declared_bytes":                dynamicBytes,
			"cached_prefix_bytes_p95_max":           maxCachedPrefixBytesP95,
			"input_tokens_p95_max":                  maxInputTokensP95,
			"fresh_input_tokens_p95_max":            maxFreshInputTokensP95,
			"output_tokens_p95_max":                 maxOutputTokensP95,
			"uncached_tokens_p95_max":               maxUncachedTokensP95,
			"activated_officers_p95_max":            maxActivatedOfficers,
			"activated_skills_p95_max":              maxActivatedSkills,
			"loaded_file_bytes_p95_max":             maxLoadedFileBytes,
			"largest_context_segment_bytes_p95_max": maxLargestContextSegmentBytes,
			"min_measured_cache_observations":       minMeasuredCacheObservations,
		},
		"warnings": warnings,
		"status":   ternaryStatus(len(failures) == 0, "pass", "fail"),
	}
	manifest := auditManifest(workspace, "context-bloat-audit", []string{
		"hotpath-manifest.json",
		"outputs/context-pack-rich.json",
		"outputs/bench-report.json",
		"tools/wuji_cli.go",
	})
	for key, value := range manifest {
		report[key] = value
	}
	if len(failures) > 0 {
		report["failures"] = failures
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(workspace, "outputs", "context-bloat-audit-report.json")
	}
	if err := writeJSON(outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return printGate("context-bloat-audit", failures)
}

func nestedIntFromMap(obj map[string]any, parent string, child string) (int, bool) {
	parentObj, ok := objectMap(obj, parent)
	if !ok {
		return 0, false
	}
	return intFromAny(parentObj[child])
}

func runtimeUsageRawPayloadFailures(record map[string]any, index int) []string {
	forbiddenKeys := map[string]bool{
		"prompt":        true,
		"messages":      true,
		"content":       true,
		"text":          true,
		"request":       true,
		"request_body":  true,
		"response":      true,
		"response_body": true,
		"body":          true,
		"raw":           true,
	}
	failures := []string{}
	for key, value := range record {
		lowerKey := strings.ToLower(strings.TrimSpace(key))
		if forbiddenKeys[lowerKey] {
			failures = append(failures, fmt.Sprintf("runtime_usage_record_%d_raw_payload_field_forbidden=%s", index, lowerKey))
			continue
		}
		switch typed := value.(type) {
		case string:
			if containsSecretLikeContent(typed) {
				failures = append(failures, fmt.Sprintf("runtime_usage_record_%d_secret_like_value_rejected=%s", index, lowerKey))
			}
		}
	}
	return failures
}

func contextSlimmingRecommendations(longContextSuspected bool, inputP95 int, cachedP95 int, freshP95 int, outputP95 int, uncachedP95 int) []string {
	actions := []string{}
	if longContextSuspected {
		actions = append(actions,
			"split-or-reset-unrelated-long-thread",
			"keep-only-small-stable-prefix-and-move-volatile-facts-late",
			"replace-long-history-with-task-state-summary-and-evidence-handles",
			"mount-one-owner-one-skill-and-triggered-officers-only",
			"summarize-tool-outputs-before-reuse-never-replay-logs",
		)
	}
	if inputP95 > maxInputTokensP95 || uncachedP95 > maxUncachedTokensP95 {
		actions = append(actions,
			"retrieve-key-ranges-instead-of-loading-whole-files",
			"use-source-cards-first-deep-read-only-survivors",
			"prefer-rag-or-handle-based-context-over-full-document-replay",
		)
	}
	if freshP95 > maxFreshInputTokensP95 {
		actions = append(actions, "move-repeated-instructions-into-stable-canon-or-delete-duplicates")
	}
	if outputP95 > maxOutputTokensP95 {
		actions = append(actions, "return-decision-and-diff-summary-not-process-transcript")
	}
	if cachedP95 > 0 && cachedP95 <= int(maxCachedPrefixBytesP95) && inputP95 <= maxInputTokensP95 && outputP95 <= maxOutputTokensP95 {
		actions = append(actions, "preserve-byte-stable-prefix-order-do-not-trade-cache-hit-for-larger-fresh-output")
	}
	return orderedUniqueStrings(actions)
}

func runtimeContextAuditCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	workspace = absClean(workspace)
	usageLog, hasUsageLog := argValue(args, "--usage-log")
	if !hasUsageLog {
		usageLog = filepath.Join(workspace, "outputs", "runtime-usage.jsonl")
	}
	reportPath, hasReport := argValue(args, "--report")
	if !hasReport {
		reportPath = filepath.Join(workspace, "outputs", "runtime-context-audit-report.json")
	}
	failures := []string{}
	warnings := []string{}
	if privateEvidencePathDenied(usageLog) {
		failures = append(failures, "runtime_usage_log_private_path_denied="+pathPrivacyRef(workspace, usageLog))
	} else if !sameOrDescendant(usageLog, workspace) {
		failures = append(failures, "runtime_usage_log_outside_workspace="+pathPrivacyRef(workspace, usageLog))
	}
	usageRecords := []map[string]any{}
	usageLogHash := ""
	if len(failures) == 0 {
		if !nonEmpty(usageLog) {
			failures = append(failures, "runtime_usage_log_missing_or_too_small="+pathPrivacyRef(workspace, usageLog))
		} else {
			records, err := loadJSONLines(usageLog)
			if err != nil {
				failures = append(failures, "runtime_usage_log_unreadable="+pathPrivacyRef(workspace, usageLog))
			} else {
				usageRecords = records
				if hash, err := fileSHA256(usageLog); err == nil {
					usageLogHash = hash
				}
			}
		}
	}

	inputValues := []int{}
	outputValues := []int{}
	cachedValues := []int{}
	freshValues := []int{}
	uncachedValues := []int{}
	cacheHits := 0
	workspaceKey := privacyHash(workspace)
	for index, record := range usageRecords {
		failures = append(failures, runtimeUsageRawPayloadFailures(record, index+1)...)
		if recordWorkspaceKey := objectString(record, "workspace_key"); recordWorkspaceKey != "" && recordWorkspaceKey != workspaceKey {
			failures = append(failures, fmt.Sprintf("runtime_usage_record_%d_workspace_key_mismatch", index+1))
		}
		usageObj := record
		if nested, ok := objectMap(record, "usage"); ok {
			usageObj = nested
		}
		inputTokens, hasInput := intFromKeys(usageObj, "input_tokens", "prompt_tokens", "total_input_tokens")
		outputTokens, hasOutput := intFromKeys(usageObj, "output_tokens", "completion_tokens", "total_output_tokens")
		cachedTokens, hasCached := intFromKeys(usageObj, "cached_tokens", "prompt_cached_tokens", "input_cached_tokens")
		if !hasCached {
			cachedTokens, hasCached = nestedIntFromMap(usageObj, "prompt_tokens_details", "cached_tokens")
		}
		if !hasCached {
			cachedTokens, hasCached = nestedIntFromMap(usageObj, "input_tokens_details", "cached_tokens")
		}
		freshTokens, hasFresh := intFromKeys(usageObj, "fresh_input_tokens", "uncached_input_tokens")
		if hasInput {
			inputValues = append(inputValues, inputTokens)
		}
		if hasOutput {
			outputValues = append(outputValues, outputTokens)
		}
		if hasCached {
			cachedValues = append(cachedValues, cachedTokens)
			if cachedTokens > 0 {
				cacheHits++
			}
		}
		if hasFresh {
			freshValues = append(freshValues, freshTokens)
			uncachedValues = append(uncachedValues, freshTokens)
		} else if hasInput && hasCached {
			uncachedValues = append(uncachedValues, maxInt([]int{inputTokens - cachedTokens, 0}))
		} else if hasInput {
			uncachedValues = append(uncachedValues, inputTokens)
		}
		if !hasInput || !hasOutput || !hasCached {
			failures = append(failures, fmt.Sprintf("runtime_usage_record_%d_missing_required_usage_metrics", index+1))
		}
	}

	usageObservations := len(usageRecords)
	if usageObservations < minRuntimeUsageObservations {
		failures = append(failures, fmt.Sprintf("runtime_usage_observations_below_floor=%d<%d", usageObservations, minRuntimeUsageObservations))
	}
	runtimeSkillPath := `C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md`
	runtimeSkillBytes := fileSize(runtimeSkillPath)
	repoInstructionBytes := fileSize(filepath.Join(workspace, "AGENTS.md"))
	mirrorBytes := fileSize(filepath.Join(workspace, "SKILL.md")) + fileSize(filepath.Join(workspace, "GLOBAL_AGENTS.md")) + fileSize(filepath.Join(workspace, "README.md"))
	if runtimeSkillBytes == 0 {
		failures = append(failures, "runtime_skill_missing_or_unreadable")
	} else if runtimeSkillBytes > maxRuntimeSkillBytes {
		failures = append(failures, fmt.Sprintf("runtime_skill_bytes_over_budget=%d>%d", runtimeSkillBytes, maxRuntimeSkillBytes))
	}
	if repoInstructionBytes > maxRuntimeRepoInstructionBytes {
		failures = append(failures, fmt.Sprintf("repo_instruction_bytes_over_budget=%d>%d", repoInstructionBytes, maxRuntimeRepoInstructionBytes))
	}
	if mirrorBytes > maxRuntimeMirrorBytes {
		failures = append(failures, fmt.Sprintf("runtime_mirror_bytes_over_budget=%d>%d", mirrorBytes, maxRuntimeMirrorBytes))
	}
	cachedP95 := percentileInt(cachedValues, 0.95)
	inputP95 := percentileInt(inputValues, 0.95)
	freshP95 := percentileInt(freshValues, 0.95)
	outputP95 := percentileInt(outputValues, 0.95)
	uncachedP95 := percentileInt(uncachedValues, 0.95)
	longContextSuspected := int64(cachedP95) > maxCachedPrefixBytesP95 || inputP95 > maxInputTokensP95
	volumeTooLarge := int64(cachedP95) > maxCachedPrefixBytesP95 ||
		inputP95 > maxInputTokensP95 ||
		freshP95 > maxFreshInputTokensP95 ||
		outputP95 > maxOutputTokensP95 ||
		uncachedP95 > maxUncachedTokensP95
	if volumeTooLarge {
		if int64(cachedP95) > maxCachedPrefixBytesP95 {
			failures = append(failures, fmt.Sprintf("runtime_cached_tokens_p95_over_budget=%d>%d", cachedP95, maxCachedPrefixBytesP95))
		}
		if inputP95 > maxInputTokensP95 {
			failures = append(failures, fmt.Sprintf("runtime_input_tokens_p95_over_budget=%d>%d", inputP95, maxInputTokensP95))
		}
		if freshP95 > maxFreshInputTokensP95 {
			failures = append(failures, fmt.Sprintf("runtime_fresh_input_tokens_p95_over_budget=%d>%d", freshP95, maxFreshInputTokensP95))
		}
		if outputP95 > maxOutputTokensP95 {
			failures = append(failures, fmt.Sprintf("runtime_output_tokens_p95_over_budget=%d>%d", outputP95, maxOutputTokensP95))
		}
		if uncachedP95 > maxUncachedTokensP95 {
			failures = append(failures, fmt.Sprintf("runtime_uncached_tokens_p95_over_budget=%d>%d", uncachedP95, maxUncachedTokensP95))
		}
	}
	cacheHitRate := 0.0
	if usageObservations > 0 {
		cacheHitRate = float64(cacheHits) / float64(usageObservations)
	}
	usageLogRef := pathPrivacyRef(workspace, usageLog)
	report := jsonObject{
		"command":                    "runtime-context-audit",
		"schema_version":             "runtime-context-audit.v1",
		"workspace_key":              workspaceKey,
		"generated_at":               time.Now().UTC().Format(time.RFC3339),
		"wuji_version":               builtinIronRulesVersion,
		"usage_log":                  usageLogRef,
		"usage_log_sha256":           usageLogHash,
		"usage_observations":         usageObservations,
		"cache_hit_rate":             cacheHitRate,
		"cached_tokens_p95":          cachedP95,
		"input_tokens_p95":           inputP95,
		"fresh_input_tokens_p95":     freshP95,
		"output_tokens_p95":          outputP95,
		"uncached_tokens_p95":        uncachedP95,
		"long_context_suspected":     longContextSuspected,
		"diagnosis":                  ternaryStatus(longContextSuspected, "cached-token-bloat-suspected-long-resident-or-outer-context", "no-long-context-bloat-signal"),
		"context_slimming_actions":   contextSlimmingRecommendations(longContextSuspected, inputP95, cachedP95, freshP95, outputP95, uncachedP95),
		"volume_gate":                ternaryStatus(!volumeTooLarge, "pass", "fail"),
		"runtime_skill_bytes":        runtimeSkillBytes,
		"repo_instruction_bytes":     repoInstructionBytes,
		"runtime_mirror_bytes":       mirrorBytes,
		"privacy_mode":               "numeric-usage-and-hash-only",
		"outer_context_claim_policy": "token-cost-cache-usage-claims-require-runtime-usage-evidence",
		"checks":                     []string{"runtime-skill-budget", "repo-instruction-budget", "runtime-mirror-budget", "outer-usage-volume", "long-context-suspect", "raw-payload-forbidden"},
		"budgets":                    jsonObject{"runtime_skill_bytes_max": maxRuntimeSkillBytes, "repo_instruction_bytes_max": maxRuntimeRepoInstructionBytes, "runtime_mirror_bytes_max": maxRuntimeMirrorBytes, "cached_prefix_tokens_p95_max": maxCachedPrefixBytesP95, "input_tokens_p95_max": maxInputTokensP95, "fresh_input_tokens_p95_max": maxFreshInputTokensP95, "output_tokens_p95_max": maxOutputTokensP95, "uncached_tokens_p95_max": maxUncachedTokensP95, "min_runtime_usage_observations": minRuntimeUsageObservations},
		"warnings":                   warnings,
		"status":                     ternaryStatus(len(failures) == 0, "pass", "fail"),
	}
	relUsage := filepath.ToSlash(strings.TrimPrefix(usageLogRef, "./"))
	manifestRelPaths := []string{"tools/wuji_cli.go", "AGENTS.md", "SKILL.md", "GLOBAL_AGENTS.md", "README.md"}
	if relUsage != "" && !strings.HasPrefix(relUsage, "external:") {
		manifestRelPaths = append([]string{relUsage}, manifestRelPaths...)
	}
	for key, value := range auditManifest(workspace, "runtime-context-audit", manifestRelPaths, runtimeSkillPath) {
		report[key] = value
	}
	if len(failures) > 0 {
		report["failures"] = uniqueStrings(failures)
	}
	if err := writeJSON(reportPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return printGate("runtime-context-audit", uniqueStrings(failures))
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
	allowUnsafe := false
	if raw, ok := argValue(args, "--allow-unsafe-command"); ok {
		parsed, err := parseBoolValue(raw)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		allowUnsafe = parsed
	}
	baseCommand := strings.ToLower(filepath.Base(command))
	previewAllowlist := map[string]bool{
		"wkhtmltoimage": true, "wkhtmltoimage.exe": true,
		"playwright": true, "playwright.cmd": true, "playwright.exe": true,
		"chrome": true, "chrome.exe": true, "msedge": true, "msedge.exe": true,
		"magick": true, "magick.exe": true,
	}
	unsafeCommands := map[string]bool{
		"powershell": true, "powershell.exe": true, "pwsh": true, "pwsh.exe": true,
		"cmd": true, "cmd.exe": true, "python": true, "python.exe": true,
		"node": true, "node.exe": true, "git": true, "git.exe": true,
	}
	if !previewAllowlist[baseCommand] && !allowUnsafe {
		return printGate("preview", []string{"preview_command_requires_allow_unsafe=" + baseCommand})
	}
	if unsafeCommands[baseCommand] && !allowUnsafe {
		return printGate("preview", []string{"preview_command_requires_allow_unsafe=" + baseCommand})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, commandArgs...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Fprintln(os.Stderr, "preview command timed out")
			return 1
		}
		fmt.Fprintln(os.Stderr, "preview command failed")
		return 1
	}
	failures := []string{}
	if !nonEmpty(output) {
		failures = append(failures, fmt.Sprintf("preview_output_missing_or_too_small=%s", pathPrivacyRef("", output)))
	}
	return printGate("preview", failures)
}

func pptTemplateInspectCommand(args []string) int {
	if _, ok := argValue(args, "--workspace"); !ok {
		usage()
		return 2
	}
	if _, ok := argValue(args, "--pptx"); !ok {
		usage()
		return 2
	}
	return runPowerShellScriptCommand("wuji-ppt-template-inspect.ps1", args)
}

func pptTemplateStarterCommand(args []string) int {
	if _, ok := argValue(args, "--workspace"); !ok {
		usage()
		return 2
	}
	if _, ok := argValue(args, "--pptx"); !ok {
		usage()
		return 2
	}
	if _, ok := argValue(args, "--map"); !ok {
		usage()
		return 2
	}
	if _, ok := argValue(args, "--out"); !ok {
		usage()
		return 2
	}
	return runPowerShellScriptCommand("wuji-ppt-template-starter.ps1", args)
}

func pptTemplateEditCommand(args []string) int {
	if _, ok := argValue(args, "--workspace"); !ok {
		usage()
		return 2
	}
	if _, ok := argValue(args, "--starter-pptx"); !ok {
		usage()
		return 2
	}
	if _, ok := argValue(args, "--map"); !ok {
		usage()
		return 2
	}
	if _, ok := argValue(args, "--out"); !ok {
		usage()
		return 2
	}
	return runPowerShellScriptCommand("wuji-ppt-template-edit.ps1", args)
}

func pptTemplateFidelityCommand(args []string) int {
	if _, ok := argValue(args, "--workspace"); !ok {
		usage()
		return 2
	}
	if _, ok := argValue(args, "--final-pptx"); !ok {
		usage()
		return 2
	}
	return runPowerShellScriptCommand("wuji-ppt-template-fidelity.ps1", args)
}

func pptHTMLFirstCommand(args []string) int {
	if _, ok := argValue(args, "--workspace"); !ok {
		usage()
		return 2
	}
	if _, ok := argValue(args, "--html"); !ok {
		usage()
		return 2
	}
	if _, ok := argValue(args, "--out"); !ok {
		usage()
		return 2
	}
	return runPowerShellScriptCommand("wuji-ppt-htmlfirst.ps1", args)
}

func pptCOMRefineCommand(args []string) int {
	if _, ok := argValue(args, "--pptx"); !ok {
		usage()
		return 2
	}
	if _, ok := argValue(args, "--out"); !ok {
		usage()
		return 2
	}
	return runPowerShellScriptCommand("wuji-ppt-com-refine.ps1", args)
}

func pptPipelineCommand(args []string) int {
	if _, ok := argValue(args, "--workspace"); !ok {
		usage()
		return 2
	}
	if _, ok := argValue(args, "--route"); !ok {
		usage()
		return 2
	}
	if _, ok := argValue(args, "--out"); !ok {
		usage()
		return 2
	}
	return runPowerShellScriptCommand("wuji-ppt-pipeline.ps1", args)
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
	styleLines := []string{"# style-lock", ""}
	pageRoleLines := []string{"# page-role-policy", ""}
	motionLines := []string{
		"# motion-plan",
		"",
		"- required: false",
		"- dynamic_source: none",
		"- motion_intent: static-ok",
		"- motion_roles: none",
		"- static_fallback: keep the editable PPTX honest; if later tasks require motion, add a live HTML demo instead of claiming PowerPoint already inherited it.",
		"- gate_note: upgrade this plan and attach a live HTML demo artifact before closeout whenever the task asks for dynamic experience.",
	}
	stylePreset := styleLockPresetForDeck(pptxPath, summary.Slides)
	styleLines = append(styleLines,
		fmt.Sprintf("- visual_system: %v", stylePreset["style_brief"]),
		fmt.Sprintf("- background_policy: %v", stylePreset["background"]),
		fmt.Sprintf("- highlight_policy: %v", stylePreset["highlights"]),
		fmt.Sprintf("- illustration_policy: %v", stylePreset["illustrations"]),
		fmt.Sprintf("- fixed_page_rule: %v", stylePreset["fixed_page_rule"]),
		fmt.Sprintf("- prompt_rule: %v", stylePreset["prompt_rule"]),
	)
	if keepDark, ok := stylePreset["keep_dark"].(bool); ok && keepDark {
		styleLines = append(styleLines, "- keep_dark_background: true")
	}
	if forbid, ok := stylePreset["forbid"].([]string); ok {
		for _, item := range forbid {
			styleLines = append(styleLines, "- forbid: "+item)
		}
	}
	pageRoleLines = append(pageRoleLines,
		"- 固定页型一旦在参考 deck 中被识别出来，只能承载同角色内容，不得挪作普通内容页。",
		"- 普通内容页优先复用内容页、图框页、信息页；不要盗用首页、目录页、单元页、总结页、结尾页的骨架。",
		"- 如果用户已经指定目录页/单元页/总结页/结尾页使用哪张模板页，同任务后续默认沿用，不再重新判断。",
	)
	lockedRoles := []jsonObject{}
	for i, slide := range summary.Slides {
		role := detectFixedPageRole(slide, i, len(summary.Slides))
		frameLines = append(frameLines, fmt.Sprintf("- slide-%02d %s: role=%s text=%d pic=%d shape=%d", i+1, slide.Name, role, slide.TextCount, slide.PicCount, slide.ShapeCount))
		illustrationMode := "无需插图"
		reason := "当前页结构较轻。"
		if slide.PicCount > 0 {
			illustrationMode = "复用参考图或参考图框"
			reason = "参考页已包含图像资产或图位。"
		} else if slideNeedsIllustration(slide) {
			illustrationMode = "补软件截图 / 步骤示意图 / image2 教学插图"
			reason = "当前页疑似教学或高密度操作内容，仅靠文字不够直观。"
		}
		signals := strings.Join(slideTeachingSignals(slide), ",")
		if signals == "" {
			signals = "none"
		}
		illustrationLines = append(illustrationLines, fmt.Sprintf("- slide-%02d [%s]: %s | signals=%s | reason=%s", i+1, role, illustrationMode, signals, reason))
		if role != "content" {
			pageRoleLines = append(pageRoleLines, fmt.Sprintf("- slide-%02d [%s]: fixed_page=true | page_type=%s | do_not_repurpose=true | source=%s", i+1, role, fixedPageTypeLabel(role), slide.Name))
			lockedRoles = append(lockedRoles, jsonObject{
				"slide":            i + 1,
				"role":             role,
				"page_type":        fixedPageTypeLabel(role),
				"fixed_page":       true,
				"do_not_repurpose": true,
				"source":           slide.Name,
			})
		}
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
	if err := os.WriteFile(filepath.Join(workspace, "style-lock.md"), []byte(strings.Join(styleLines, "\n")+"\n"), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := writeJSON(filepath.Join(workspace, "style-lock.json"), jsonObject{
		"style_name":      stylePreset["style_name"],
		"style_brief":     stylePreset["style_brief"],
		"background":      stylePreset["background"],
		"highlights":      stylePreset["highlights"],
		"illustrations":   stylePreset["illustrations"],
		"fixed_page_rule": stylePreset["fixed_page_rule"],
		"prompt_rule":     stylePreset["prompt_rule"],
		"keep_dark":       stylePreset["keep_dark"],
		"forbid":          stylePreset["forbid"],
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := os.WriteFile(filepath.Join(workspace, "page-role-policy.md"), []byte(strings.Join(pageRoleLines, "\n")+"\n"), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := os.WriteFile(filepath.Join(workspace, "motion-plan.md"), []byte(strings.Join(motionLines, "\n")+"\n"), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := writeJSON(filepath.Join(workspace, "page-role-policy.json"), jsonObject{
		"locked_roles": lockedRoles,
		"default_rules": []string{
			"fixed roles stay fixed",
			"do not repurpose cover/agenda/section/summary/ending frames",
			"content pages must use content-capable layouts",
		},
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := writeJSON(filepath.Join(workspace, "motion-plan.json"), jsonObject{
		"required":        false,
		"dynamic_source":  "none",
		"motion_intent":   "static-ok",
		"motion_roles":    []string{},
		"static_fallback": "keep the editable PPTX honest; if later tasks require motion, add a live HTML demo instead of claiming PowerPoint already inherited it.",
		"gate_note":       "upgrade this plan and attach a live HTML demo artifact before closeout whenever the task asks for dynamic experience.",
	}); err != nil {
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
	placeholderSlides := []string{}
	blankSlides := []string{}
	templateResidueSlides := []string{}
	for _, slide := range summary.Slides {
		if slide.PicCount > 0 && slide.TextCount == 0 && slide.ShapeCount <= 1 {
			imageOnlySlides = append(imageOnlySlides, slide.Name)
		}
		if len(slide.PlaceholderHits) > 0 {
			placeholderSlides = append(placeholderSlides, fmt.Sprintf("%s[%s]", slide.Name, strings.Join(slide.PlaceholderHits, "|")))
		}
		if slide.TextChars == 0 && slide.PicCount == 0 {
			if slide.ShapeCount == 0 {
				blankSlides = append(blankSlides, slide.Name)
			} else {
				templateResidueSlides = append(templateResidueSlides, slide.Name)
			}
		}
	}
	if len(imageOnlySlides) > 0 {
		failures = append(failures, "pptx_contains_image_only_slides="+strings.Join(imageOnlySlides, ","))
	}
	if len(placeholderSlides) > 0 {
		failures = append(failures, "pptx_contains_placeholder_text="+strings.Join(placeholderSlides, ","))
	}
	if len(blankSlides) > 0 {
		failures = append(failures, "pptx_contains_blank_slides="+strings.Join(blankSlides, ","))
	}
	if len(templateResidueSlides) > 0 {
		failures = append(failures, "pptx_contains_template_residue_slides="+strings.Join(templateResidueSlides, ","))
	}
	report := jsonObject{
		"pptx_path":               summary.PPTXPath,
		"slide_count":             len(summary.Slides),
		"media_count":             len(summary.Media),
		"layout_count":            len(summary.Layouts),
		"image_only_slides":       imageOnlySlides,
		"placeholder_slides":      placeholderSlides,
		"blank_slides":            blankSlides,
		"template_residue_slides": templateResidueSlides,
		"slides":                  summary.Slides,
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

func requiredPPTPlanStems() []string {
	return []string{"reference-frame-map", "reusable-asset-map", "illustration-plan", "style-lock", "page-role-policy", "motion-plan"}
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
	for _, stem := range requiredPPTPlanStems() {
		if path, ok := existingPlanFile(workspace, stem); ok {
			if !nonEmpty(path) {
				failures = append(failures, fmt.Sprintf("plan_file_too_small=%s", path))
			}
		} else {
			failures = append(failures, fmt.Sprintf("missing_required_plan=%s", stem))
		}
	}
	failures = append(failures, styleLockFailures(workspace)...)
	failures = append(failures, pageRolePolicyFailures(workspace)...)
	failures = append(failures, motionPlanFailures(workspace)...)
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
	illustrationPlanPath := ""
	requireDarkStyleLock := styleLockRequiresDarkBackground(workspace)
	for _, stem := range requiredPPTPlanStems() {
		if path, ok := existingPlanFile(workspace, stem); ok {
			if !nonEmpty(path) {
				failures = append(failures, fmt.Sprintf("plan_file_too_small=%s", path))
			} else if stem == "illustration-plan" {
				illustrationPlanPath = path
			}
		} else {
			failures = append(failures, fmt.Sprintf("missing_required_plan=%s", stem))
		}
	}
	failures = append(failures, illustrationPlanFailures(workspace)...)
	failures = append(failures, styleLockFailures(workspace)...)
	failures = append(failures, pageRolePolicyFailures(workspace)...)
	failures = append(failures, motionPlanFailures(workspace)...)
	if illustrationPlanPath != "" && hasTeachingSignalsInPlan(illustrationPlanPath) {
		if path, ok := existingPlanFile(workspace, "outline"); !ok || !nonEmpty(path) {
			failures = append(failures, "missing_required_content_artifact=outline")
		}
		if path, ok := existingPlanFile(workspace, "speaker-notes"); !ok || !nonEmpty(path) {
			failures = append(failures, "missing_required_content_artifact=speaker-notes")
		}
	}
	for _, stem := range []string{"pilot-preview", "pilot-page"} {
		if path, ok := existingPilotFile(workspace, stem); ok {
			if !nonEmpty(path) {
				failures = append(failures, fmt.Sprintf("pilot_file_too_small=%s", path))
			} else if stem == "pilot-preview" {
				failures = append(failures, pilotPreviewFailures(path)...)
				failures = append(failures, styleLockedPilotPreviewFailures(path, requireDarkStyleLock)...)
			}
		} else {
			failures = append(failures, fmt.Sprintf("missing_required_pilot=%s", stem))
		}
	}
	if path, ok := existingPlanFile(workspace, "pilot-preview-layout"); ok && nonEmpty(path) {
		failures = append(failures, pilotPreviewLayoutFailures(path)...)
	}
	if path, ok := existingPlanFile(workspace, "pilot-score"); ok {
		if !nonEmpty(path) {
			failures = append(failures, fmt.Sprintf("pilot_score_too_small=%s", path))
		}
	} else {
		failures = append(failures, "missing_required_pilot=pilot-score")
	}
	if path, ok := existingPlanFile(workspace, "pilot-approval"); ok {
		if approved, failure := pilotApprovalGranted(path); !approved {
			failures = append(failures, failure)
		}
	} else {
		failures = append(failures, "missing_required_pilot=pilot-approval")
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
	if privateEvidencePathDenied(manifestPath) {
		return printGate("mcp-guard", []string{"mcp_manifest_private_denied=" + pathPrivacyRef(workspace, manifestPath)})
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
				if privateEvidencePathDenied(path) {
					failures = append(failures, "filesystem_permission_private_denied="+pathPrivacyRef(workspace, path))
				} else if workspace != "" && !sameOrDescendant(path, workspace) {
					failures = append(failures, "filesystem_permission_outside_workspace="+pathPrivacyRef(workspace, path))
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

func supplyChainCommand(args []string) int {
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
	if privateEvidencePathDenied(manifestPath) {
		return printGate("supply-chain", []string{"supply_chain_manifest_private_denied=" + pathPrivacyRef(workspace, manifestPath)})
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
	for _, field := range []string{"name", "version", "source", "ref", "sha256", "license"} {
		if objectString(manifest, field) == "" {
			failures = append(failures, "supply_chain_manifest_missing_"+field)
		}
	}
	ref := objectString(manifest, "ref")
	if ref != "" && !regexp.MustCompile(`^[0-9a-fA-F]{40}$`).MatchString(ref) {
		failures = append(failures, "supply_chain_ref_must_be_40_char_commit_sha")
	}
	sha := objectString(manifest, "sha256")
	if sha != "" && !regexp.MustCompile(`^[0-9a-fA-F]{64}$`).MatchString(sha) {
		failures = append(failures, "supply_chain_sha256_invalid")
	}
	source := strings.ToLower(objectString(manifest, "source"))
	if (strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")) && !allowNetwork {
		failures = append(failures, "supply_chain_network_source_requires_explicit_allow")
	}
	if source == "local" && objectString(manifest, "local_path") == "" {
		failures = append(failures, "supply_chain_local_path_required_for_local_source")
	}
	if localPath := objectString(manifest, "local_path"); localPath != "" {
		localPathAllowed := true
		if workspace == "" {
			failures = append(failures, "supply_chain_local_path_requires_workspace")
			localPathAllowed = false
		} else if !sameOrDescendant(localPath, workspace) {
			failures = append(failures, "supply_chain_local_path_outside_workspace="+pathPrivacyRef(workspace, localPath))
			localPathAllowed = false
		} else if privateEvidencePathDenied(localPath) {
			failures = append(failures, "supply_chain_local_path_private_denied="+pathPrivacyRef(workspace, localPath))
			localPathAllowed = false
		}
		if localPathAllowed {
			hash, err := fileSHA256(localPath)
			if err != nil {
				failures = append(failures, "supply_chain_local_path_missing_or_unreadable="+pathPrivacyRef(workspace, localPath))
			} else if sha != "" && !strings.EqualFold(hash, sha) {
				failures = append(failures, "supply_chain_local_path_hash_mismatch="+pathPrivacyRef(workspace, localPath))
			}
		}
	}
	if execute, ok := objectBool(manifest, "execute_after_fetch"); ok && execute {
		failures = append(failures, "supply_chain_execute_after_fetch_forbidden")
	}
	if containsSecretLikeContent(string(rawManifest)) {
		failures = append(failures, "supply_chain_manifest_contains_secret_like_content")
	}
	return printGate("supply-chain", failures)
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
			"name":           name,
			"owner":          owner,
			"source":         source,
			"license":        license,
			"capability":     capability,
			"transport":      transport,
			"permissions":    permissions,
			"decision":       decision,
			"reason":         reason,
			"evidence_level": evidenceLevelFromDecision(decision),
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
		"bearer ",
		"gh" + "p_",
		"gh" + "o_",
		"xoxb" + "-",
		"-----begin private key-----",
		"-----begin rsa private key-----",
		"-----begin openssh private key-----",
	}
	for _, marker := range markers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	secretPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bsk-(?:proj-|live-)?[a-z0-9_-]{20,}`),
		regexp.MustCompile(`(?i)\b(?:api[_-]?key|apikey|password|cookie|authorization|token|secret)\s*[:=]\s*["']?[^"'\s]{12,}`),
	}
	for _, pattern := range secretPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func privacyHash(value string) string {
	normalized := normalizeSpace(strings.ToLower(value))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])[:16]
}

func hashStrings(values []string) []string {
	hashed := []string{}
	for _, value := range uniqueStrings(values) {
		hashed = append(hashed, privacyHash(value))
	}
	return hashed
}

func markerHits(text string, markers []string) []string {
	lower := strings.ToLower(text)
	hits := []string{}
	seen := map[string]bool{}
	for _, marker := range markers {
		normalized := strings.ToLower(strings.TrimSpace(marker))
		if normalized == "" || seen[normalized] {
			continue
		}
		if strings.Contains(lower, normalized) {
			hits = append(hits, marker)
			seen[normalized] = true
		}
	}
	return hits
}

func routeKeywordMatches(lowerQuery string, keywords []string) []string {
	matches := []string{}
	seen := map[string]bool{}
	for _, keyword := range keywords {
		trimmed := strings.TrimSpace(keyword)
		normalized := strings.ToLower(trimmed)
		if normalized == "" || seen[normalized] {
			continue
		}
		if strings.Contains(lowerQuery, normalized) {
			matches = append(matches, trimmed)
			seen[normalized] = true
		}
	}
	return dropSubsumedRouteMatches(matches)
}

func dropSubsumedRouteMatches(matches []string) []string {
	result := []string{}
	for index, match := range matches {
		normalized := strings.ToLower(strings.TrimSpace(match))
		if normalized == "" {
			continue
		}
		subsumed := false
		for otherIndex, other := range matches {
			if index == otherIndex {
				continue
			}
			otherNormalized := strings.ToLower(strings.TrimSpace(other))
			if len(otherNormalized) > len(normalized) && strings.Contains(otherNormalized, normalized) {
				subsumed = true
				break
			}
		}
		if !subsumed {
			result = append(result, match)
		}
	}
	return result
}

func routeTierSignalCount(routeID string, matchedCount int, query string) int {
	if strings.EqualFold(routeID, "execution-base") && len(markerHits(query, performanceRouteMarkers())) > 0 && matchedCount > 3 {
		return 3
	}
	return matchedCount
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

func orderedUniqueStrings(values []string) []string {
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
	return result
}

func cloneBuiltinModelProfiles() map[string]modelProfile {
	cloned := map[string]modelProfile{}
	for tier, profile := range builtinModelProfiles {
		cloned[tier] = profile
	}
	return cloned
}

func mergedModelProfiles(config map[string]any) map[string]modelProfile {
	profiles := cloneBuiltinModelProfiles()
	rawProfiles, ok := objectMap(config, "model_profiles")
	if !ok {
		return profiles
	}
	for tier, raw := range rawProfiles {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		profile := profiles[tier]
		if providerID := objectString(item, "provider_id"); providerID != "" {
			profile.ProviderID = providerID
		}
		if model, ok := item["model"].(string); ok {
			profile.Model = strings.TrimSpace(model)
		}
		if effort := objectString(item, "reasoning_effort"); effort != "" {
			profile.ReasoningEffort = effort
		}
		profiles[tier] = profile
	}
	return profiles
}

func resolvedDefaultModelTier(config map[string]any) string {
	tier := objectString(config, "default_model_tier")
	if tier == "" {
		tier = builtinDefaultModelTier
	}
	if _, ok := builtinModelProfiles[tier]; ok {
		return tier
	}
	return builtinDefaultModelTier
}

func cloneBuiltinRoutingRules() []routeRule {
	canonRules := canonRoutingRules()
	cloned := make([]routeRule, 0, len(canonRules))
	for _, rule := range canonRules {
		copied := rule
		copied.Keywords = append([]string{}, rule.Keywords...)
		cloned = append(cloned, copied)
	}
	return cloned
}

func canonTopLevelRoles() []string {
	return []string{
		"aji",
		"staff-runtime",
		"nuwa-preflight",
		"white-hat",
		"guard-office",
		"root-cause-officer",
		"audit",
		"quality-inspection",
		"performance-benchmark-on-demand",
		"compliance-on-demand",
		"evolution-profile",
	}
}

func performanceRouteMarkers() []string {
	return []string{
		"performance", "benchmark", "bench", "bench-report", "latency", "throughput", "hit rate", "cache hit", "cache bloat",
		"token", "token cost", "token volume", "cached tokens", "cost", "speed", "slow", "p95", "p99", "cpu", "memory", "resource",
		"rework cost", "headroom", "prefix cache", "context-bloat", "context bloat", "context volume", "token optimization",
		"long context", "large context", "context window", "cached token volume", "200k", "blue hit",
		"上下文", "长上下文", "上下文太长", "蓝色命中", "命中体量", "后台token", "后台 token", "命中的数量",
		"mtp", "tpm", "tq", "sageattention", "attention acceleration", "cache acceleration",
	}
}

func analysisCompletenessMarkers() []string {
	return []string{
		"architecture", "architecture analysis", "system analysis", "system design", "design review", "route review", "routing review",
		"rule fusion", "kernel review", "kernel architecture", "whole system", "full-system", "structural analysis", "analyze architecture",
		"analyze system", "unknown modules", "incomplete docs", "complete materials", "material coverage", "source inventory",
		"\u67b6\u6784", "\u67b6\u6784\u5206\u6790", "\u5206\u6790\u67b6\u6784", "\u7cfb\u7edf\u5206\u6790",
		"\u8def\u7531\u590d\u67e5", "\u89c4\u5219\u878d\u5408", "\u8bbe\u8ba1\u5ba1\u67e5", "\u5168\u9762\u8d44\u6599",
		"\u8d44\u6599\u4e0d\u5168", "\u4e0d\u8981\u731c", "\u4e0d\u80fd\u731c", "\u4e0d\u5b8c\u6574", "\u6ca1\u62ff\u5168",
	}
}

func analysisCompletenessRequired(query string) bool {
	return len(markerHits(query, analysisCompletenessMarkers())) > 0
}

func canonRoutingRules() []routeRule {
	return []routeRule{
		{ID: "search", Name: "search-intelligence", Keywords: []string{"search", "find", "research", "web", "latest", "cite", "github"}, ProviderID: "deepseek-web", Priority: 100},
		{ID: "code", Name: "code-development", Keywords: []string{"code", "fix", "debug", "bugfix", "compile", "function", "script", "test", "tests", "refactor", "implement", "plugin", "python", "powershell", "rust", "architecture", "system analysis", "module", "module map", "dependency", "repository", "repo"}, ProviderID: "deepseek-web", Priority: 80},
		{ID: "execution-base", Name: "go-execution-base", Keywords: append([]string{"wuji-cli", "go", "guard", "audit", "claim-guard", "reference-guard", "truth-state", "finish-or-block", "closeout-check", "pptx-preflight", "pptx-batch-gate", "pptx-audit", "asset-map", "time-guard", "mcp-guard", "reference-frame-map", "reusable-asset-map", "illustration-plan", "pilot-page", "pilot-preview", "pilot-score"}, performanceRouteMarkers()...), ProviderID: "deepseek-web", Priority: 82},
		{ID: "content", Name: "content-document", Keywords: []string{"article", "document", "report", "word", "docx", "markdown", "blog", "story", "course", "proposal"}, ProviderID: "deepseek-web", Priority: 70},
		{ID: "visual", Name: "visual-ppt-ui", Keywords: []string{"ppt", "presentation", "slide", "deck", "design", "ui", "frontend", "html", "css", "page", "landing", "preview", "opendesign", "remotion"}, ProviderID: "deepseek-web", Priority: 60},
		{ID: "video", Name: "video-motion", Keywords: []string{"video", "motion", "animated", "animation", "stage", "sprite", "timeline", "product demo", "demo video", "mp4", "gif", "reel"}, ProviderID: "deepseek-web", Priority: 55},
		{ID: "imagegen", Name: "imagegen-direct-image", Keywords: []string{"image", "illustration", "poster", "cover", "generate image", "imagegen"}, Priority: 52},
		{ID: "prompt", Name: "prompt-engineering", Keywords: []string{"prompt", "storyboard-spec", "image-spec", "rewrite prompt", "expand prompt"}, ProviderID: "deepseek-web", Priority: 50},
		{ID: "spreadsheet", Name: "spreadsheet-data", Keywords: []string{"spreadsheet", "excel", "xlsx", "csv", "table", "ledger"}, ProviderID: "deepseek-web", Priority: 58},
		{ID: "comfyui", Name: "comfyui-workflow", Keywords: []string{"comfyui", "workflow json", "node graph", "screenshot", "render"}, ProviderID: "deepseek-web", Priority: 95},
		{ID: "quality-inspection", Name: "quality-inspection", Keywords: []string{"quality-inspection", "qa", "quality review", "acceptance review", "final acceptance", "final verification", "white-hat review", "guard-office review", "release acceptance", "验收", "质检"}, ProviderID: "deepseek-web", Priority: 40},
		{ID: "evolve", Name: "evolution-distillation", Keywords: []string{"evolve", "distill", "fusion", "rules", "skills", "source pool", "recipe", "eval set", "promotion", "retire", "learning"}, ProviderID: "deepseek-web", Priority: 30},
		{ID: "chat", Name: "chat", Keywords: []string{}, ProviderID: "deepseek-web", Priority: 0},
	}
}

func mergedRoutingRules(config map[string]any) []routeRule {
	rules := cloneBuiltinRoutingRules()
	indexByID := map[string]int{}
	for index, rule := range rules {
		indexByID[strings.ToLower(rule.ID)] = index
	}
	rawRules, ok := objectSlice(config, "routing_rules")
	if !ok {
		return rules
	}
	for _, raw := range rawRules {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		id := objectString(item, "id")
		if id == "" {
			continue
		}
		key := strings.ToLower(id)
		rule := routeRule{ID: id}
		targetIndex, exists := indexByID[key]
		if exists {
			rule = rules[targetIndex]
		}
		if name := objectString(item, "name"); name != "" {
			rule.Name = name
		}
		if keywords, ok := objectSlice(item, "keywords"); ok {
			keywordList := []string{}
			for _, rawKeyword := range keywords {
				if keyword, ok := rawKeyword.(string); ok {
					keywordList = append(keywordList, keyword)
				}
			}
			rule.Keywords = orderedUniqueStrings(keywordList)
		}
		if strings.EqualFold(rule.ID, "execution-base") {
			rule.Keywords = orderedUniqueStrings(append(rule.Keywords, performanceRouteMarkers()...))
		}
		if providerID := objectString(item, "provider_id"); providerID != "" {
			rule.ProviderID = providerID
		}
		if model, ok := item["model"].(string); ok {
			rule.Model = strings.TrimSpace(model)
		}
		if priority, ok := intFromAny(item["priority"]); ok {
			rule.Priority = priority
		}
		if exists {
			rules[targetIndex] = rule
			continue
		}
		indexByID[key] = len(rules)
		rules = append(rules, rule)
	}
	return rules
}

func modelProfilesAsJSON(profiles map[string]modelProfile) jsonObject {
	keys := make([]string, 0, len(profiles))
	for key := range profiles {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := jsonObject{}
	for _, key := range keys {
		profile := profiles[key]
		result[key] = jsonObject{
			"provider_id":      profile.ProviderID,
			"model":            profile.Model,
			"reasoning_effort": profile.ReasoningEffort,
		}
	}
	return result
}

func routeRulesAsJSON(rules []routeRule) []jsonObject {
	result := make([]jsonObject, 0, len(rules))
	for _, rule := range rules {
		result = append(result, jsonObject{
			"id":          rule.ID,
			"name":        rule.Name,
			"keywords":    append([]string{}, rule.Keywords...),
			"provider_id": rule.ProviderID,
			"model":       rule.Model,
			"priority":    rule.Priority,
		})
	}
	return result
}

func routeRuleConflictFailures(rules []routeRule) []string {
	failures := []string{}
	ids := map[string]string{}
	keywords := map[string]string{}
	for _, rule := range rules {
		id := strings.ToLower(strings.TrimSpace(rule.ID))
		if id == "" {
			failures = append(failures, "routing_rule_missing_id")
		} else if prior, ok := ids[id]; ok {
			failures = append(failures, "routing_rule_duplicate_id="+prior+"|"+rule.ID)
		} else {
			ids[id] = rule.ID
		}
		for _, keyword := range rule.Keywords {
			key := strings.ToLower(strings.TrimSpace(keyword))
			if key == "" {
				continue
			}
			if owner, ok := keywords[key]; ok && owner != rule.ID {
				failures = append(failures, "routing_keyword_conflict="+keyword+":"+owner+"|"+rule.ID)
				continue
			}
			keywords[key] = rule.ID
		}
	}
	return uniqueStrings(failures)
}

func pluginBindingsAsJSON() []jsonObject {
	result := make([]jsonObject, 0, len(builtinPluginBindings))
	for _, binding := range builtinPluginBindings {
		result = append(result, jsonObject{
			"plugin":  binding.Plugin,
			"owners":  append([]string{}, binding.Owners...),
			"purpose": binding.Purpose,
		})
	}
	return result
}

func canonPluginBindingsAsJSON() []jsonObject {
	bindings := []pluginBinding{
		{Plugin: "Browser", Owners: []string{"visual-profile", "development-profile", "intelligence-profile", "quality-inspection"}, Purpose: "open, inspect, interact, test, and screenshot browser surfaces"},
		{Plugin: "Documents", Owners: []string{"content-profile", "intelligence-profile", "audit"}, Purpose: "create, edit, organize, and archive Word/document artifacts"},
		{Plugin: "Spreadsheets", Owners: []string{"data-profile", "intelligence-profile", "content-profile", "performance-benchmark-on-demand"}, Purpose: "spreadsheets, structured data, analysis, and delivery artifacts"},
		{Plugin: "Presentations", Owners: []string{"visual-profile", "quality-inspection"}, Purpose: "create, modify, export, and preview PPTX artifacts"},
	}
	result := make([]jsonObject, 0, len(bindings))
	for _, binding := range bindings {
		result = append(result, jsonObject{
			"plugin":  binding.Plugin,
			"owners":  append([]string{}, binding.Owners...),
			"purpose": binding.Purpose,
		})
	}
	return result
}

func mcpPoliciesAsJSON() []jsonObject {
	result := make([]jsonObject, 0, len(builtinMCPPolicies))
	for _, policy := range builtinMCPPolicies {
		result = append(result, jsonObject{
			"scope":    policy.Scope,
			"decision": policy.Decision,
			"reason":   policy.Reason,
		})
	}
	return result
}

func canonMCPPoliciesAsJSON() []jsonObject {
	policies := []canonDecision{
		{Scope: "local-low-permission", Decision: "absorb", Reason: "local low-permission tools may enter candidate capability pools when boundaries are clear"},
		{Scope: "network-or-account", Decision: "defer", Reason: "network, account, OAuth, or external write access requires task-level authorization"},
		{Scope: "high-risk-or-secrets", Decision: "reject/defer", Reason: "high-permission, secret-bearing, unclear-source, or unclear-license tools do not enter the default main chain"},
	}
	result := make([]jsonObject, 0, len(policies))
	for _, policy := range policies {
		result = append(result, jsonObject{
			"scope":    policy.Scope,
			"decision": policy.Decision,
			"reason":   policy.Reason,
		})
	}
	return result
}

func distilledAtomNamesByResidency(residency string) []string {
	names := []string{}
	for _, atom := range distilledAtomRegistry {
		if atom.Residency == residency {
			names = append(names, atom.Name)
		}
	}
	return names
}

func distilledAtomKnownMap() map[string]bool {
	known := map[string]bool{}
	for _, atom := range distilledAtomRegistry {
		known[atom.Name] = true
	}
	return known
}

func distilledAtomOwnerMap() jsonObject {
	owners := jsonObject{}
	for _, atom := range distilledAtomRegistry {
		owners[atom.Name] = atom.Owner
	}
	owners["assumption-ledger"] = "white-hat+quality-inspection+audit"
	owners["claim-fact-check"] = "white-hat+audit+intelligence-profile"
	owners["reversible-evidence-handle"] = "go-execution-base+audit+quality-inspection"
	owners["content-type-compression-router"] = "go-execution-base+performance-benchmark-on-demand"
	owners["version-doc-mcp"] = "development-profile+intelligence-profile+guard-office"
	owners["guarded-realtime-source-search"] = "intelligence-profile+guard-office+audit"
	owners["research-evidence-pack"] = "intelligence-profile+audit+content-profile"
	owners["skill-stocktake-daily-library"] = "evolution-profile+audit+white-hat"
	owners["verified-learning-loop"] = "evolution-profile+quality-inspection+go-execution-base+performance-benchmark-on-demand"
	owners["disciplined-debug-loop"] = "development-profile+quality-inspection"
	owners["prior-art-solution-search"] = "intelligence-profile+owner-profile+guard-office"
	owners["root-cause-radar"] = "root-cause-officer+development-profile+quality-inspection+white-hat"
	owners["parallel-hypothesis-fanout"] = "staff-runtime+quality-inspection"
	owners["patch-debt-root-cure"] = "evolution-profile+root-cause-officer+audit+performance-benchmark-on-demand"
	owners["terminal-real-run-verification"] = "quality-inspection+audit+go-execution-base"
	owners["html-native-design-canvas"] = "visual-profile+quality-inspection"
	owners["brand-asset-protocol"] = "visual-profile+intelligence-profile+guard-office"
	owners["anti-ai-slop-visual-rules"] = "visual-profile+quality-inspection+white-hat"
	owners["design-direction-triad"] = "staff-runtime+visual-profile+nuwa-preflight"
	owners["html-deck-to-editable-pptx"] = "visual-profile+go-execution-base+quality-inspection"
	owners["motion-stage-sprite-engine"] = "visual-profile+performance-benchmark-on-demand"
	return owners
}

func distilledAtomPresenceMap() map[string]bool {
	presence := map[string]bool{}
	for _, atom := range distilledAtomRegistry {
		presence[atom.Name] = false
	}
	return presence
}

func intelligenceProfileContract() jsonObject {
	return jsonObject{
		"role":          "candidate-scout-not-research-system",
		"search_scope":  "wide-recall-shallow-first",
		"github_status": "first-class-source-for-code-tool-plugin-skill-bug-prior-art",
		"may_do": []string{
			"search",
			"candidate-metadata",
			"dedupe-cluster",
			"evidence-handle",
		},
		"must_not_do": []string{
			"final-analysis",
			"deep-extract-by-default",
			"distillation-decision",
			"adoption-decision",
			"install-or-execute",
		},
		"candidate_fields": []string{
			"title",
			"url",
			"source_type",
			"snippet",
			"updated_at",
			"stars_or_activity",
			"license",
			"suspected_use",
			"risk_signal",
			"next_gate",
		},
		"handoff": "main-chain-assigns-deep-read-analysis-distillation-execution",
	}
}

func conciseExecutionContract() jsonObject {
	return jsonObject{
		"objective": "short-precise-high-hit-low-total-cost",
		"must_do": []string{
			"single-message-precision",
			"minimal-needed-context",
			"prior-art-before-invention-when-uncertain",
			"root-cause-before-patch",
			"fresh-output-uncached-volume-gated",
			"cached-volume-before-hit-rate-claim",
			"stable-prefix-small-not-just-cacheable",
		},
		"must_not_do": []string{
			"verbose-status-padding",
			"unneeded-preflight-loop",
			"context-shift-from-cached-to-uncached",
			"from-scratch-tooling-when-existing-solution-fits",
			"short-answer-that-causes-rework",
		},
		"cost_vector": []string{
			"cached_tokens_p95",
			"input_tokens_p95",
			"fresh_input_tokens_p95",
			"output_tokens_p95",
			"uncached_tokens_p95",
			"cache_hit_rate",
			"tokens_per_success",
			"retries",
		},
	}
}

func executionBudgetContract() jsonObject {
	return jsonObject{
		"objective": "scoped-fast-real-completion",
		"policy":    "use the smallest verification tier that preserves first-pass success, evidence, and user constraints",
		"tiers": []jsonObject{
			{
				"id":                  "FAST_REPLY",
				"scope":               "discussion, direct answer, or analysis with no requested file changes",
				"officer_mode":        "perspective-only-when-named",
				"verification":        "no tool gate unless factual, current, or local evidence is required",
				"full_audit":          false,
				"full_suite_max_runs": 0,
			},
			{
				"id":                  "LIGHT_TASK",
				"scope":               "small scoped edit or single-owner task",
				"officer_mode":        "only explicitly triggered officers, one concise verdict",
				"verification":        "targeted command, artifact, or focused test",
				"full_audit":          false,
				"full_suite_max_runs": 0,
			},
			{
				"id":                  "STRUCTURAL_TASK",
				"scope":               "router, kernel, officer, gate, multi-file, or root-cause work",
				"officer_mode":        "triggered officers may run in parallel, then exit after merge",
				"verification":        "targeted gates during work, one full verification at final if touched surfaces require it",
				"full_audit":          "final-only-when-surface-requires",
				"full_suite_max_runs": 1,
			},
			{
				"id":                  "RELEASE_TASK",
				"scope":               "explicit full-legion scan, release, broad cleanup, or final completion claim",
				"officer_mode":        "all relevant officers may review once under single main-chain merge",
				"verification":        "full audit and real-run closeout once at final",
				"full_audit":          true,
				"full_suite_max_runs": 1,
			},
		},
		"must_do": []string{
			"bind-current-scope-before-expansion",
			"run-targeted-verification-before-full-suite",
			"keep-officers-on-demand-and-exit-after-merge",
			"treat-runtime-context-audit-as-token-cost-cache-claim-only",
			"finish-current-scope-without-reopen-ceremony",
		},
		"must_not_do": []string{
			"escalate-small-task-to-full-legion-scan",
			"spawn-sidecars-for-officer-perspectives",
			"repeat-full-suite-after-small-edits",
			"block-non-token-work-on-missing-runtime-usage-log",
			"continue-low-value-sweep-outside-current-scope",
		},
	}
}

func executionBudgetContractFailures(contract map[string]any, prefix string) []string {
	failures := []string{}
	if objectString(contract, "objective") != "scoped-fast-real-completion" {
		failures = append(failures, prefix+"bad_objective")
	}
	for _, required := range []string{
		"bind-current-scope-before-expansion",
		"run-targeted-verification-before-full-suite",
		"keep-officers-on-demand-and-exit-after-merge",
		"treat-runtime-context-audit-as-token-cost-cache-claim-only",
		"finish-current-scope-without-reopen-ceremony",
	} {
		if !containsExactString(stringSlice(contract, "must_do"), required) {
			failures = append(failures, prefix+"missing_must_do="+required)
		}
	}
	for _, forbidden := range []string{
		"escalate-small-task-to-full-legion-scan",
		"spawn-sidecars-for-officer-perspectives",
		"repeat-full-suite-after-small-edits",
		"block-non-token-work-on-missing-runtime-usage-log",
		"continue-low-value-sweep-outside-current-scope",
	} {
		if !containsExactString(stringSlice(contract, "must_not_do"), forbidden) {
			failures = append(failures, prefix+"missing_must_not_do="+forbidden)
		}
	}
	return failures
}

func analysisCompletenessContract() jsonObject {
	return jsonObject{
		"objective": "complete-materials-before-architecture-analysis",
		"applies_to": []string{
			"architecture-analysis",
			"system-analysis",
			"rule-fusion",
			"route-review",
			"design-review",
		},
		"must_do": []string{
			"collect-material-inventory",
			"state-coverage-and-gaps",
			"ask-user-for-missing-materials-when-critical",
			"separate-fact-inference-and-unknown",
			"no-final-conclusion-from-incomplete-evidence",
		},
		"must_not_do": []string{
			"guess-architecture-from-partial-materials",
			"treat-sample-as-whole-system",
			"hide-coverage-gaps",
			"promote-uncertain-claim-to-fact",
		},
		"outputs": []string{
			"material_inventory",
			"coverage_gaps",
			"confidence_level",
			"missing_material_request",
		},
		"handoff": "main-chain-may-answer-only-with-evidence-bound-scope-until-critical-materials-arrive",
	}
}

func analysisCompletenessContractFailures(contract map[string]any, prefix string) []string {
	failures := []string{}
	if objectString(contract, "objective") != "complete-materials-before-architecture-analysis" {
		failures = append(failures, prefix+"bad_objective")
	}
	for _, required := range []string{
		"collect-material-inventory",
		"state-coverage-and-gaps",
		"ask-user-for-missing-materials-when-critical",
		"separate-fact-inference-and-unknown",
		"no-final-conclusion-from-incomplete-evidence",
	} {
		if !containsExactString(stringSlice(contract, "must_do"), required) {
			failures = append(failures, prefix+"missing_must_do="+required)
		}
	}
	for _, forbidden := range []string{
		"guess-architecture-from-partial-materials",
		"treat-sample-as-whole-system",
		"hide-coverage-gaps",
		"promote-uncertain-claim-to-fact",
	} {
		if !containsExactString(stringSlice(contract, "must_not_do"), forbidden) {
			failures = append(failures, prefix+"missing_must_not_do="+forbidden)
		}
	}
	return failures
}

func builtinCanonReport() jsonObject {
	return jsonObject{
		"iron_rules_version": builtinIronRulesVersion,
		"default_model_tier": builtinDefaultModelTier,
		"top_level_roles":    canonTopLevelRoles(),
		"kernel_source":      "kernel-source.json",
		"routing_kernel": jsonObject{
			"version": "three-layer-v1",
			"layers": []jsonObject{
				{
					"id":      "task-routing",
					"owner":   "staff-runtime",
					"purpose": "state, owner profile, oversight chain, closeout policy",
				},
				{
					"id":      "capability-mount",
					"owner":   "staff-runtime+owner-profile",
					"purpose": "mount distilled OpenSquilla atoms, plugins, MCP, and gap-fill capabilities only when needed",
				},
				{
					"id":      "deterministic-execution",
					"owner":   "go-execution-base",
					"purpose": "run repeatable local gates, audits, preflight, bench, and context packing",
				},
			},
		},
		"model_profiles":     modelProfilesAsJSON(cloneBuiltinModelProfiles()),
		"routing_rules":      routeRulesAsJSON(canonRoutingRules()),
		"built_in_plugins":   canonPluginBindingsAsJSON(),
		"mcp_default_policy": canonMCPPoliciesAsJSON(),
		"distilled_atom_kernel": jsonObject{
			"policy":               "gap-fill-replace-weaker-atoms-no-stack",
			"resident_light_atoms": distilledAtomNamesByResidency("resident-light"),
			"on_demand_atoms":      distilledAtomNamesByResidency("on-demand"),
			"owner_map":            distilledAtomOwnerMap(),
		},
		"intelligence_profile_contract":  intelligenceProfileContract(),
		"concise_execution_contract":     conciseExecutionContract(),
		"execution_budget_contract":      executionBudgetContract(),
		"analysis_completeness_contract": analysisCompletenessContract(),
		"optimization_kernel": jsonObject{
			"objective":               "smaller-stable-prefix-with-equal-or-better-hit-rate",
			"resident_policy":         "minimal-stable-skeleton-only",
			"mount_policy":            "minimal-gap-first",
			"tool_output_policy":      "compress-before-reuse-preserve-evidence",
			"evidence_retention":      "raw-handle-kept-summary-fed",
			"prefix_cache_discipline": "byte-stable-prefix-volatile-facts-late",
			"prefix_canon_policy":     "ordered-fields-short-canon-no-duplicate-phrasing",
			"retire_policy":           "replace-weaker-atoms-instead-of-stacking",
		},
		"canon_source": "go-builtin",
	}
}

func kernelSourcePath(workspace string) string {
	if strings.TrimSpace(workspace) == "" {
		return absClean(filepath.Join(".", "kernel-source.json"))
	}
	return absClean(filepath.Join(workspace, "kernel-source.json"))
}

func containsAllMarkers(text string, markers []string) bool {
	for _, marker := range markers {
		if !strings.Contains(text, marker) {
			return false
		}
	}
	return true
}

func mirrorDriftFailures(path string, requiredMarkers []string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return []string{"mirror_unreadable=" + absClean(path)}
	}
	text := string(content)
	lowerText := strings.ToLower(text)
	failures := []string{}
	for _, marker := range requiredMarkers {
		if !strings.Contains(text, marker) {
			failures = append(failures, "mirror_missing_marker="+filepath.Base(path)+":"+marker)
		}
	}
	if strings.Contains(lowerText, "second router") &&
		!strings.Contains(lowerText, "not a second router") &&
		!strings.Contains(lowerText, "do not let") &&
		!strings.Contains(lowerText, "must not") &&
		!strings.Contains(lowerText, "forbidden") {
		failures = append(failures, "mirror_ambiguous_second_router_claim="+filepath.Base(path))
	}
	return failures
}

func sourcePoolTokenSet(values ...[]string) map[string]bool {
	result := map[string]bool{}
	for _, list := range values {
		for _, value := range list {
			for _, token := range splitSourcePoolTokens(value) {
				result[strings.ToLower(token)] = true
			}
		}
	}
	return result
}

func splitSourcePoolTokens(sourcePool string) []string {
	sourcePool = strings.ReplaceAll(sourcePool, ",", "+")
	parts := strings.Split(sourcePool, "+")
	result := []string{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func fusionSourcePoolCatalog(kernelObj map[string]any, fusionMatrixObj map[string]any) map[string]bool {
	catalog := sourcePoolTokenSet(stringSlice(kernelObj, "source_pools"), stringSlice(fusionMatrixObj, "source_pools"))
	for _, token := range []string{
		"AnySearch",
		"Agent-Reach",
		"AiToEarn",
		"assumption-checker",
		"audit",
		"audit-trace-style",
		"ChinaTextbook",
		"codex-plugins",
		"codex-runtime",
		"codex-runtime-style",
		"codex-subagents",
		"context7-mcp",
		"delta-debugging",
		"development-profile",
		"ECC",
		"external-agent-shells",
		"fault-localization",
		"GhostTrack",
		"go-execution-base",
		"goose",
		"goose-style",
		"guard-office",
		"hallucination-detector",
		"headroom",
		"Hermes",
		"html-ppt-skill",
		"huashu-design",
		"intelligence-profile",
		"last30days",
		"llama.cpp",
		"multi-agent-rca",
		"nuwa-preflight",
		"open-notebook",
		"openai-plugins",
		"opencv-style",
		"parallel-sidecar-execution",
		"pg_durable",
		"pg_durable-style",
		"project-nomad",
		"quality-inspection",
		"reasonix",
		"research-evidence-pack",
		"research-mode",
		"source-pool-not-shell",
		"spreadsheet-profile",
		"staff-runtime",
		"Superpowers",
		"taste-skill",
		"tolaria",
		"trace-analysis",
		"turbovec-style",
		"verified-learning-loop",
		"visual-profile",
		"web-search-rag",
		"white-hat",
	} {
		catalog[strings.ToLower(token)] = true
	}
	return catalog
}

func sourcePoolFailures(atom string, sourcePool string, catalog map[string]bool) []string {
	failures := []string{}
	for _, token := range splitSourcePoolTokens(sourcePool) {
		if !catalog[strings.ToLower(token)] {
			failures = append(failures, "fusion_matrix_unknown_source_pool="+atom+":"+token)
		}
	}
	return failures
}

func isActiveFusionDecision(decision string) bool {
	switch decision {
	case "resident", "mount-on-demand", "replace":
		return true
	default:
		return false
	}
}

func activeFusionDecisionLooksLikeRuntimeAtom(atom string, sourcePool string) bool {
	atomLower := strings.ToLower(atom)
	sourcePoolLower := strings.ToLower(sourcePool)
	if strings.Contains(sourcePoolLower, "github-trending-20260608-style") {
		return true
	}
	markers := []string{
		"github-trending",
		"last30days",
		"taste-skill",
		"open-notebook",
		"tolaria",
		"turbovec",
		"goose",
		"pg_durable",
		"opencv-style",
		"openai-plugins",
	}
	return containsAny(atomLower, markers)
}

func fusionDecisionNeedsReject(atom string, sourcePool string, reason string, fusionPolicy string) bool {
	decisionSurface := strings.ToLower(atom + " " + sourcePool)
	directRejectMarkers := []string{
		"ghosttrack",
		"chinatextbook",
		"aitoearn",
		"project-nomad",
		"external-agent-shell",
		"privacy-invasive",
		"copyright-heavy",
		"monetization",
		"hype",
		"plugin-runtime",
		"risk_surface",
		"risk-surface",
		"tracking",
	}
	if containsAny(decisionSurface, directRejectMarkers) {
		return true
	}
	if !containsAny(decisionSurface, []string{"risk", "github-trending", "plugin-runtime", "external-agent-shell"}) {
		return false
	}
	riskDetail := strings.ToLower(reason + " " + fusionPolicy)
	return containsAny(riskDetail, []string{"high-risk", "write-all", "secrets", "copyright-heavy", "privacy-invasive", "tracking"})
}

func opensquillaRuntimeSurfaceFailures(paths []string) []string {
	failures := []string{}
	launchMarkers := []string{
		"gateway start",
		"gateway stop",
		"gateway restart",
		"mcp-server run",
		"opensquilla agent",
	}
	for _, path := range paths {
		if dirExists(path) {
			failures = append(failures, "opensquilla_runtime_directory_present="+absClean(path))
			continue
		}
		if !fileExists(path) {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, "opensquilla_runtime_surface_unreadable="+absClean(path))
			continue
		}
		lowerText := strings.ToLower(string(raw))
		if strings.Contains(lowerText, "opensquilla-commander") {
			failures = append(failures, "opensquilla_commander_skill_present="+absClean(path))
		}
		if hits := markerHits(lowerText, launchMarkers); len(hits) > 0 {
			failures = append(failures, "opensquilla_external_launcher_present="+absClean(path)+":"+strings.Join(hits, "|"))
		}
	}
	return failures
}

func activeLegacySkillSurfaceFailures(paths []string) []string {
	failures := []string{}
	for _, path := range paths {
		if dirExists(path) || fileExists(path) {
			failures = append(failures, "active_legacy_skill_surface_present="+absClean(path))
		}
	}
	return failures
}

func nestedExampleSkillSurfaceFailures(patterns []string) []string {
	failures := []string{}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			failures = append(failures, "nested_example_skill_glob_invalid="+absClean(pattern))
			continue
		}
		for _, match := range matches {
			if fileExists(match) {
				failures = append(failures, "nested_example_skill_surface_present="+absClean(match))
			}
		}
	}
	return failures
}

func runtimeResidueFailures(workspace string) []string {
	failures := []string{}
	checks := []struct {
		rel      string
		patterns []string
	}{
		{
			rel: "tools/wuji_cli.go",
			patterns: []string{
				`var\s+builtinRoutingRules\b`,
				`quality-inspection-profile`,
				`performance_benchmark\b`,
				`Owner:\s*"performance-benchmark(\+|")`,
				`Owner:\s*"[^"]*\+performance-benchmark(\+|")`,
				`"performance-benchmark"\s*[,}]`,
			},
		},
		{
			rel: "kernel-source.json",
			patterns: []string{
				`quality-inspection-profile`,
				`"performance-benchmark"\s*[,}]`,
				`"performance_benchmark"\s*:`,
				`"compliance"\s*:`,
			},
		},
		{
			rel: "fusion-matrix.json",
			patterns: []string{
				`"owner"\s*:\s*"performance-benchmark(\+|")`,
				`"owner"\s*:\s*"[^"]*\+performance-benchmark(\+|")`,
				`"source_pool"\s*:\s*"nuwa(\+|")`,
				`"source_pool"\s*:\s*"[^"]*\+nuwa(\+|")`,
				`quality-inspection-profile`,
			},
		},
		{
			rel: "acceptance-checklists.json",
			patterns: []string{
				`"performance_benchmark"\s*:`,
				`"compliance"\s*:`,
			},
		},
		{
			rel: "purification-charter.json",
			patterns: []string{
				`"seat"\s*:\s*"performance-benchmark"`,
				`"seat"\s*:\s*"compliance"`,
				`"item"\s*:\s*"performance benchmark"`,
				`"item"\s*:\s*"compliance"`,
			},
		},
		{
			rel: "README.md",
			patterns: []string{
				"`质检`",
				`quality-inspection-profile`,
			},
		},
	}
	for _, check := range checks {
		path := filepath.Join(workspace, filepath.FromSlash(check.rel))
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		text := string(raw)
		if check.rel == "tools/wuji_cli.go" {
			if start := strings.Index(text, "func runtimeResidueFailures("); start >= 0 {
				if end := strings.Index(text[start:], "\nfunc fusionAuditCommand("); end > 0 {
					text = text[:start] + text[start+end:]
				}
			}
		}
		for _, pattern := range check.patterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				failures = append(failures, "runtime_residue_bad_pattern="+check.rel+":"+pattern)
				continue
			}
			if re.MatchString(text) {
				failures = append(failures, "runtime_residue="+check.rel+":"+pattern)
			}
		}
	}
	return failures
}

func fusionAuditCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	failures := []string{}
	warnings := []string{}

	kernelPath := kernelSourcePath(workspace)
	if !nonEmpty(kernelPath) {
		failures = append(failures, "missing_kernel_source="+kernelPath)
	}
	kernelObj, kernelErr := loadJSONObject(kernelPath)
	if kernelErr != nil && nonEmpty(kernelPath) {
		failures = append(failures, "kernel_source_unreadable="+kernelPath)
	}

	configPath := filepath.Join(workspace, "config.json")
	configObj, configErr := loadJSONObject(configPath)
	if configErr != nil {
		failures = append(failures, "config_unreadable="+absClean(configPath))
	}
	readmePath := filepath.Join(workspace, "README.md")
	readmeBytes, readmeErr := os.ReadFile(readmePath)
	if readmeErr != nil {
		failures = append(failures, "readme_unreadable="+absClean(readmePath))
	}

	if kernelErr == nil && configErr == nil {
		kernelVersion := objectString(kernelObj, "kernel_version")
		configVersion := objectString(configObj, "iron_rules_version")
		if kernelVersion == "" {
			failures = append(failures, "kernel_version_missing")
		}
		if configVersion != kernelVersion {
			failures = append(failures, "version_drift=config:"+configVersion+" kernel:"+kernelVersion)
		}
		if readmeErr == nil && kernelVersion != "" && !strings.Contains(string(readmeBytes), "v"+kernelVersion) {
			failures = append(failures, "readme_version_drift")
		}
		rules := mergedRoutingRules(configObj)
		failures = append(failures, routeRuleConflictFailures(rules)...)
		enabledProviders := map[string]bool{}
		if providers, ok := objectSlice(configObj, "providers"); ok {
			for _, rawProvider := range providers {
				provider, ok := rawProvider.(map[string]any)
				if !ok {
					continue
				}
				id := objectString(provider, "id")
				if enabled, ok := objectBool(provider, "enabled"); id != "" && ok && enabled {
					enabledProviders[id] = true
				}
			}
		}
		profiles := mergedModelProfiles(configObj)
		for tier, profile := range profiles {
			if profile.ProviderID != "" && !enabledProviders[profile.ProviderID] {
				failures = append(failures, "model_profile_provider_disabled="+tier+":"+profile.ProviderID)
			}
		}
		for _, rule := range rules {
			if rule.ProviderID != "" && !enabledProviders[rule.ProviderID] {
				failures = append(failures, "routing_rule_provider_disabled="+rule.ID+":"+rule.ProviderID)
			}
		}
	}
	failures = append(failures, runtimeResidueFailures(workspace)...)
	if kernelErr == nil {
		residualShellPolicy, _ := objectMap(kernelObj, "residual_shell_policy")
		if objectString(residualShellPolicy, "opensquilla_commander") != "retired-deleted" {
			failures = append(failures, "kernel_opensquilla_commander_not_retired_deleted")
		}
		for _, key := range []string{"external_opensquilla_execution", "opensquilla_gateway_launcher", "opensquilla_mcp_bridge_launcher", "parallel_router"} {
			if objectString(residualShellPolicy, key) != "forbidden" {
				failures = append(failures, "kernel_policy_not_forbidden="+key)
			}
		}
		sensitivePolicy, _ := objectMap(kernelObj, "sensitive_state_policy")
		for _, key := range []string{"user_secrets", "accounts_keys_addresses_tokens_cookies"} {
			if objectString(sensitivePolicy, key) == "" {
				failures = append(failures, "sensitive_state_policy_missing="+key)
			}
		}
		intelContract, ok := objectMap(kernelObj, "intelligence_profile_contract")
		if !ok {
			failures = append(failures, "intelligence_profile_contract_missing")
		} else {
			if objectString(intelContract, "role") != "candidate-scout-not-research-system" {
				failures = append(failures, "intelligence_profile_contract_bad_role")
			}
			if objectString(intelContract, "search_scope") != "wide-recall-shallow-first" {
				failures = append(failures, "intelligence_profile_contract_bad_search_scope")
			}
			if !containsExactString(stringSlice(intelContract, "may_do"), "candidate-metadata") ||
				!containsExactString(stringSlice(intelContract, "may_do"), "evidence-handle") {
				failures = append(failures, "intelligence_profile_contract_missing_may_do")
			}
			for _, forbidden := range []string{"final-analysis", "deep-extract-by-default", "distillation-decision", "adoption-decision", "install-or-execute"} {
				if !containsExactString(stringSlice(intelContract, "must_not_do"), forbidden) {
					failures = append(failures, "intelligence_profile_contract_missing_must_not_do="+forbidden)
				}
			}
			if !containsExactString(stringSlice(intelContract, "candidate_fields"), "source_type") ||
				!containsExactString(stringSlice(intelContract, "candidate_fields"), "next_gate") {
				failures = append(failures, "intelligence_profile_contract_missing_candidate_fields")
			}
		}
		conciseContract, ok := objectMap(kernelObj, "concise_execution_contract")
		if !ok {
			failures = append(failures, "concise_execution_contract_missing")
		} else {
			if objectString(conciseContract, "objective") != "short-precise-high-hit-low-total-cost" {
				failures = append(failures, "concise_execution_contract_bad_objective")
			}
			for _, required := range []string{"single-message-precision", "minimal-needed-context", "prior-art-before-invention-when-uncertain", "fresh-output-uncached-volume-gated"} {
				if !containsExactString(stringSlice(conciseContract, "must_do"), required) {
					failures = append(failures, "concise_execution_contract_missing_must_do="+required)
				}
			}
			for _, forbidden := range []string{"verbose-status-padding", "unneeded-preflight-loop", "context-shift-from-cached-to-uncached", "from-scratch-tooling-when-existing-solution-fits"} {
				if !containsExactString(stringSlice(conciseContract, "must_not_do"), forbidden) {
					failures = append(failures, "concise_execution_contract_missing_must_not_do="+forbidden)
				}
			}
			for _, metric := range []string{"cached_tokens_p95", "fresh_input_tokens_p95", "output_tokens_p95", "uncached_tokens_p95", "tokens_per_success", "retries"} {
				if !containsExactString(stringSlice(conciseContract, "cost_vector"), metric) {
					failures = append(failures, "concise_execution_contract_missing_cost_vector="+metric)
				}
			}
		}
		executionBudgetContractObj, ok := objectMap(kernelObj, "execution_budget_contract")
		if !ok {
			failures = append(failures, "execution_budget_contract_missing")
		} else {
			failures = append(failures, executionBudgetContractFailures(executionBudgetContractObj, "execution_budget_contract_")...)
		}
		optimizationKernel, ok := objectMap(kernelObj, "optimization_kernel")
		if !ok {
			failures = append(failures, "optimization_kernel_missing")
		} else {
			if objectString(optimizationKernel, "runtime_context_gate") != "runtime-context-audit" {
				failures = append(failures, "optimization_kernel_runtime_context_gate_missing")
			}
			runtimePolicy := strings.ToLower(objectString(optimizationKernel, "runtime_context_policy"))
			for _, marker := range []string{"numeric", "raw", "prompt", "messages", "content"} {
				if !strings.Contains(runtimePolicy, marker) {
					failures = append(failures, "optimization_kernel_runtime_context_policy_missing="+marker)
				}
			}
		}
		if !containsExactString(stringSlice(kernelObj, "required_audits"), "runtime-context-audit-for-token-cost-cache-usage-claims") {
			failures = append(failures, "required_audit_missing=runtime-context-audit-for-token-cost-cache-usage-claims")
		}
		analysisContract, ok := objectMap(kernelObj, "analysis_completeness_contract")
		if !ok {
			failures = append(failures, "analysis_completeness_contract_missing")
		} else {
			failures = append(failures, analysisCompletenessContractFailures(analysisContract, "analysis_completeness_contract_")...)
		}
		distilledKernel, ok := objectMap(kernelObj, "distilled_atom_kernel")
		if !ok {
			failures = append(failures, "distilled_atom_kernel_missing")
		} else {
			for _, key := range []string{"resident_light_atoms", "on_demand_atoms", "owner_map"} {
				if _, present := distilledKernel[key]; !present {
					failures = append(failures, "distilled_atom_kernel_missing="+key)
				}
			}
			residentAtoms := stringSlice(distilledKernel, "resident_light_atoms")
			onDemandAtoms := stringSlice(distilledKernel, "on_demand_atoms")
			if len(residentAtoms)+len(onDemandAtoms) != expectedDistilledAtomCount {
				failures = append(failures, "kernel_distilled_atom_count="+strconv.Itoa(len(residentAtoms)+len(onDemandAtoms))+"!="+strconv.Itoa(expectedDistilledAtomCount))
			}
			policy := strings.ToLower(objectString(distilledKernel, "policy"))
			for _, marker := range []string{"replace", "existing 21", "never stack"} {
				if !strings.Contains(policy, marker) {
					failures = append(failures, "distilled_atom_kernel_policy_missing="+marker)
				}
			}
		}
		if !containsExactString(stringSlice(kernelObj, "source_pools"), "github-trending-20260608-style") {
			failures = append(failures, "kernel_source_pool_missing=github-trending-20260608-style")
		}
	}
	failures = append(failures, mixedDistilledSourceLineageAtomFailures()...)
	failures = append(failures, routeDistilledAtomCoverageFailures()...)
	failures = append(failures, distilledAtomRegistryFailures()...)

	opensquillaRuntimeSurfaces := []string{
		`C:\Users\Administrator\.codex\skills\opensquilla-commander`,
		`C:\Users\Administrator\.codex\skills\opensquilla-commander\SKILL.md`,
		`C:\Users\Administrator\.codex\skills\opensquilla-commander\scripts\opensquilla-command.ps1`,
		`C:\Users\Administrator\.codex\skills\opensquilla-commander\scripts\opensquilla-mcp-bridge.ps1`,
		`C:\Users\Administrator\.agents\skills\opensquilla-commander`,
		`C:\Users\Administrator\.agents\skills\opensquilla-commander\SKILL.md`,
		`C:\Users\Administrator\.agents\skills\opensquilla-commander\scripts\opensquilla-command.ps1`,
		`C:\Users\Administrator\.agents\skills\opensquilla-commander\scripts\opensquilla-mcp-bridge.ps1`,
	}
	failures = append(failures, opensquillaRuntimeSurfaceFailures(opensquillaRuntimeSurfaces)...)
	failures = append(failures, activeLegacySkillSurfaceFailures([]string{
		`C:\Users\Administrator\.agents\skills\_backup-wuji-legion-20260606-015130`,
		`C:\Users\Administrator\.agents\skills\_backup-wuji-legion-20260606-015130\SKILL.md`,
		`C:\Users\Administrator\.agents\skills\wuji-legion-codex`,
		`C:\Users\Administrator\.agents\skills\wuji-legion-codex\SKILL.md`,
		`C:\Users\Administrator\.codex\skills\wuji-legion-codex`,
		`C:\Users\Administrator\.codex\skills\wuji-legion-codex\SKILL.md`,
	})...)
	failures = append(failures, nestedExampleSkillSurfaceFailures([]string{
		`C:\Users\Administrator\.agents\skills\huashu-nuwa\examples\*\SKILL.md`,
		`C:\Users\Administrator\.codex\skills\huashu-nuwa\examples\*\SKILL.md`,
	})...)

	if info, err := os.Stat(filepath.Join(workspace, "kernel-source.json")); err == nil && info.IsDir() {
		failures = append(failures, "kernel_source_must_be_file")
	}
	fusionMatrixPath := filepath.Join(workspace, "fusion-matrix.json")
	if !nonEmpty(fusionMatrixPath) {
		failures = append(failures, "missing_fusion_matrix="+absClean(fusionMatrixPath))
	}
	if fusionMatrixObj, err := loadJSONObject(fusionMatrixPath); err == nil {
		if !containsExactString(stringSlice(fusionMatrixObj, "source_pools"), "github-trending-20260608-style") {
			failures = append(failures, "fusion_matrix_source_pool_missing=github-trending-20260608-style")
		}
		if decisions, ok := fusionMatrixObj["decisions"].([]any); !ok || len(decisions) == 0 {
			failures = append(failures, "fusion_matrix_has_no_decisions")
		} else {
			requiredRetiredAtoms := map[string]bool{
				"opensquilla-external-executor": false,
				"opensquilla-commander-skill":   false,
				"opensquilla-gateway-launcher":  false,
				"parallel-router-shell":         false,
			}
			requiredDistilledAtoms := distilledAtomPresenceMap()
			sourcePoolCatalog := fusionSourcePoolCatalog(kernelObj, fusionMatrixObj)
			sourcePoolCatalogFailures := []string{}
			seenTrendingMarkers := map[string]bool{
				"guarded-realtime-source-search:github-trending-20260608-style": false,
				"research-evidence-pack:open-notebook":                          false,
				"research-evidence-pack:tolaria":                                false,
				"anti-ai-slop-visual-rules:taste-skill":                         false,
				"anti-ai-slop-visual-rules:opencv-style":                        false,
				"data-large-file-workflow:turbovec-style":                       false,
				"terminal-real-run-verification:pg_durable-style":               false,
			}
			riskSurfaceRejected := false
			for _, rawDecision := range decisions {
				decision, ok := rawDecision.(map[string]any)
				if !ok {
					continue
				}
				atom := objectString(decision, "atom")
				decisionValue := objectString(decision, "decision")
				sourcePool := objectString(decision, "source_pool")
				sourcePoolCatalogFailures = append(sourcePoolCatalogFailures, sourcePoolFailures(atom, sourcePool, sourcePoolCatalog)...)
				if isActiveFusionDecision(decisionValue) && !distilledAtomKnownMap()[atom] && activeFusionDecisionLooksLikeRuntimeAtom(atom, sourcePool) {
					failures = append(failures, "fusion_matrix_extra_active_atom="+atom)
				}
				if fusionDecisionNeedsReject(atom, sourcePool, objectString(decision, "reason"), objectString(decision, "fusion_policy")) && decisionValue != "reject" {
					failures = append(failures, "fusion_matrix_risk_surface_not_reject="+atom)
				}
				if atom == "github-trending-risk-surfaces" && decisionValue == "reject" {
					riskSurfaceRejected = true
				}
				for key := range seenTrendingMarkers {
					parts := strings.SplitN(key, ":", 2)
					if len(parts) == 2 && atom == parts[0] && strings.Contains(sourcePool, parts[1]) {
						seenTrendingMarkers[key] = true
					}
				}
				if _, required := requiredRetiredAtoms[atom]; required {
					if decisionValue == "retire" {
						requiredRetiredAtoms[atom] = true
					}
				}
				if _, required := requiredDistilledAtoms[atom]; required {
					if decisionValue == "resident" || decisionValue == "mount-on-demand" || decisionValue == "replace" {
						requiredDistilledAtoms[atom] = true
					}
				}
			}
			failures = append(failures, sourcePoolCatalogFailures...)
			for atom, retired := range requiredRetiredAtoms {
				if !retired {
					failures = append(failures, "fusion_matrix_atom_not_retired="+atom)
				}
			}
			for atom, present := range requiredDistilledAtoms {
				if !present {
					failures = append(failures, "fusion_matrix_distilled_atom_missing="+atom)
				}
			}
			for marker, seen := range seenTrendingMarkers {
				if !seen {
					failures = append(failures, "fusion_matrix_trending_marker_missing="+marker)
				}
			}
			if !riskSurfaceRejected {
				failures = append(failures, "fusion_matrix_github_trending_risk_surfaces_not_rejected")
			}
		}
	} else if fileExists(fusionMatrixPath) {
		failures = append(failures, "fusion_matrix_unreadable="+absClean(fusionMatrixPath))
	}
	residualPath := filepath.Join(workspace, "residual-entrypoints.json")
	if !nonEmpty(residualPath) {
		failures = append(failures, "missing_residual_entrypoints="+absClean(residualPath))
	}
	if residualObj, err := loadJSONObject(residualPath); err == nil {
		if entries, ok := residualObj["entries"].([]any); !ok || len(entries) == 0 {
			failures = append(failures, "residual_entrypoints_empty")
		} else {
			allowedStatuses := map[string]bool{
				"main-chain":                        true,
				"on-demand":                         true,
				"fuse-into-kernel":                  true,
				"retired-deleted":                   true,
				"retire-and-label":                  true,
				"delete-now":                        true,
				"delete-now-except-latest-evidence": true,
				"delete-when-not-building":          true,
				"excluded":                          true,
			}
			coveragePatterns := []string{}
			coverageStatus := map[string]string{}
			for _, rawEntry := range entries {
				entry, ok := rawEntry.(map[string]any)
				if !ok {
					continue
				}
				entryPath := objectString(entry, "path")
				status := objectString(entry, "status")
				if !allowedStatuses[status] {
					failures = append(failures, "residual_entrypoint_unknown_status="+entryPath+":"+status)
				}
				if status == "compat-only" {
					failures = append(failures, "residual_entrypoint_compat_only_forbidden="+entryPath)
				}
				if strings.Contains(strings.ToLower(entryPath), "opensquilla-commander") && status != "retired-deleted" {
					failures = append(failures, "opensquilla_commander_residual_not_retired_deleted="+entryPath)
				}
				if pattern, local := workspaceEntryPattern(workspace, entryPath); local && pattern != "" && !isRetiredResidualStatus(status) {
					coveragePatterns = append(coveragePatterns, pattern)
					coverageStatus[pattern] = status
				}
			}
			inventory, inventoryErr := workspaceFileInventory(workspace)
			if inventoryErr != nil {
				failures = append(failures, "workspace_inventory_unreadable="+absClean(workspace))
			} else {
				uncovered := []string{}
				for _, rel := range inventory {
					covered := false
					for _, pattern := range coveragePatterns {
						if slashPatternMatch(pattern, rel) {
							covered = true
							break
						}
					}
					if !covered {
						uncovered = append(uncovered, rel)
					}
				}
				if len(uncovered) > 0 {
					failures = append(failures, "residual_inventory_uncovered="+strings.Join(uncovered, "|"))
				}
				for _, pattern := range coveragePatterns {
					if !strings.ContainsAny(pattern, "*?[") {
						found := false
						for _, rel := range inventory {
							if slashPatternMatch(pattern, rel) {
								found = true
								break
							}
						}
						if !found {
							failures = append(failures, "residual_entrypoint_missing_current_file="+pattern+":"+coverageStatus[pattern])
						}
					}
				}
			}
		}
	} else if fileExists(residualPath) {
		failures = append(failures, "residual_entrypoints_unreadable="+absClean(residualPath))
	}
	checklistsPath := filepath.Join(workspace, "acceptance-checklists.json")
	if !nonEmpty(checklistsPath) {
		failures = append(failures, "missing_acceptance_checklists="+absClean(checklistsPath))
	}
	if checklistObj, err := loadJSONObject(checklistsPath); err == nil {
		for _, key := range []string{"white_hat", "guard_office", "root_cause_officer", "audit", "quality_inspection", "performance_benchmark_on_demand", "compliance_on_demand"} {
			if items, ok := checklistObj[key].([]any); !ok || len(items) == 0 {
				failures = append(failures, "acceptance_checklist_empty="+key)
			}
		}
	} else if fileExists(checklistsPath) {
		failures = append(failures, "acceptance_checklists_unreadable="+absClean(checklistsPath))
	}
	purityCharterPath := filepath.Join(workspace, "purification-charter.json")
	if !nonEmpty(purityCharterPath) {
		failures = append(failures, "missing_purification_charter="+absClean(purityCharterPath))
	} else if charterObj, err := loadJSONObject(purityCharterPath); err == nil {
		for _, key := range []string{"main_chain", "keep_on_demand", "fuse_into_kernel", "retire_and_label", "delete_now", "hard_gates"} {
			if items, ok := charterObj[key].([]any); !ok || len(items) == 0 {
				failures = append(failures, "purification_charter_empty="+key)
			}
		}
	} else {
		failures = append(failures, "purification_charter_unreadable="+absClean(purityCharterPath))
	}

	mirrorChecks := []struct {
		path    string
		markers []string
	}{
		{filepath.Join(workspace, "SKILL.md"), []string{"kernel-source.json", "task-routing", "capability-mount", "deterministic-execution", "distilled_atom_kernel", "assumption-ledger", "version-doc-mcp", "root-cause-radar", "root-cause officer", "Closeout Sound"}},
		{filepath.Join(workspace, "GLOBAL_AGENTS.md"), []string{"kernel-source.json", "fusion-audit", "optimization-audit", "distilled_atom_kernel", "claim-fact-check", "prior-art-solution-search", "root-cause-radar", "根因雷达官", "terminal-real-run-verification", "Closeout Sound"}},
		{filepath.Join(workspace, "README.md"), []string{"kernel-source.json", "fusion-audit", "optimization-audit", "distilled_atom_kernel", "reversible-evidence-handle", "patch-debt-root-cure", "root-cause officer", "Closeout Sound"}},
		{filepath.Join(workspace, "experts", "INDEX.md"), []string{"kernel-source.json", "fusion-matrix.json", "residual-entrypoints.json"}},
		{filepath.Join(workspace, "units", "context_router.md"), []string{"kernel-source.json", "task-routing", "capability-mount", "deterministic-execution", "distilled_atoms", "source_lineage_atoms"}},
		{filepath.Join(workspace, "units", "execution_base.md"), []string{"kernel-source.json", "deterministic-execution", "wuji-cli", "terminal-real-run-verification", "scripts/beep.ps1", "goose", "pg_durable-style", "openai-plugins"}},
		{filepath.Join(workspace, "units", "staff.md"), []string{"kernel-source.json", "task-routing", "capability-mount", "minimal-gap-first", "distilled_atoms", "source_lineage_atoms", "parallel-hypothesis-fanout"}},
		{filepath.Join(workspace, "units", "dev.md"), []string{"root-cause-radar", "parallel-hypothesis-fanout", "terminal-real-run-verification"}},
		{filepath.Join(workspace, "units", "expedition.md"), []string{"parallel-hypothesis-fanout", "单一主链"}},
		{filepath.Join(workspace, "units", "auto_evolve.md"), []string{"patch-debt-root-cure", "不打补丁"}},
		{filepath.Join(workspace, "units", "mcp_plugins.md"), []string{"kernel-source.json", "capability-mount", "do not let MCP or plugins become a second router"}},
		{filepath.Join(workspace, "units", "distillation.md"), []string{"kernel-source.json", "fusion-matrix.json", "residual-entrypoints.json", "distilled_atom_kernel", "version-doc-mcp", "verified-learning-loop", "root-cause-radar", "github-trending-20260608-style", "existing 21"}},
		{filepath.Join(workspace, "units", "oversight.md"), []string{"kernel-source.json", "white-hat", "root-cause-officer", "audit", "质检", "assumption-ledger", "claim-fact-check", "distilled_atoms", "terminal-real-run-verification"}},
		{filepath.Join(workspace, "units", "security.md"), []string{"kernel-source.json", "guard-office", "security", "compliance-on-demand", "GhostTrack", "ChinaTextbook", "AiToEarn", "project-nomad"}},
		{filepath.Join(workspace, "units", "intel.md"), []string{"候选侦察", "GitHub", "候选卡片", "证据句柄", "不做最终分析", "待主链裁决", "last30days", "open-notebook", "tolaria", "openai-plugins"}},
		{filepath.Join(workspace, "units", "data.md"), []string{"turbovec-style", "open-notebook", "tolaria", "Spreadsheets", "不把整表整库回灌"}},
		{filepath.Join(workspace, "experts", "oversight", "白帽纠察官.md"), []string{"assumption-ledger", "claim-fact-check"}},
		{filepath.Join(workspace, "experts", "oversight", "根因雷达官.md"), []string{"root-cause-radar", "parallel-hypothesis-fanout", "patch-debt-root-cure", "terminal-real-run-verification"}},
		{filepath.Join(workspace, "experts", "oversight", "审计官.md"), []string{"research-evidence-pack", "reversible-evidence-handle"}},
		{filepath.Join(workspace, "experts", "oversight", "质检官.md"), []string{"claim-fact-check", "disciplined-debug-loop", "taste-skill", "opencv-style"}},
		{filepath.Join(workspace, "experts", "security", "保卫科.md"), []string{"guarded-realtime-source-search", "version-doc-mcp", "GhostTrack", "ChinaTextbook", "AiToEarn", "project-nomad"}},
		{filepath.Join(workspace, "experts", "staff", "参谋主帅.md"), []string{"distilled_atoms", "source_lineage_atoms"}},
		{filepath.Join(workspace, "experts", "dev", "开发主帅.md"), []string{"version-doc-mcp", "disciplined-debug-loop"}},
		{filepath.Join(workspace, "experts", "intel", "情报主帅.md"), []string{"guarded-realtime-source-search", "research-evidence-pack", "prior-art-solution-search", "候选来源卡片", "证据句柄", "不可以：做最终分析", "last30days", "open-notebook", "tolaria", "openai-plugins"}},
		{filepath.Join(workspace, "experts", "data", "数据主帅.md"), []string{"turbovec-style", "open-notebook", "tolaria", "content-type-compression-router", "reversible-evidence-handle"}},
		{filepath.Join(workspace, "experts", "execution_base", "执行底座主帅.md"), []string{"goose", "pg_durable-style", "openai-plugins", "durable checkpoint"}},
		{filepath.Join(workspace, "experts", "evolve", "进化主帅.md"), []string{"skill-stocktake-daily-library", "verified-learning-loop", "github-trending-20260608-style", "既有 21"}},
		{`C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md`, []string{"kernel-source.json", "task-routing", "capability-mount", "deterministic-execution", "distilled_atom_kernel", "assumption-ledger", "version-doc-mcp", "root-cause-officer", "terminal-real-run-verification", "Closeout Sound"}},
	}
	mirrorChecks = append(mirrorChecks,
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "units", "execution_base.md"), []string{"execution_budget_contract", "LIGHT_TASK", "runtime-context-audit"}},
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "units", "staff.md"), []string{"execution_budget_contract", "LIGHT_TASK", "STRUCTURAL_TASK"}},
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "units", "oversight.md"), []string{"execution_budget_contract", "LIGHT_TASK", "runtime-context-audit"}},
	)
	mirrorChecks = append(mirrorChecks,
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "SKILL.md"), []string{"html-native-design-canvas", "anti-ai-slop-visual-rules"}},
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "units", "mcp_plugins.md"), []string{"huashu-design", "html-native-design-canvas"}},
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "units", "visual.md"), []string{"kernel-source.json", "huashu-design", "html-native-design-canvas", "anti-ai-slop-visual-rules", "html-deck-to-editable-pptx", "taste-skill", "opencv-style"}},
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "units", "html_slides_master.md"), []string{"kernel-source.json", "huashu-design", "design-direction-triad", "motion-stage-sprite-engine"}},
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "experts", "visual", "视觉主帅.md"), []string{"huashu-design", "html-native-design-canvas", "anti-ai-slop-visual-rules", "motion-stage-sprite-engine", "taste-skill", "opencv-style"}},
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "experts", "oversight", "质检官.md"), []string{"anti-ai-slop-visual-rules", "html-deck-to-editable-pptx"}},
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "experts", "security", "保卫科.md"), []string{"brand-asset-protocol"}},
		struct {
			path    string
			markers []string
		}{`C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md`, []string{"html-native-design-canvas", "anti-ai-slop-visual-rules"}},
	)
	mirrorChecks = append(mirrorChecks,
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "SKILL.md"), []string{"hotpath-manifest.json", "context-bloat-audit", "runtime-context-audit", "concise_execution_contract", "execution_budget_contract"}},
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "GLOBAL_AGENTS.md"), []string{"hotpath-manifest.json", "context-bloat-audit", "runtime-context-audit", "concise_execution_contract", "execution_budget_contract"}},
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "README.md"), []string{"hotpath-manifest.json", "context-bloat-audit", "runtime-context-audit", "concise_execution_contract", "execution_budget_contract"}},
		struct {
			path    string
			markers []string
		}{`C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md`, []string{"hotpath-manifest.json", "context-bloat-audit", "runtime-context-audit", "concise_execution_contract", "execution_budget_contract"}},
	)
	mirrorChecks = append(mirrorChecks,
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "SKILL.md"), []string{"analysis_completeness_contract", "complete-materials-before-architecture-analysis"}},
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "GLOBAL_AGENTS.md"), []string{"analysis_completeness_contract", "complete-materials-before-architecture-analysis"}},
		struct {
			path    string
			markers []string
		}{filepath.Join(workspace, "README.md"), []string{"analysis_completeness_contract", "complete-materials-before-architecture-analysis"}},
		struct {
			path    string
			markers []string
		}{`C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md`, []string{"analysis_completeness_contract", "complete-materials-before-architecture-analysis"}},
	)
	for _, check := range mirrorChecks {
		failures = append(failures, mirrorDriftFailures(check.path, check.markers)...)
	}

	report := jsonObject{
		"workspace_key":         privacyHash(absClean(workspace)),
		"kernel_source":         pathPrivacyRef(workspace, kernelPath),
		"fusion_matrix":         pathPrivacyRef(workspace, fusionMatrixPath),
		"residual_entrypoints":  pathPrivacyRef(workspace, residualPath),
		"acceptance_checklists": pathPrivacyRef(workspace, checklistsPath),
		"purification_charter":  pathPrivacyRef(workspace, purityCharterPath),
		"config":                pathPrivacyRef(workspace, configPath),
		"readme":                pathPrivacyRef(workspace, readmePath),
		"checks": []string{
			"single-version-source",
			"kernel-config-alignment",
			"readme-version-alignment",
			"opensquilla-runtime-surface-absent",
			"fusion-matrix-present",
			"residual-entrypoints-labeled",
			"acceptance-checklists-present",
			"purification-charter-present",
			"mirror-doc-drift-detection",
		},
		"warnings": warnings,
		"status":   ternaryStatus(len(failures) == 0, "pass", "fail"),
	}
	manifest := auditManifest(workspace, "fusion-audit", []string{
		"kernel-source.json",
		"config.json",
		"fusion-matrix.json",
		"residual-entrypoints.json",
		"acceptance-checklists.json",
		"purification-charter.json",
		"hotpath-manifest.json",
		"README.md",
		"tools/wuji_cli.go",
		"scripts/beep.ps1",
		"SKILL.md",
		"GLOBAL_AGENTS.md",
		"experts/INDEX.md",
		"experts/oversight/白帽纠察官.md",
		"experts/oversight/根因雷达官.md",
		"experts/oversight/审计官.md",
		"experts/oversight/质检官.md",
		"experts/oversight/性能基准官.md",
		"experts/security/保卫科.md",
		"experts/security/合规审计官.md",
		"experts/security/安全主帅.md",
		"experts/staff/参谋主帅.md",
		"units/context_router.md",
		"units/execution_base.md",
		"units/staff.md",
		"units/dev.md",
		"units/expedition.md",
		"units/auto_evolve.md",
		"units/mcp_plugins.md",
		"units/visual.md",
		"units/html_slides_master.md",
		"units/distillation.md",
		"units/oversight.md",
		"units/security.md",
		"units/intel.md",
	}, `C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md`)
	for key, value := range manifest {
		report[key] = value
	}
	if len(failures) > 0 {
		report["failures"] = failures
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(workspace, "outputs", "fusion-audit-report.json")
	}
	if err := writeJSON(outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return printGate("fusion-audit", failures)
}

func optimizationAuditCommand(args []string) int {
	workspace, ok := argValue(args, "--workspace")
	if !ok {
		usage()
		return 2
	}
	reportPath, hasReport := argValue(args, "--report")
	failures := []string{}
	warnings := []string{}

	configPath := filepath.Join(workspace, "config.json")
	configObj, configErr := loadJSONObject(configPath)
	if configErr != nil {
		failures = append(failures, "config_unreadable="+absClean(configPath))
	}
	contextPackPath := filepath.Join(workspace, "outputs", "context-pack-rich.json")
	if !nonEmpty(contextPackPath) {
		contextPackPath = filepath.Join(workspace, "outputs", "context-pack.json")
	}
	contextPackBytes := fileSize(contextPackPath)
	contextObj, contextErr := loadJSONObject(contextPackPath)
	if contextErr != nil {
		failures = append(failures, "context_pack_unreadable="+absClean(contextPackPath))
	}
	if contextPackBytes > maxOptimizationContextPackBytes {
		failures = append(failures, fmt.Sprintf("context_pack_over_budget=%s:%d>%d", absClean(contextPackPath), contextPackBytes, maxOptimizationContextPackBytes))
	}
	richContextPackPath := filepath.Join(workspace, "outputs", "context-pack-rich.json")
	staleContextPackPath := filepath.Join(workspace, "outputs", "context-pack.json")
	if fileExists(richContextPackPath) && fileExists(staleContextPackPath) {
		failures = append(failures, "stale_context_pack_json_present="+absClean(staleContextPackPath))
	}
	outputsPath := filepath.Join(workspace, "outputs")
	outputFiles, outputBytes, outputStatsErr := directoryStats(outputsPath)
	if outputStatsErr != nil {
		failures = append(failures, "outputs_stats_unreadable="+absClean(outputsPath))
	}
	if outputBytes > maxOptimizationOutputsBytes {
		failures = append(failures, fmt.Sprintf("outputs_over_budget=%s:%d>%d", absClean(outputsPath), outputBytes, maxOptimizationOutputsBytes))
	}
	if outputFiles > maxOptimizationOutputsFiles {
		failures = append(failures, fmt.Sprintf("outputs_file_count_over_budget=%s:%d>%d", absClean(outputsPath), outputFiles, maxOptimizationOutputsFiles))
	}
	toolsPath := filepath.Join(workspace, ".wuji-tools")
	toolsFiles, toolsBytes, toolsStatsErr := directoryStats(toolsPath)
	if toolsStatsErr != nil {
		failures = append(failures, "wuji_tools_stats_unreadable="+absClean(toolsPath))
	}
	if toolsBytes > maxOptimizationToolsBytes {
		failures = append(failures, fmt.Sprintf("wuji_tools_over_budget=%s:%d>%d", absClean(toolsPath), toolsBytes, maxOptimizationToolsBytes))
	}
	if toolsFiles > maxOptimizationToolsFiles {
		failures = append(failures, fmt.Sprintf("wuji_tools_file_count_over_budget=%s:%d>%d", absClean(toolsPath), toolsFiles, maxOptimizationToolsFiles))
	}

	if configErr == nil {
		cacheConfig, _ := objectMap(configObj, "cache_config")
		if objectString(cacheConfig, "stable_prefix_policy") == "" {
			failures = append(failures, "stable_prefix_policy_missing")
		}
		if objectString(cacheConfig, "optimization_objective") == "" {
			failures = append(failures, "optimization_objective_missing")
		}
		if objectString(cacheConfig, "concise_execution_policy") == "" {
			failures = append(failures, "concise_execution_policy_missing")
		}
	}
	if contextErr == nil {
		stablePrefix, _ := objectMap(contextObj, "stable_prefix")
		if len(stablePrefix) > maxOptimizationStablePrefixFields {
			failures = append(failures, fmt.Sprintf("stable_prefix_field_count_over_budget=%d>%d", len(stablePrefix), maxOptimizationStablePrefixFields))
		}
		if objectString(stablePrefix, "stable_prefix_policy") == "" {
			failures = append(failures, "context_pack_missing_stable_prefix_policy")
		}
		prefixCanon, ok := objectMap(contextObj, "stable_prefix_canon")
		if !ok {
			failures = append(failures, "context_pack_missing_prefix_canon")
		} else {
			if _, ok := prefixCanon["canon_text"]; ok {
				failures = append(failures, "context_pack_prefix_canon_text_over_budget")
			}
			if _, ok := prefixCanon["ordered_fields"]; ok {
				failures = append(failures, "context_pack_prefix_canon_ordered_fields_over_budget")
			}
			if objectString(prefixCanon, "canon_hash") == "" {
				failures = append(failures, "context_pack_prefix_canon_hash_missing")
			}
		}
		conciseContract, ok := objectMap(contextObj, "concise_execution_contract")
		if !ok || objectString(conciseContract, "objective") != "short-precise-high-hit-low-total-cost" {
			failures = append(failures, "context_pack_missing_concise_execution_contract")
		}
		executionBudgetContractObj, ok := objectMap(contextObj, "execution_budget_contract")
		if !ok {
			failures = append(failures, "context_pack_missing_execution_budget_contract")
		} else {
			failures = append(failures, executionBudgetContractFailures(executionBudgetContractObj, "context_pack_execution_budget_contract_")...)
		}
		routeSummary, _ := objectMap(contextObj, "route_summary")
		if _, ok := routeSummary["execution_budget"]; !ok {
			failures = append(failures, "context_pack_missing_execution_budget")
		}
		if objectBoolValue(routeSummary, "analysis_required") {
			analysisContract, ok := objectMap(contextObj, "analysis_completeness_contract")
			if !ok {
				failures = append(failures, "context_pack_missing_analysis_completeness_contract")
			} else {
				failures = append(failures, analysisCompletenessContractFailures(analysisContract, "context_pack_analysis_completeness_contract_")...)
			}
		}
		dynamicContext, _ := objectMap(contextObj, "dynamic_context")
		distilledAtoms, _ := dynamicContext["distilled_atoms"].([]any)
		if len(distilledAtoms) == 0 {
			failures = append(failures, "distilled_atoms_missing")
		}
		if _, ok := dynamicContext["execution_summaries"]; !ok {
			failures = append(failures, "execution_summaries_missing")
		}
		if _, ok := dynamicContext["audit_summaries"]; !ok {
			failures = append(failures, "audit_summaries_missing")
		}
		reviewGates, _ := contextObj["review_gates"].([]any)
		if len(reviewGates) == 0 {
			failures = append(failures, "review_gates_missing")
		}
		if optimizationPolicy, ok := objectMap(contextObj, "optimization_policy"); ok {
			if objectString(optimizationPolicy, "objective") == "" {
				failures = append(failures, "optimization_policy_objective_missing")
			}
		}
		artifactSummaries, _ := contextObj["artifact_summaries"].([]any)
		if len(artifactSummaries) == 0 {
			failures = append(failures, "artifact_summaries_empty")
		}
		for index, rawSummary := range artifactSummaries {
			summary, ok := rawSummary.(map[string]any)
			if !ok {
				failures = append(failures, fmt.Sprintf("artifact_summary_%d_invalid", index+1))
				continue
			}
			if objectString(summary, "evidence_handle") == "" {
				failures = append(failures, fmt.Sprintf("artifact_summary_%d_missing_evidence_handle", index+1))
			}
			mode := objectString(summary, "summary_mode")
			kind := objectString(summary, "kind")
			if mode == "handle-only" && kind != "binary" {
				failures = append(failures, fmt.Sprintf("artifact_summary_%d_text_handle_only=%s", index+1, objectString(summary, "path_ref")))
			}
		}
	}
	hotpathManifestPath := filepath.Join(workspace, "hotpath-manifest.json")
	if !nonEmpty(hotpathManifestPath) {
		failures = append(failures, "hotpath_manifest_missing="+absClean(hotpathManifestPath))
	} else if hotpathObj, err := loadJSONObject(hotpathManifestPath); err != nil {
		failures = append(failures, "hotpath_manifest_unreadable="+absClean(hotpathManifestPath))
	} else {
		for _, key := range []string{"resident", "on_demand", "cold_ledger", "forbidden_resident"} {
			if items, ok := objectSlice(hotpathObj, key); !ok || len(items) == 0 {
				failures = append(failures, "hotpath_manifest_empty="+key)
			}
		}
		if !hotpathColdLedgerHandleOnly(hotpathObj, "outputs/runtime-context-audit-report.json") {
			failures = append(failures, "hotpath_manifest_missing_runtime_context_audit_handle")
		}
		for _, marker := range []string{"outputs/runtime-usage.jsonl", "raw prompt", "messages", "content"} {
			if !hotpathForbiddenContains(hotpathObj, marker) {
				failures = append(failures, "hotpath_manifest_forbidden_runtime_surface_missing="+marker)
			}
		}
	}
	checklistsPath := filepath.Join(workspace, "acceptance-checklists.json")
	if checklistObj, err := loadJSONObject(checklistsPath); err == nil {
		if items, ok := checklistObj["quality_inspection"].([]any); !ok || len(items) == 0 {
			failures = append(failures, "quality_inspection_acceptance_checklist_missing")
		}
		if items, ok := checklistObj["white_hat"].([]any); !ok || len(items) == 0 {
			failures = append(failures, "white_hat_acceptance_checklist_missing")
		}
		if !checklistContainsMarker(checklistObj, "runtime-context-audit") {
			failures = append(failures, "acceptance_checklists_missing_runtime_context_audit")
		}
	} else {
		failures = append(failures, "acceptance_checklists_unreadable="+absClean(checklistsPath))
	}

	report := jsonObject{
		"workspace_key":         privacyHash(absClean(workspace)),
		"context_pack":          pathPrivacyRef(workspace, contextPackPath),
		"acceptance_checklists": pathPrivacyRef(workspace, checklistsPath),
		"checks": []string{
			"stable-small-prefix",
			"context-pack-byte-budget",
			"outputs-scan-budget",
			"hotpath-manifest-present",
			"typed-lightweight-assembly",
			"evidence-preserved-without-full-replay",
			"context-bloat-audit-required-for-token-optimization",
			"runtime-context-audit-required-for-token-cost-cache-usage-claims",
			"anti-token-overoptimization-gate",
			"acceptance-checklists-present",
		},
		"budgets": jsonObject{
			"context_pack_max_bytes":          maxOptimizationContextPackBytes,
			"context_pack_bytes":              contextPackBytes,
			"stable_prefix_max_fields":        maxOptimizationStablePrefixFields,
			"outputs_max_bytes":               maxOptimizationOutputsBytes,
			"outputs_bytes":                   outputBytes,
			"outputs_max_files":               maxOptimizationOutputsFiles,
			"outputs_files":                   outputFiles,
			"wuji_tools_max_bytes":            maxOptimizationToolsBytes,
			"wuji_tools_bytes":                toolsBytes,
			"wuji_tools_max_files":            maxOptimizationToolsFiles,
			"wuji_tools_files":                toolsFiles,
			"stale_context_pack_forbidden":    true,
			"large_artifacts_policy":          "handle-or-summary-only",
			"anti_overoptimization_condition": "must preserve evidence, audit boundaries, and first-pass hit rate",
		},
		"warnings": warnings,
		"status":   ternaryStatus(len(failures) == 0, "pass", "fail"),
	}
	manifest := auditManifest(workspace, "optimization-audit", []string{
		"config.json",
		"acceptance-checklists.json",
		"outputs/context-pack-rich.json",
		"hotpath-manifest.json",
		"tools/wuji_cli.go",
	})
	for key, value := range manifest {
		report[key] = value
	}
	if len(failures) > 0 {
		report["failures"] = failures
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(workspace, "outputs", "optimization-audit-report.json")
	}
	if err := writeJSON(outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return printGate("optimization-audit", failures)
}

func canonReportCommand(args []string) int {
	reportPath, hasReport := argValue(args, "--report")
	if !hasReport {
		reportPath = filepath.Join(".", "wuji-canon-report.json")
	}
	if err := writeJSON(reportPath, builtinCanonReport()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("GO canon-report\n- report=%s\n", absClean(reportPath))
	return 0
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
	preferKeys := hashStrings(preferTerms)
	avoidKeys := hashStrings(avoidTerms)
	entry := jsonObject{
		"task_key":            privacyHash(task),
		"prefer_signal_keys":  preferKeys,
		"avoid_signal_keys":   avoidKeys,
		"prefer_signal_count": len(preferKeys),
		"avoid_signal_count":  len(avoidKeys),
		"note_key":            privacyHash(note),
		"note_present":        strings.TrimSpace(note) != "",
		"source":              source,
		"logged_at":           time.Now().Format(time.RFC3339),
		"privacy_mode":        "hash-only-no-user-text",
	}
	if err := appendJSONLine(logPath, entry); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	report := jsonObject{
		"log_ref":             pathPrivacyRef(workspace, logPath),
		"task_key":            entry["task_key"],
		"prefer_signal_count": entry["prefer_signal_count"],
		"avoid_signal_count":  entry["avoid_signal_count"],
		"source":              source,
		"privacy_mode":        entry["privacy_mode"],
	}
	if hasReport {
		if err := writeJSON(reportPath, report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	fmt.Printf("GO feedback-log\n- log=feedback/feedback-log.jsonl\n")
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
	taskCounts := map[string]int{}
	taskPreferCounts := map[string]int{}
	taskAvoidCounts := map[string]int{}
	for index, record := range records {
		preferKeys := uniqueStrings(stringSlice(record, "prefer_signal_keys"))
		avoidKeys := uniqueStrings(stringSlice(record, "avoid_signal_keys"))
		task := objectString(record, "task_key")
		if task == "" {
			task = objectString(record, "task")
		}
		if len(preferKeys) == 0 && len(avoidKeys) == 0 {
			continue
		}
		taskCounts[task]++
		taskPreferCounts[task] += len(preferKeys)
		taskAvoidCounts[task] += len(avoidKeys)
		caseID := fmt.Sprintf("feedback-%02d", index+1)
		cases = append(cases, jsonObject{
			"id":                    caseID,
			"task_key":              task,
			"required_signal_keys":  preferKeys,
			"forbidden_signal_keys": avoidKeys,
		})
		allPrefer = append(allPrefer, preferKeys...)
		allAvoid = append(allAvoid, avoidKeys...)
	}
	if len(cases) == 0 {
		fmt.Fprintln(os.Stderr, "feedback-dataset found no usable feedback cases")
		return 1
	}
	classifications := []jsonObject{}
	for task, count := range taskCounts {
		classification, reason := classifyStrategyResidency(task, count, taskPreferCounts[task], taskAvoidCounts[task])
		classifications = append(classifications, jsonObject{
			"task_key":              task,
			"occurrences":           count,
			"prefer_signal_count":   taskPreferCounts[task],
			"avoid_signal_count":    taskAvoidCounts[task],
			"classification":        classification,
			"classification_reason": reason,
		})
	}
	sort.Slice(classifications, func(i, j int) bool {
		left := fmt.Sprint(classifications[i]["task_key"])
		right := fmt.Sprint(classifications[j]["task_key"])
		return left < right
	})
	report := jsonObject{
		"log_key": privacyHash(absClean(logPath)),
		"summary": jsonObject{
			"cases":                   len(cases),
			"prefer_signal_key_count": len(uniqueStrings(allPrefer)),
			"avoid_signal_key_count":  len(uniqueStrings(allAvoid)),
			"privacy_mode":            "hash-only-no-user-text",
		},
		"cases":            cases,
		"classifications":  classifications,
		"evolution_report": evolutionDistillReport(classifications),
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
	imageTaskMarkers := []string{"生图", "出图", "图片", "图像", "插图", "海报", "封面", "image", "illustration", "poster", "cover"}
	imageProbeMarkers := []string{
		"本地的出图入口",
		"当前环境里可用的出图入口",
		"出图技能说明",
		"skill.md",
		"可调用的生成能力",
		"别的内置入口",
		"api key",
		"openai_api_key",
		"node ok",
		"pip install openai",
		"试通道",
		"查环境",
		"check the local image entrypoint",
		"available generation capabilities",
	}
	candidateNarrative := strings.Join([]string{
		objectString(candidate, "name"),
		objectString(candidate, "objective"),
		promptTemplate,
	}, "\n")
	if len(markerHits(candidateNarrative, imageTaskMarkers)) > 0 {
		if hits := markerHits(candidateNarrative, imageProbeMarkers); len(hits) > 0 {
			failures = append(failures, "image_task_contains_preflight_probe="+strings.Join(hits, "|"))
		}
	}
	if hits := markerHits(candidateNarrative, closeoutLeakMarkers); len(hits) > 0 {
		failures = append(failures, "candidate_reopens_closeout="+strings.Join(hits, "|"))
	}
	ceremonyHits := markerHits(candidateNarrative, managementCeremonyMarkers)
	pauseHits := markerHits(candidateNarrative, managementPauseMarkers)
	if len(ceremonyHits) > 0 && len(pauseHits) > 0 {
		failures = append(failures, "candidate_contains_management_pause_loop="+strings.Join(append(ceremonyHits, pauseHits...), "|"))
	}
	roleHits := []string{}
	for _, role := range builtinTopLevelRoles {
		if strings.Contains(candidateNarrative, role) {
			roleHits = append(roleHits, role)
		}
	}
	if len(roleHits) >= 4 && (strings.Contains(candidateNarrative, "负责") || strings.Contains(strings.ToLower(candidateNarrative), "handoff")) {
		warnings = append(warnings, "candidate_role_theater_bloat="+strings.Join(roleHits, "|"))
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
		"evidence_level":   evidenceLevelFromDecision(decision),
	}
	if hasReport {
		if err := writeJSON(reportPath, report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return printGate("prompt-distill", failures)
}

func routeOwnerProfile(routeID string) string {
	switch strings.ToLower(routeID) {
	case "search":
		return "intelligence-profile"
	case "code":
		return "development-profile"
	case "execution-base":
		return "execution-base-profile"
	case "content":
		return "content-profile"
	case "visual", "imagegen", "comfyui", "video":
		return "visual-profile"
	case "spreadsheet":
		return "data-profile"
	case "prompt":
		return "content-profile"
	case "qa", "quality-inspection":
		return "staff-runtime"
	case "evolve":
		return "evolution-profile"
	default:
		return "staff-runtime"
	}
}

func routeComplexitySignalCount(routeID string, query string) int {
	signals := 0
	routeID = strings.ToLower(routeID)
	if routeID == "execution-base" || isQualityInspectionRoute(routeID) || routeID == "evolve" {
		signals++
	}
	if analysisCompletenessRequired(query) {
		signals += 2
	}
	markers := []string{
		"multiple files", "across", "regression", "migration", "security", "release", "production", "root cause", "patch debt", "parallel", "fanout",
		"多文件", "跨", "回归", "迁移", "安全", "发布", "根因", "补丁债", "并发",
	}
	signals += len(markerHits(query, markers))
	return signals
}

func routeTaskState(routeID string, tierSignalCount int, complexitySignals int) string {
	routeID = strings.ToLower(routeID)
	if routeID == "chat" {
		return "FAST_REPLY"
	}
	if routeID == "execution-base" || isQualityInspectionRoute(routeID) || routeID == "evolve" {
		return "LEGION_TASK"
	}
	_ = tierSignalCount
	if complexitySignals >= 2 {
		return "LEGION_TASK"
	}
	return "SINGLE_COMMANDER"
}

func releaseBudgetRequired(query string) bool {
	markers := []string{
		"full legion", "full scan", "whole system", "release", "ship", "publish", "all officers", "all independent officers",
		"completion claim", "final completion", "broad cleanup", "purification", "deep cleanup",
	}
	return len(markerHits(query, markers)) > 0
}

func runtimeContextAuditRequired(query string) bool {
	markers := []string{
		"token", "tokens", "cost", "cache", "hit rate", "backend usage", "runtime usage", "outer-context", "context cache",
	}
	return len(markerHits(query, markers)) > 0
}

func routeExecutionBudget(routeID string, taskState string, complexitySignals int, oversightChain []string, query string) jsonObject {
	budgetID := "LIGHT_TASK"
	verification := "targeted"
	fullAudit := any(false)
	fullSuiteMaxRuns := 0
	sidecarMode := "off"
	reason := "small_scoped_task"
	routeID = strings.ToLower(routeID)
	if taskState == "FAST_REPLY" {
		budgetID = "FAST_REPLY"
		verification = "none-unless-needed"
		reason = "discussion_or_direct_answer"
	} else if releaseBudgetRequired(query) {
		budgetID = "RELEASE_TASK"
		verification = "full-final-once"
		fullAudit = true
		fullSuiteMaxRuns = 1
		sidecarMode = "all-relevant-once"
		reason = "explicit_broad_release_or_full_scan"
	} else if taskState == "LEGION_TASK" || complexitySignals >= 2 ||
		routeID == "execution-base" || routeID == "evolve" || isQualityInspectionRoute(routeID) ||
		containsExactString(oversightChain, "root-cause-officer") ||
		containsExactString(oversightChain, "performance-benchmark-on-demand") {
		budgetID = "STRUCTURAL_TASK"
		verification = "targeted-then-final-once-if-required"
		fullAudit = "final-only-when-surface-requires"
		fullSuiteMaxRuns = 1
		sidecarMode = "triggered-only"
		reason = "structural_or_high_risk_surface"
	}
	return jsonObject{
		"id":                           budgetID,
		"reason":                       reason,
		"verification_tier":            verification,
		"full_audit":                   fullAudit,
		"full_suite_max_runs":          fullSuiteMaxRuns,
		"sidecar_mode":                 sidecarMode,
		"officer_default":              "perspective-only-unless-triggered",
		"runtime_context_audit_policy": "only-for-token-cost-cache-backend-usage-claims",
		"scope_policy":                 "current-scope-only-no-low-value-expansion",
	}
}

func canonicalOversightSeat(name string) string {
	switch strings.TrimSpace(strings.ToLower(name)) {
	case "qa":
		return "quality-inspection"
	default:
		return strings.TrimSpace(name)
	}
}

func routeOversightChain(routeID string, query string) []string {
	routeID = strings.ToLower(routeID)
	lowerQuery := strings.ToLower(query)
	chain := []string{}
	add := func(name string) {
		name = canonicalOversightSeat(name)
		if name == "" {
			return
		}
		for _, existing := range chain {
			if existing == name {
				return
			}
		}
		chain = append(chain, name)
	}
	if routeID == "execution-base" || isQualityInspectionRoute(routeID) || routeID == "evolve" {
		add("white-hat")
		add("audit")
	}
	if analysisCompletenessRequired(query) {
		add("white-hat")
		add("audit")
	}
	rootCauseMarkers := []string{
		"root cause", "root-cause", "fault localization", "failure", "failing", "rca", "symptom", "reproduce", "repro",
		"low efficiency", "inefficient", "rework", "slow fix", "slow repair", "repeat fix", "repeated fix",
		"patch debt", "workaround", "temporary patch", "hotfix", "low efficiency", "rework",
		"根因", "定位", "排查", "错误", "故障", "失败", "复现", "低效", "返工", "补丁", "临时", "绕过",
	}
	if len(markerHits(query, rootCauseMarkers)) > 0 {
		add("root-cause-officer")
	}
	if routeID == "search" || strings.Contains(lowerQuery, "github") || strings.Contains(lowerQuery, "download") ||
		strings.Contains(lowerQuery, "plugin") || strings.Contains(lowerQuery, "插件") ||
		strings.Contains(lowerQuery, "mcp") || strings.Contains(lowerQuery, "联网") ||
		strings.Contains(lowerQuery, "install") || strings.Contains(lowerQuery, "仓库") {
		add("guard-office")
	}
	if len(markerHits(query, performanceRouteMarkers())) > 0 {
		add("performance-benchmark-on-demand")
	}
	complianceMarkers := []string{
		"license", "licence", "attribution", "privacy", "pii", "personal data", "copyright", "publish", "release",
		"spdx", "sbom", "cyclonedx", "slsa", "provenance", "source compliance",
	}
	if len(markerHits(query, complianceMarkers)) > 0 {
		add("compliance-on-demand")
	}
	qualityMarkers := []string{
		"quality-inspection", "quality review", "acceptance review", "final acceptance", "final verification", "release acceptance",
		"real run", "real-run", "verified", "verification", "verify", "validated", "browser check", "program check", "command check",
		"验收", "质检", "最终验证", "真实运行", "实测", "通过",
	}
	if isQualityInspectionRoute(routeID) || len(markerHits(query, qualityMarkers)) > 0 {
		add("quality-inspection")
	}
	return chain
}

func routeSourceLineageAtoms(routeID string) []string {
	switch strings.ToLower(routeID) {
	case "search":
		return []string{"skill-discovery-filter", "session-routine", "memory-routine", "mcp-gateway-status-awareness"}
	case "code":
		return []string{"capability-gap-detection", "model-provider-routing-hint", "session-routine", "memory-routine"}
	case "execution-base":
		return []string{"context-minimization", "prefix-cache-discipline", "tool-output-compression"}
	case "qa", "quality-inspection":
		return []string{"tool-output-compression", "session-routine", "memory-routine"}
	case "evolve":
		return []string{"skill-discovery-filter", "capability-gap-detection", "memory-routine"}
	case "content":
		return []string{"context-minimization"}
	case "visual", "imagegen", "comfyui", "video":
		return []string{"context-minimization"}
	case "spreadsheet":
		return []string{"context-minimization"}
	default:
		return []string{"context-minimization"}
	}
}

func routeDistilledAtoms(routeID string, oversightChain []string) []string {
	routeID = strings.ToLower(routeID)
	atoms := []string{}
	add := func(values ...string) {
		atoms = append(atoms, values...)
	}
	switch routeID {
	case "search":
		add("guarded-realtime-source-search", "research-evidence-pack", "claim-fact-check")
	case "code":
		add("version-doc-mcp", "disciplined-debug-loop", "assumption-ledger")
	case "execution-base":
		add("reversible-evidence-handle", "content-type-compression-router")
	case "qa", "quality-inspection":
		add("assumption-ledger", "claim-fact-check", "reversible-evidence-handle")
	case "evolve":
		add("skill-stocktake-daily-library", "verified-learning-loop")
	case "content", "prompt":
		add("research-evidence-pack", "assumption-ledger")
	case "visual", "imagegen", "comfyui", "video", "spreadsheet":
		add("content-type-compression-router", "reversible-evidence-handle")
	}
	switch routeID {
	case "visual":
		add("html-native-design-canvas", "brand-asset-protocol", "anti-ai-slop-visual-rules", "design-direction-triad", "html-deck-to-editable-pptx")
	case "imagegen":
		add("brand-asset-protocol", "anti-ai-slop-visual-rules")
	case "video":
		add("motion-stage-sprite-engine", "anti-ai-slop-visual-rules")
	}
	for _, seat := range oversightChain {
		switch seat {
		case "white-hat":
			add("assumption-ledger", "claim-fact-check")
		case "guard-office":
			add("guarded-realtime-source-search", "version-doc-mcp")
		case "root-cause-officer":
			add("root-cause-radar")
		case "audit":
			add("research-evidence-pack", "reversible-evidence-handle")
		case "quality-inspection":
			add("disciplined-debug-loop", "claim-fact-check")
		case "performance-benchmark-on-demand":
			add("content-type-compression-router", "terminal-real-run-verification", "reversible-evidence-handle")
		case "compliance-on-demand":
			add("claim-fact-check", "research-evidence-pack", "guarded-realtime-source-search")
		case "qa", "质检":
			add("disciplined-debug-loop", "claim-fact-check")
		}
	}
	return uniqueStrings(atoms)
}

func mixedDistilledSourceLineageAtomFailures() []string {
	failures := []string{}
	distilled := distilledAtomKnownMap()
	for _, routeID := range []string{"search", "code", "execution-base", "quality-inspection", "evolve", "content", "prompt", "visual", "imagegen", "comfyui", "video", "spreadsheet"} {
		for _, atom := range routeSourceLineageAtoms(routeID) {
			if distilled[atom] {
				failures = append(failures, "source_lineage_atoms_contains_distilled_atom="+routeID+":"+atom)
			}
		}
	}
	return failures
}

func routeDistilledAtomCoverageFailures() []string {
	failures := []string{}
	known := distilledAtomKnownMap()
	requiredByRoute := map[string][]string{
		"search":             {"guarded-realtime-source-search", "research-evidence-pack"},
		"code":               {"version-doc-mcp", "disciplined-debug-loop"},
		"execution-base":     {"reversible-evidence-handle", "content-type-compression-router"},
		"quality-inspection": {"assumption-ledger", "claim-fact-check"},
		"evolve":             {"skill-stocktake-daily-library", "verified-learning-loop"},
		"content":            {"research-evidence-pack", "assumption-ledger"},
		"prompt":             {"research-evidence-pack", "assumption-ledger"},
		"visual":             {"html-native-design-canvas", "anti-ai-slop-visual-rules", "html-deck-to-editable-pptx"},
		"imagegen":           {"brand-asset-protocol", "anti-ai-slop-visual-rules"},
		"comfyui":            {"content-type-compression-router", "reversible-evidence-handle"},
		"video":              {"motion-stage-sprite-engine", "anti-ai-slop-visual-rules"},
		"spreadsheet":        {"content-type-compression-router", "reversible-evidence-handle"},
	}
	for routeID, required := range requiredByRoute {
		atoms := routeDistilledAtoms(routeID, nil)
		maxAtoms := 4
		if routeID == "visual" {
			maxAtoms = 7
		}
		if len(atoms) > maxAtoms {
			failures = append(failures, "route_distilled_atoms_too_many="+routeID+":"+strconv.Itoa(len(atoms)))
		}
		present := map[string]bool{}
		for _, atom := range atoms {
			if !known[atom] {
				failures = append(failures, "route_distilled_atom_unknown="+routeID+":"+atom)
			}
			present[atom] = true
		}
		for _, atom := range required {
			if !present[atom] {
				failures = append(failures, "route_distilled_atom_missing="+routeID+":"+atom)
			}
		}
	}
	return failures
}

func distilledAtomRegistryFailures() []string {
	failures := []string{}
	seen := map[string]bool{}
	for _, atom := range distilledAtomRegistry {
		if strings.TrimSpace(atom.Name) == "" {
			failures = append(failures, "distilled_atom_registry_empty_name")
			continue
		}
		if seen[atom.Name] {
			failures = append(failures, "distilled_atom_registry_duplicate="+atom.Name)
		}
		seen[atom.Name] = true
		if atom.Residency != "resident-light" && atom.Residency != "on-demand" {
			failures = append(failures, "distilled_atom_registry_bad_residency="+atom.Name+":"+atom.Residency)
		}
		if strings.TrimSpace(atom.Owner) == "" {
			failures = append(failures, "distilled_atom_registry_missing_owner="+atom.Name)
		}
	}
	if len(seen) != expectedDistilledAtomCount {
		failures = append(failures, "distilled_atom_registry_count="+strconv.Itoa(len(seen))+"!="+strconv.Itoa(expectedDistilledAtomCount))
	}
	return failures
}

func routePluginCandidates(routeID string) []string {
	switch strings.ToLower(routeID) {
	case "search":
		return []string{"GitHub", "Browser"}
	case "code":
		return []string{"GitHub", "Browser"}
	case "visual":
		return []string{"Presentations", "Canva", "Browser"}
	case "spreadsheet":
		return []string{"Spreadsheets"}
	case "content":
		return []string{"Documents"}
	case "qa", "quality-inspection":
		return []string{"Browser", "Presentations"}
	default:
		return []string{}
	}
}

func routeMCPPolicy(routeID string, oversightChain []string) string {
	switch strings.ToLower(routeID) {
	case "search", "execution-base", "qa", "quality-inspection":
		return "guard-before-mount"
	default:
		if len(oversightChain) > 0 {
			return "review-before-mount"
		}
		return "mount-only-if-gap"
	}
}

func routeDeterministicCommands(routeID string, codeMapRequired bool, oversightChain []string, query string) []string {
	commands := []string{}
	if codeMapRequired {
		commands = append(commands, "code-map")
	}
	if containsExactString(oversightChain, "root-cause-officer") {
		commands = append(commands, "root-cause-radar")
	}
	if containsExactString(oversightChain, "performance-benchmark-on-demand") {
		commands = append(commands, "bench", "bench-report", "context-bloat-audit")
		if runtimeContextAuditRequired(query) {
			commands = append(commands, "runtime-context-audit")
		}
	}
	switch strings.ToLower(routeID) {
	case "execution-base":
		commands = append(commands, "route-task")
	case "qa", "quality-inspection":
		commands = append(commands, "quality-guard", "closeout-check")
	case "search":
		if len(oversightChain) > 0 {
			commands = append(commands, "mcp-guard")
		}
	case "visual":
		commands = append(commands, "preview")
	}
	return uniqueStrings(commands)
}

func routeQueryDistilledAtoms(routeID string, oversightChain []string, query string) []string {
	atoms := routeDistilledAtoms(routeID, oversightChain)
	if analysisCompletenessRequired(query) {
		atoms = append(atoms, "assumption-ledger", "claim-fact-check", "research-evidence-pack", "reversible-evidence-handle")
	}
	priorArtMarkers := []string{
		"existing solution", "prior art", "from scratch", "open source", "open-source", "research", "search", "github issue", "known solution", "tooling",
		"方案", "解决", "修复", "根因", "问题", "全网", "搜索", "借鉴", "现成", "开源", "工具", "不要从0", "不要从零",
	}
	priorArtMarkers = append(priorArtMarkers, "方案", "解决", "修复", "根因", "问题", "全网", "搜索", "借鉴", "现成", "开源", "工具", "不要从零")
	if len(markerHits(query, priorArtMarkers)) > 0 {
		atoms = append(atoms, "prior-art-solution-search")
	}
	rootCauseMarkers := []string{
		"root cause", "root-cause", "fault localization", "failure", "failing", "rca", "symptom", "reproduce", "repro",
		"low efficiency", "inefficient", "rework", "slow fix", "slow repair", "repeat fix", "repeated fix",
		"根因", "定位", "排查", "错误", "故障", "失败", "复现", "低效", "返工",
	}
	if len(markerHits(query, rootCauseMarkers)) > 0 {
		atoms = append(atoms, "root-cause-radar")
	}
	parallelHypothesisMarkers := []string{
		"parallel", "fanout", "multiple candidates", "candidate", "hypothesis", "hypotheses", "nine places", "several places", "simultaneously", "concurrent",
		"并发", "同时", "候选", "假设", "多个地方", "多个位置", "九个", "多开", "平行",
	}
	if len(markerHits(query, parallelHypothesisMarkers)) > 0 {
		atoms = append(atoms, "parallel-hypothesis-fanout")
	}
	patchDebtMarkers := []string{
		"patch debt", "bloat", "bloated", "temporary patch", "hotfix", "workaround", "rule debt", "technical debt", "debt", "root cure",
		"补丁", "臃肿", "债务", "治本", "根治", "越修越胖", "隐患", "系统资源",
	}
	if len(markerHits(query, patchDebtMarkers)) > 0 {
		atoms = append(atoms, "root-cause-radar", "patch-debt-root-cure")
	}
	terminalVerificationMarkers := []string{
		"complete", "completed", "done", "final", "real run", "real-run", "verification", "verify", "validated", "browser", "test", "command", "terminal", "local run",
		"完成", "实测", "真实", "验证", "电脑", "权威", "浏览器", "命令", "本地运行", "通过", "别停", "继续到底",
	}
	if len(markerHits(query, terminalVerificationMarkers)) > 0 {
		atoms = append(atoms, "terminal-real-run-verification")
	}
	if strings.EqualFold(routeID, "visual") {
		motionMarkers := []string{"motion", "animated", "animation", "stage", "sprite", "timeline", "product demo", "demo video", "mp4", "gif", "动效", "动画", "视频"}
		if len(markerHits(query, motionMarkers)) > 0 {
			atoms = append(atoms, "motion-stage-sprite-engine")
		}
	}
	return uniqueStrings(atoms)
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
	rules := mergedRoutingRules(config)
	modelProfiles := mergedModelProfiles(config)
	defaultTier := resolvedDefaultModelTier(config)
	lowerQuery := strings.ToLower(query)
	bestScore := -1
	bestPriority := -1
	bestRule := jsonObject{}
	matches := []string{}
	for _, rule := range rules {
		ruleMatches := routeKeywordMatches(lowerQuery, rule.Keywords)
		score := len(ruleMatches)
		priority := rule.Priority
		if score > bestScore || (score == bestScore && priority > bestPriority) {
			bestScore = score
			bestPriority = priority
			bestRule = jsonObject{
				"id":          rule.ID,
				"name":        rule.Name,
				"provider_id": rule.ProviderID,
				"model":       rule.Model,
				"priority":    priority,
			}
			matches = ruleMatches
		}
	}
	if bestScore < 0 {
		fmt.Fprintln(os.Stderr, "no routing rule found")
		return 1
	}
	if bestScore == 0 {
		for _, rule := range rules {
			if strings.EqualFold(rule.ID, "chat") {
				bestRule = jsonObject{
					"id":          rule.ID,
					"name":        rule.Name,
					"provider_id": rule.ProviderID,
					"model":       rule.Model,
					"priority":    rule.Priority,
				}
				bestPriority = rule.Priority
				matches = []string{}
				break
			}
		}
	}
	bestRouteID := strings.ToLower(fmt.Sprint(bestRule["id"]))
	comfyTechnicalMarkers := []string{"comfyui", "工作流", "workflow", "节点", "node", "插件", "plugin", "批量生成", "batch generation", "视频管线", "video pipeline", "animatediff", "controlnet", "技术美术", "主界面截图", "comfyui截图"}
	if bestRouteID == "comfyui" && len(markerHits(query, comfyTechnicalMarkers)) == 0 {
		for _, rule := range rules {
			if !strings.EqualFold(rule.ID, "imagegen") {
				continue
			}
			imageMatches := []string{}
			imageMatches = routeKeywordMatches(lowerQuery, rule.Keywords)
			bestRule = jsonObject{
				"id":          rule.ID,
				"name":        rule.Name,
				"provider_id": rule.ProviderID,
				"model":       rule.Model,
				"priority":    rule.Priority,
			}
			bestPriority = rule.Priority
			matches = imageMatches
			bestRouteID = "imagegen"
			break
		}
	}
	if bestRouteID == "video" {
		visualDeliverableMarkers := []string{
			"ppt", "pptx", "presentation", "slide", "deck", "html deck", "editable pptx",
			"ui", "interface", "web page", "landing page", "browser deck",
			"PPT", "PPTX", "演示文稿", "幻灯片", "页面", "界面", "可编辑",
		}
		if len(markerHits(query, visualDeliverableMarkers)) > 0 {
			for _, rule := range rules {
				if !strings.EqualFold(rule.ID, "visual") {
					continue
				}
				visualMatches := []string{}
				visualMatches = routeKeywordMatches(lowerQuery, rule.Keywords)
				bestRule = jsonObject{
					"id":          rule.ID,
					"name":        rule.Name,
					"provider_id": rule.ProviderID,
					"model":       rule.Model,
					"priority":    rule.Priority,
				}
				bestPriority = rule.Priority
				matches = visualMatches
				bestRouteID = "visual"
				break
			}
		}
	}
	tierSignalCount := routeTierSignalCount(bestRouteID, bestScore, query)
	complexitySignals := routeComplexitySignalCount(bestRouteID, query)
	complexityTier := defaultTier
	reasoningEffort := "low"
	tierReason := "default_low_cost_route"
	if bestRouteID == "imagegen" {
		complexityTier = "low"
		reasoningEffort = "low"
		tierReason = "direct_image_task"
	} else if tierSignalCount >= 6 || complexitySignals >= 3 {
		complexityTier = "high"
		reasoningEffort = "high"
		tierReason = "dense_or_high_risk_task"
	} else if tierSignalCount >= 3 || complexitySignals >= 1 {
		complexityTier = "standard"
		reasoningEffort = "medium"
		tierReason = "multi_signal_or_risk_task"
	} else {
		complexityTier = "low"
		reasoningEffort = "low"
		tierReason = "simple_or_low_risk_task"
	}
	selectedProfile := jsonObject{}
	if profile, ok := modelProfiles[complexityTier]; ok {
		selectedProfile = jsonObject{
			"provider_id":      profile.ProviderID,
			"model":            profile.Model,
			"reasoning_effort": profile.ReasoningEffort,
		}
		if profile.ReasoningEffort != "" {
			reasoningEffort = profile.ReasoningEffort
		}
	}
	ownerProfile := routeOwnerProfile(bestRouteID)
	taskState := routeTaskState(bestRouteID, tierSignalCount, complexitySignals)
	oversightChain := routeOversightChain(bestRouteID, query)
	executionBudget := routeExecutionBudget(bestRouteID, taskState, complexitySignals, oversightChain, query)
	analysisRequired := analysisCompletenessRequired(query)
	if taskState == "FAST_REPLY" && len(oversightChain) > 0 {
		taskState = "SINGLE_COMMANDER"
		executionBudget = routeExecutionBudget(bestRouteID, taskState, complexitySignals, oversightChain, query)
	}
	report := jsonObject{
		"query_key":           privacyHash(query),
		"query_length":        len(query),
		"privacy_mode":        "hash-only-no-raw-query",
		"matched_route":       bestRule,
		"matched_count":       bestScore,
		"tier_signal_count":   tierSignalCount,
		"matched_terms":       matches,
		"complexity_signals":  complexitySignals,
		"recommended_tier":    complexityTier,
		"recommended_profile": selectedProfile,
		"reasoning_effort":    reasoningEffort,
		"tier_reason":         tierReason,
		"canon_source":        "go-builtin+config-overlay",
		"route_rule_count":    len(rules),
	}
	report["analysis_completeness_required"] = analysisRequired
	codeMapRequired := false
	if bestRouteID == "code" && (tierSignalCount >= 3 || complexitySignals >= 2) {
		codeMapRequired = true
	}
	report["code_map_required"] = codeMapRequired
	if codeMapRequired {
		report["next_required_artifact"] = "code-map"
	}
	prefixClass := "minimal"
	if taskState == "FAST_REPLY" {
		prefixClass = "small"
	} else if taskState == "LEGION_TASK" {
		prefixClass = "structured"
	}
	goGateClass := "light"
	if bestRouteID == "execution-base" {
		goGateClass = "heavy"
	} else if isQualityInspectionRoute(bestRouteID) {
		goGateClass = "verification"
	}
	report["task_route"] = jsonObject{
		"state":                 taskState,
		"owner_profile":         ownerProfile,
		"route_id":              bestRule["id"],
		"route_name":            bestRule["name"],
		"oversight_chain":       oversightChain,
		"closeout_policy":       "finish-with-verification",
		"resident_prefix_class": prefixClass,
	}
	report["execution_budget"] = executionBudget
	report["capability_mounts"] = jsonObject{
		"distilled_atoms":      routeQueryDistilledAtoms(bestRouteID, oversightChain, query),
		"source_lineage_atoms": routeSourceLineageAtoms(bestRouteID),
		"plugin_candidates":    routePluginCandidates(bestRouteID),
		"mcp_policy":           routeMCPPolicy(bestRouteID, oversightChain),
		"mount_strategy":       "minimal-gap-first",
		"resident_policy":      "minimal-stable-skeleton-only",
		"retire_policy":        "replace-weaker-atoms-instead-of-stacking",
	}
	capabilityMounts, _ := report["capability_mounts"].(jsonObject)
	if bestRouteID == "search" || containsExactString(stringSlice(map[string]any(capabilityMounts), "distilled_atoms"), "prior-art-solution-search") {
		report["intelligence_profile_contract"] = intelligenceProfileContract()
	}
	report["deterministic_execution"] = jsonObject{
		"required":           codeMapRequired || strings.EqualFold(bestRouteID, "execution-base") || isQualityInspectionRoute(bestRouteID) || containsExactString(oversightChain, "root-cause-officer") || containsExactString(oversightChain, "performance-benchmark-on-demand"),
		"command_candidates": routeDeterministicCommands(bestRouteID, codeMapRequired, oversightChain, query),
		"go_gate_class":      goGateClass,
		"tool_output_policy": "compress-before-reuse-preserve-evidence",
		"evidence_retention": "raw-handle-kept-summary-fed",
	}
	report["optimization_policy"] = jsonObject{
		"objective":               "smaller-stable-prefix-with-equal-or-better-hit-rate",
		"prefix_cache_discipline": "byte-stable-prefix-volatile-facts-late",
		"dynamic_tail_policy":     "timestamps-paths-temp-state-late",
		"measurement_loop":        "bench-when-cost-hit-rate-speed-disputed",
	}
	report["concise_execution_contract"] = conciseExecutionContract()
	report["execution_budget_contract"] = executionBudgetContract()
	if analysisRequired {
		report["analysis_completeness_contract"] = analysisCompletenessContract()
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
	taskRoute, _ := objectMap(routeReport, "task_route")
	capabilityMounts, _ := objectMap(routeReport, "capability_mounts")
	deterministicExecution, _ := objectMap(routeReport, "deterministic_execution")
	executionBudget, _ := objectMap(routeReport, "execution_budget")
	cacheConfig, _ := objectMap(config, "cache_config")
	ironRulesVersion := objectString(config, "iron_rules_version")
	if ironRulesVersion == "" {
		ironRulesVersion = builtinIronRulesVersion
	}
	stablePrefix := jsonObject{
		"iron_rules_version":       ironRulesVersion,
		"route_id":                 objectString(routeInfo, "id"),
		"task_state":               objectString(taskRoute, "state"),
		"execution_budget_id":      objectString(executionBudget, "id"),
		"owner_profile":            objectString(taskRoute, "owner_profile"),
		"target_hit_rate":          cacheConfig["target_hit_rate"],
		"flatten_threshold":        cacheConfig["flatten_threshold"],
		"stable_prefix_policy":     objectString(cacheConfig, "stable_prefix_policy"),
		"mount_policy":             objectString(cacheConfig, "mount_policy"),
		"tool_output_policy":       objectString(cacheConfig, "tool_output_policy"),
		"concise_execution_policy": objectString(cacheConfig, "concise_execution_policy"),
		"optimization_objective":   objectString(cacheConfig, "optimization_objective"),
		"canon_source":             "go-builtin+config-overlay",
	}
	artifactSummaries := []jsonObject{}
	for _, artifact := range artifacts {
		artifactSummaries = append(artifactSummaries, summarizeArtifactSafe(workspace, artifact))
	}
	executionSummaries, auditSummaries := splitArtifactSummaries(artifactSummaries)
	prefixCanon := stablePrefixCanon(stablePrefix)
	assemblyReview := reviewOptimizationAssembly(stablePrefix, artifactSummaries, executionSummaries, auditSummaries)
	dynamicContext := jsonObject{
		"query_key":           privacyHash(query),
		"query_length":        len(query),
		"artifact_count":      len(artifacts),
		"workspace_key":       privacyHash(absClean(workspace)),
		"provider_id":         objectString(routeInfo, "provider_id"),
		"model_tier":          objectString(routeReport, "recommended_tier"),
		"reasoning_effort":    objectString(routeReport, "reasoning_effort"),
		"volatile_tail_rule":  "timestamps-paths-temp-state-late",
		"distilled_atoms":     capabilityMounts["distilled_atoms"],
		"execution_summaries": executionSummaries,
		"audit_summaries":     auditSummaries,
	}
	cacheKey := objectString(prefixCanon, "canon_hash")
	routeSummary := jsonObject{
		"query_key":              objectString(routeReport, "query_key"),
		"query_length":           routeReport["query_length"],
		"route_id":               objectString(routeInfo, "id"),
		"task_state":             objectString(taskRoute, "state"),
		"execution_budget":       executionBudget,
		"owner_profile":          objectString(taskRoute, "owner_profile"),
		"oversight_chain":        taskRoute["oversight_chain"],
		"recommended_tier":       objectString(routeReport, "recommended_tier"),
		"reasoning_effort":       objectString(routeReport, "reasoning_effort"),
		"complexity_signals":     routeReport["complexity_signals"],
		"code_map_required":      routeReport["code_map_required"],
		"analysis_required":      routeReport["analysis_completeness_required"],
		"deterministic_required": deterministicExecution["required"],
		"command_candidates":     deterministicExecution["command_candidates"],
	}
	contextPack := jsonObject{
		"stable_prefix":              stablePrefix,
		"stable_prefix_canon":        prefixCanon,
		"dynamic_context":            dynamicContext,
		"route_summary":              routeSummary,
		"cache_key":                  cacheKey,
		"cache_strategy":             "stable-prefix-first",
		"concise_execution_contract": conciseExecutionContract(),
		"execution_budget_contract":  executionBudgetContract(),
		"artifact_summaries":         artifactSummaries,
		"review_gates": []jsonObject{
			assemblyReview,
		},
		"optimization_policy": jsonObject{
			"objective":           objectString(cacheConfig, "optimization_objective"),
			"evidence_retention":  objectString(cacheConfig, "evidence_retention_policy"),
			"compression_policy":  objectString(cacheConfig, "tool_output_policy"),
			"concise_execution":   objectString(cacheConfig, "concise_execution_policy"),
			"prefix_canon_policy": "ordered-fields-short-canon-no-duplicate-phrasing",
		},
	}
	if objectBoolValue(routeReport, "analysis_completeness_required") {
		contextPack["analysis_completeness_contract"] = analysisCompletenessContract()
	}
	outputPath := reportPath
	if !hasReport {
		outputPath = filepath.Join(workspace, "outputs", "context-pack-rich.json")
	}
	if err := writeCompactJSON(outputPath, contextPack); err != nil {
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
	if os.Args[1] == "help" {
		if len(os.Args) > 2 && os.Args[2] == "all" {
			usageAll()
		} else {
			usage()
		}
		os.Exit(0)
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
	case "code-map":
		code = codeMapCommand(args)
	case "root-cause-radar":
		code = rootCauseRadarCommand(args)
	case "bugfix-guard":
		code = bugfixGuardCommand(args)
	case "qa-guard", "quality-guard":
		code = qualityGuardCommand(args)
	case "migration-guard":
		code = migrationGuardCommand(args)
	case "closeout-check":
		code = closeoutCheckCommand(args)
	case "finish-or-block":
		code = finishOrBlockCommand(args)
	case "repeat-candidates":
		code = repeatCandidatesCommand(args)
	case "evidence-grade":
		code = evidenceGradeCommand(args)
	case "truth-state":
		code = truthStateCommand(args)
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
	case "ppt-template-inspect":
		code = pptTemplateInspectCommand(args)
	case "ppt-template-starter":
		code = pptTemplateStarterCommand(args)
	case "ppt-template-edit":
		code = pptTemplateEditCommand(args)
	case "ppt-template-fidelity":
		code = pptTemplateFidelityCommand(args)
	case "ppt-htmlfirst":
		code = pptHTMLFirstCommand(args)
	case "ppt-com-refine":
		code = pptCOMRefineCommand(args)
	case "ppt-pipeline":
		code = pptPipelineCommand(args)
	case "mcp-guard":
		code = mcpGuard(args)
	case "supply-chain":
		code = supplyChainCommand(args)
	case "mcp-distill":
		code = mcpDistill(args)
	case "canon-report":
		code = canonReportCommand(args)
	case "fusion-audit":
		code = fusionAuditCommand(args)
	case "optimization-audit":
		code = optimizationAuditCommand(args)
	case "context-bloat-audit":
		code = contextBloatAuditCommand(args)
	case "runtime-context-audit":
		code = runtimeContextAuditCommand(args)
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
