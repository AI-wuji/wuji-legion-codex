package main

import (
	"archive/zip"
	"bytes"
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

var slideTextPattern = regexp.MustCompile(`(?is)<a:t[^>]*>(.*?)</a:t>`)

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

const builtinIronRulesVersion = "10.6"
const builtinDefaultModelTier = "low"

var builtinTopLevelRoles = []string{
	"阿极",
	"参谋本部",
	"女娲",
	"白帽",
	"质检",
	"安全",
	"合规",
}

var builtinModelProfiles = map[string]modelProfile{
	"low": {
		ProviderID:      "openai-api",
		Model:           "gpt-5.4-mini",
		ReasoningEffort: "low",
	},
	"standard": {
		ProviderID:      "openai-api",
		Model:           "gpt-5.4",
		ReasoningEffort: "medium",
	},
	"high": {
		ProviderID:      "openai-api",
		Model:           "gpt-5.5",
		ReasoningEffort: "high",
	},
}

var builtinRoutingRules = []routeRule{
	{
		ID:         "search",
		Name:       "搜索/信息检索",
		Keywords:   []string{"搜索", "查", "调研", "研究", "搜一下", "find", "search"},
		ProviderID: "deepseek-web",
		Priority:   100,
	},
	{
		ID:         "code",
		Name:       "代码生成/开发",
		Keywords:   []string{"写", "开发", "实现", "函数", "生成", "创建", "bug", "修复", "重构", "Rust", "Tauri", "小程序", "ComfyUI插件", "插件", "Python", "PowerShell", "自动化", "AI工程", "RAG", "编程", "compile"},
		ProviderID: "deepseek-web",
		Priority:   80,
	},
	{
		ID:         "execution-base",
		Name:       "执行底座",
		Keywords:   []string{"执行底座", "执行底座主帅", "wuji-cli", "Go", "执行引擎", "guard", "task", "sync", "audit", "workflow", "beep", "bench", "preview调度", "pptx-preflight", "pptx-batch-gate", "pptx-audit", "asset-map", "time-guard", "mcp-guard", "MCP门禁", "插件门禁", "reference-frame-map", "reusable-asset-map", "illustration-plan", "pilot-page", "pilot-preview", "pilot-score"},
		ProviderID: "deepseek-web",
		Priority:   82,
	},
	{
		ID:         "content",
		Name:       "文案/内容创作",
		Keywords:   []string{"文章", "文案", "脚本", "剧本", "分镜", "小说", "短篇", "长篇", "爽文", "教程", "教案", "课程", "计划书", "营销方案", "卖点", "博客", "故事", "内容", "文档", "报告", "Word", "docx", "markdown"},
		ProviderID: "deepseek-web",
		Priority:   70,
	},
	{
		ID:         "visual",
		Name:       "视觉/PPT/UI",
		Keywords:   []string{"PPT", "presentation", "演示文稿", "幻灯片", "slide", "deck", "OpenDesign", "Remotion", "设计系统", "动态演示", "设计", "design", "美化", "画", "UI", "界面", "页面", "落地页", "官网", "前端", "html", "css", "预览图", "页面图"},
		ProviderID: "deepseek-web",
		Priority:   60,
	},
	{
		ID:         "video",
		Name:       "视频/短视频/短剧",
		Keywords:   []string{"视频", "短视频", "短剧", "短片", "动画", "reel"},
		ProviderID: "deepseek-web",
		Priority:   55,
	},
	{
		ID:         "imagegen",
		Name:       "imagegen/direct-image",
		Keywords:   []string{"生图", "出图", "生成图片", "图片生成", "图像生成", "插图", "教学插图", "海报", "封面", "配图", "画一张图", "做一张图", "image", "illustration", "poster", "cover", "generate image"},
		ProviderID: "imagegen",
		Priority:   52,
	},
	{
		ID:         "prompt",
		Name:       "提示词工程",
		Keywords:   []string{"prompt", "提示词", "提示词扩写", "扩写", "分镜", "故事板", "storyboard", "image-spec", "图片提示词"},
		ProviderID: "deepseek-web",
		Priority:   50,
	},
	{
		ID:         "spreadsheet",
		Name:       "表格/结构化数据",
		Keywords:   []string{"表格", "Excel", "xlsx", "spreadsheet", "数据表", "清单", "台账"},
		ProviderID: "deepseek-web",
		Priority:   58,
	},
	{
		ID:         "comfyui",
		Name:       "ComfyUI/图像生成",
		Keywords:   []string{"ComfyUI", "工作流", "生图", "生成图片", "图片生成", "图像生成", "截图", "主界面", "App截图", "海报", "封面", "图像", "图片", "渲染"},
		ProviderID: "deepseek-web",
		Priority:   95,
	},
	{
		ID:         "qa",
		Name:       "审计/红队",
		Keywords:   []string{"审计", "审核", "检查", "验收", "审查", "质量", "反对", "批评", "白帽", "质检", "合规", "许可证", "安全", "性能", "token", "bug一堆", "不好用", "丑", "崩", "报错"},
		ProviderID: "deepseek-web",
		Priority:   40,
	},
	{
		ID:         "evolve",
		Name:       "进化/复盘",
		Keywords:   []string{"复盘", "进化", "优化", "自动", "学习", "改进", "分析失败", "蒸馏", "融合skill", "融合 skill", "升级skill", "升级 skill", "官方源", "源码核验", "能力融合", "工作流工件", "可审计轨迹", "workflow", "packet"},
		ProviderID: "deepseek-web",
		Priority:   30,
	},
	{
		ID:         "chat",
		Name:       "日常对话",
		Keywords:   []string{},
		ProviderID: "deepseek-web",
		Priority:   0,
	},
}

var builtinPluginBindings = []pluginBinding{
	{
		Plugin:  "Browser",
		Owners:  []string{"视觉主帅", "开发主帅"},
		Purpose: "网页打开、检查、交互测试、截图",
	},
	{
		Plugin:  "Documents",
		Owners:  []string{"内容主帅"},
		Purpose: "Word/文档生成、整理、归档",
	},
	{
		Plugin:  "Spreadsheets",
		Owners:  []string{"情报主帅", "内容主帅"},
		Purpose: "表格、结构化数据、分析交付",
	},
	{
		Plugin:  "Presentations",
		Owners:  []string{"视觉主帅"},
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
	fmt.Fprintln(os.Stderr, "  wuji-cli workflow-guard --workspace <dir> [--stage scaffold|final]")
	fmt.Fprintln(os.Stderr, "  wuji-cli claim-guard --claim <text> [--evidence <file>]...")
	fmt.Fprintln(os.Stderr, "  wuji-cli time-guard --kind <non-code|general> --elapsed-minutes <n> [--artifact <file>] [--phase <name>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli task --workspace <dir> --event <start|heartbeat|blocked|end> [--status <running|blocked|needs_decision|done>] [--artifact <file>]... [--note <text>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli sync --source <dir> --dest <dir>")
	fmt.Fprintln(os.Stderr, "  wuji-cli audit --path <dir> [--report <file>] [--sarif <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli bench --workspace <dir> --name <run> [--input-tokens <n>] [--output-tokens <n>] [--duration-ms <n>] [--tool-calls <n>] [--retries <n>] [--qa-pass <true|false>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli bench-report --workspace <dir> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli code-map --workspace <dir> --goal <text> --entry <text> [--dependency <text>]... [--risk <text>]... [--verify <text>]... [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli closeout-check --workspace <dir> --goal <text> [--artifact <file>]... [--verify <text>]... [--next-gap <text>]... [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli preview --command <exe> [--arg <value>]... --output <file>")
	fmt.Fprintln(os.Stderr, "  wuji-cli asset-map --pptx <file> --workspace <dir>")
	fmt.Fprintln(os.Stderr, "  wuji-cli pptx-audit --pptx <file> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli pptx-preflight --workspace <dir> [--generator <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli pptx-batch-gate --workspace <dir> [--generator <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli ppt-template-inspect --workspace <dir> --pptx <file> [--out-dir <dir>] [--scale <n>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli ppt-template-starter --workspace <dir> --pptx <file> --map <file> --out <file> [--preview-dir <dir>] [--layout-dir <dir>] [--inspect <file>] [--contact-sheet <file>] [--scale <n>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli ppt-template-edit --workspace <dir> --starter-pptx <file> --map <file> --out <file> [--preview-dir <dir>] [--layout-dir <dir>] [--report <file>] [--scale <n>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli ppt-template-fidelity --workspace <dir> --final-pptx <file> [--map <file>] [--starter-pptx <file>] [--starter-layout-dir <dir>] [--final-layout-dir <dir>] [--edit-dir <dir>] [--agent-log <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli ppt-htmlfirst --workspace <dir> --html <file> --out <file> [--title <text>] [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli ppt-com-refine --pptx <file> --out <file> [--instructions <file>] [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli ppt-pipeline --workspace <dir> --route <html-first|template-following> --out <file> [--html <file>] [--pptx <file>] [--map <file>] [--report <file>] [--auto-approve true|false] [--pilot-approval <file>] [--com-refine true|false] [--refine-instructions <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli mcp-guard --manifest <file> [--workspace <dir>] [--allow-network true|false]")
	fmt.Fprintln(os.Stderr, "  wuji-cli mcp-distill --catalog <file> [--report <file>]")
	fmt.Fprintln(os.Stderr, "  wuji-cli canon-report [--report <file>]")
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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
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
		"goal":         strings.TrimSpace(goal),
		"entry":        strings.TrimSpace(entry),
		"dependencies": dependencies,
		"risks":        risks,
		"verifications": verifications,
		"generated_at": time.Now().UTC().Format(time.RFC3339),
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
	reportPath, hasReport := argValue(args, "--report")
	failures := []string{}
	if len(artifacts) == 0 {
		failures = append(failures, "closeout_requires_artifact")
	}
	for _, artifact := range artifacts {
		if !nonEmpty(artifact) {
			failures = append(failures, "artifact_missing_or_too_small="+artifact)
		}
	}
	if len(verifications) == 0 {
		failures = append(failures, "closeout_requires_verification")
	}
	for _, gap := range nextGaps {
		if strings.TrimSpace(gap) != "" {
			failures = append(failures, "closeout_gap_remaining="+gap)
		}
	}
	report := jsonObject{
		"workspace":      absClean(workspace),
		"goal":           strings.TrimSpace(goal),
		"artifacts":      artifacts,
		"verifications":  verifications,
		"remaining_gaps": nextGaps,
		"status":         "pass",
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
	for _, stem := range []string{"reference-frame-map", "reusable-asset-map", "illustration-plan", "style-lock", "page-role-policy"} {
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
	for _, stem := range []string{"reference-frame-map", "reusable-asset-map", "illustration-plan", "style-lock", "page-role-policy"} {
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
	cloned := make([]routeRule, 0, len(builtinRoutingRules))
	for _, rule := range builtinRoutingRules {
		copied := rule
		copied.Keywords = append([]string{}, rule.Keywords...)
		cloned = append(cloned, copied)
	}
	return cloned
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

func builtinCanonReport() jsonObject {
	return jsonObject{
		"iron_rules_version": builtinIronRulesVersion,
		"default_model_tier": builtinDefaultModelTier,
		"top_level_roles":    append([]string{}, builtinTopLevelRoles...),
		"model_profiles":     modelProfilesAsJSON(cloneBuiltinModelProfiles()),
		"routing_rules":      routeRulesAsJSON(cloneBuiltinRoutingRules()),
		"built_in_plugins":   pluginBindingsAsJSON(),
		"mcp_default_policy": mcpPoliciesAsJSON(),
		"canon_source":       "go-builtin",
	}
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
	rules := mergedRoutingRules(config)
	modelProfiles := mergedModelProfiles(config)
	defaultTier := resolvedDefaultModelTier(config)
	lowerQuery := strings.ToLower(query)
	bestScore := -1
	bestPriority := -1
	bestRule := jsonObject{}
	matches := []string{}
	for _, rule := range rules {
		score := 0
		ruleMatches := []string{}
		for _, keyword := range rule.Keywords {
			if keyword != "" && strings.Contains(lowerQuery, strings.ToLower(keyword)) {
				score++
				ruleMatches = append(ruleMatches, keyword)
			}
		}
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
			for _, keyword := range rule.Keywords {
				if keyword != "" && strings.Contains(lowerQuery, strings.ToLower(keyword)) {
					imageMatches = append(imageMatches, keyword)
				}
			}
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
	complexityTier := defaultTier
	reasoningEffort := "low"
	tierReason := "default_low_cost_route"
	if bestRouteID == "imagegen" {
		complexityTier = "low"
		reasoningEffort = "low"
		tierReason = "direct_image_task"
	} else if bestScore >= 6 || bestPriority >= 90 {
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
	report := jsonObject{
		"query":               query,
		"matched_route":       bestRule,
		"matched_count":       bestScore,
		"matched_terms":       matches,
		"recommended_tier":    complexityTier,
		"recommended_profile": selectedProfile,
		"reasoning_effort":    reasoningEffort,
		"tier_reason":         tierReason,
		"canon_source":        "go-builtin+config-overlay",
		"route_rule_count":    len(rules),
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
	ironRulesVersion := objectString(config, "iron_rules_version")
	if ironRulesVersion == "" {
		ironRulesVersion = builtinIronRulesVersion
	}
	stablePrefix := jsonObject{
		"iron_rules_version": ironRulesVersion,
		"route_id":           objectString(routeInfo, "id"),
		"provider_id":        objectString(routeInfo, "provider_id"),
		"model_tier":         objectString(routeReport, "recommended_tier"),
		"reasoning_effort":   objectString(routeReport, "reasoning_effort"),
		"target_hit_rate":    cacheConfig["target_hit_rate"],
		"flatten_threshold":  cacheConfig["flatten_threshold"],
		"canon_source":       "go-builtin+config-overlay",
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
	case "code-map":
		code = codeMapCommand(args)
	case "closeout-check":
		code = closeoutCheckCommand(args)
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
	case "mcp-distill":
		code = mcpDistill(args)
	case "canon-report":
		code = canonReportCommand(args)
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
