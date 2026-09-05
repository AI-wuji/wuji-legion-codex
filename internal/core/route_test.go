package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryAutomaticSourcesHaveSemanticRoutes(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	items, err := LoadManifests(root)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		query      string
		capability string
		skill      string
		sources    []string
	}{
		{"审查这段代码的潜在问题", "code-review", "wuji-code-review-suite", []string{"wuji-code-review-suite-unified"}},
		{"分析这个CSV数据并给出趋势", "data", "wuji-data-suite", []string{"wuji-data-suite-unified"}},
		{"把这份Word文档转为PDF并检查版式", "documents", "wuji-document-suite", []string{"wuji-document-suite-unified"}},
		{"构建一个React仪表盘页面", "frontend", "wuji-frontend-suite", []string{"wuji-frontend-suite-unified"}},
		{"生成一张产品宣传图", "image", "wuji-image-suite", []string{"wuji-image-suite-unified"}},
		{"做一份可编辑的PPT", "presentation", "wuji-editable-deck", []string{"wuji-editable-deck-unified"}},
		{"创建一个浏览器可编辑的HTML演示", "presentation", "wuji-web-deck", []string{"wuji-web-deck-unified", "wuji-dashiai-deck-adapter"}},
		{"搜索GitHub和官方文档找到最新解决方案", "search", "wuji-research-suite", []string{"wuji-research-suite-unified"}},
		{"扫描仓库中的安全漏洞和密钥泄露", "security", "wuji-security-suite", []string{"wuji-security-suite-unified"}},
		{"制作一个产品演示视频", "video", "wuji-video-suite", []string{"wuji-video-suite-unified"}},
		{"优化这个界面的视觉设计", "visual", "wuji-visual-suite", []string{"wuji-visual-suite-unified"}},
		{"把这篇文章翻译并润色成中文", "writing", "wuji-writing-suite", []string{"wuji-writing-suite-unified"}},
	}
	for _, test := range cases {
		t.Run(test.capability+"/"+test.skill, func(t *testing.T) {
			got := Route(test.query, items)
			if got.Capability != test.capability || got.PrimarySkill != test.skill {
				t.Fatalf("query %q routed to %s/%s, want %s/%s", test.query, got.Capability, got.PrimarySkill, test.capability, test.skill)
			}
			if got.SourceActivationError != "" {
				t.Fatalf("query %q reported source activation error: %s", test.query, got.SourceActivationError)
			}
			mounted := map[string]bool{}
			for _, source := range got.MountedSources {
				mounted[source.ID] = true
			}
			for _, source := range test.sources {
				if !mounted[source] {
					t.Fatalf("query %q did not mount automatic source %q: %#v", test.query, source, got.MountedSources)
				}
			}
			activated := map[string]SourceExecutionContract{}
			for _, contract := range got.SourceExecution {
				activated[contract.SourceID] = contract
			}
			for _, source := range test.sources {
				contract, ok := activated[source]
				if !ok || contract.InvocationKind != sourceEntrypointInvocationKind || contract.Entrypoint == "" || contract.EntrypointSHA256 == "" || contract.EntrypointContent == "" {
					t.Fatalf("query %q did not load automatic source %q into an execution contract: %#v", test.query, source, got.SourceExecution)
				}
			}
			for _, worker := range got.Workers {
				if len(worker.SourceExecution) != len(got.SourceExecution) {
					t.Fatalf("worker %q did not receive route source contracts: %#v", worker.ID, worker.SourceExecution)
				}
			}
		})
	}
}

func TestOfficeCLIAdapterOnlyMountsForNarrowDocumentInspection(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	items, err := LoadManifests(root)
	if err != nil {
		t.Fatal(err)
	}
	inspection := Route("导出Excel结构为 JSON", items)
	if inspection.Capability != "documents" || inspection.PrimarySkill != "wuji-document-suite" {
		t.Fatalf("OfficeCLI inspection displaced the documents route: %#v", inspection)
	}
	if !containsMountedSource(inspection.MountedSources, "officecli-stateless-adapter") {
		t.Fatalf("OfficeCLI inspection did not mount its narrow adapter: %#v", inspection.MountedSources)
	}
	ordinary := Route("创建一份 Word 文档", items)
	if containsMountedSource(ordinary.MountedSources, "officecli-stateless-adapter") {
		t.Fatalf("ordinary document work mounted OfficeCLI: %#v", ordinary.MountedSources)
	}
	pptx := Route("导出 PPTX 结构", items)
	if pptx.Capability != "presentation" || containsMountedSource(pptx.MountedSources, "officecli-stateless-adapter") {
		t.Fatalf("presentation request leaked into OfficeCLI: %#v", pptx)
	}
}

func containsMountedSource(sources []MountedSource, id string) bool {
	for _, source := range sources {
		if source.ID == id {
			return true
		}
	}
	return false
}

func TestRoutePresentationAndNoNuwa(t *testing.T) {
	items := []Manifest{{ID: "presentation", Triggers: []string{"ppt"}, Status: "primary", PrimarySkill: "presentations:Presentations"}}
	got := Route("做一个高级PPT", items)
	if got.Capability != "presentation" || got.Nuwa || got.WriteAuthority != "assigned-execution-nodes-only; scoped-artifact-write; staff-and-aji-read-only" {
		t.Fatalf("unexpected route: %#v", got)
	}
	if len(got.Workers) != 1 || !got.Workers[0].Writes || got.Workers[0].ID != "task-execution" {
		t.Fatalf("presentation artifact was not assigned to a scoped execution node: %#v", got.Workers)
	}
	if len(got.Officers) != 0 {
		t.Fatalf("officers must stay cold: %#v", got.Officers)
	}
	if len(got.OfficerWorkers) != 0 {
		t.Fatalf("officer workers must stay cold: %#v", got.OfficerWorkers)
	}
	if len(got.FinishLine) < 4 {
		t.Fatalf("finish line missing fused claim guard: %#v", got.FinishLine)
	}
}

func TestSearchFansOutIndependentSources(t *testing.T) {
	items := []Manifest{{ID: "search", Triggers: []string{"全网搜索"}, Status: "callable", Engines: []Engine{{ID: "web-research", Default: true}}}}
	got := Route("全网搜索上下文压缩方案", items)
	if !got.Parallel || len(got.Workers) != 3 {
		t.Fatalf("expected three parallel source branches: %#v", got.Workers)
	}
	for _, worker := range got.Workers {
		if worker.Writes {
			t.Fatal("research workers must not own writes")
		}
		if !worker.ExecutionEvidenceRequired || len(worker.ExecutionEvidenceFields) != len(workerExecutionEvidenceFields) {
			t.Fatalf("worker can be reported complete without execution evidence: %#v", worker)
		}
	}
}

func TestExplicitFullWebResearchOverridesPriorArtPreflight(t *testing.T) {
	items := []Manifest{
		{ID: "presentation", Triggers: []string{"ppt"}, Status: "behavior-verified", Engines: []Engine{{ID: "editable-pptx", Default: true}}},
		{ID: "search", Triggers: []string{"全网搜索"}, Status: "callable", Engines: []Engine{{ID: "web-research", Default: true}}},
	}
	got := Route("全网搜索 PPT 1.0 和 2.0 的能力融合方式", items)
	if got.Capability != "search" || got.SearchFirstPolicy.Required || len(got.PreflightWorkers) != 0 || len(got.Workers) != 3 {
		t.Fatalf("explicit full-web research was mistaken for a bounded preflight: %#v", got)
	}
	for _, worker := range got.Workers {
		if worker.MaxSources != 0 || worker.TimeBudgetSeconds != fullResearchTimeBudgetSec || !containsString(worker.StopConditions, "two successive relevant sources add no material claim or contradiction") {
			t.Fatalf("full-web worker retained an arbitrary source cap: %#v", worker)
		}
	}
}

func TestSpecializedExtractionDoesNotFanOut(t *testing.T) {
	items := []Manifest{{
		ID: "search", Triggers: []string{"youtube字幕"}, Status: "callable",
		Engines: []Engine{{ID: "web-research", Default: true}, {ID: "content-extraction", Triggers: []string{"youtube字幕"}}},
	}}
	got := Route("提取youtube字幕", items)
	if got.Engine != "content-extraction" || got.Parallel || len(got.Workers) != 1 || got.Workers[0].Model != "gpt-5.6-terra" {
		t.Fatalf("specialized extraction did not receive bounded task judgment: %#v", got)
	}
}

func TestOfficerRequiresExplicitTerm(t *testing.T) {
	items := []Manifest{{ID: "code", Triggers: []string{"修复"}, Status: "callable"}}
	quiet := Route("修复一个接口问题", items)
	if len(quiet.Officers) != 0 {
		t.Fatal("routine repair launched an officer")
	}
	if len(quiet.OfficerWorkers) != 0 {
		t.Fatal("routine repair created an officer task")
	}
	explicit := Route("白帽检查这个修复", items)
	if len(explicit.Officers) != 1 || explicit.Officers[0] != "white-hat" {
		t.Fatalf("explicit officer not routed: %#v", explicit.Officers)
	}
	if len(explicit.OfficerWorkers) != 1 {
		t.Fatalf("explicit officer has no executable task: %#v", explicit.OfficerWorkers)
	}
	worker := explicit.OfficerWorkers[0]
	if worker.ID != "officer-white-hat" || worker.Stage != "officer" || worker.Writes || !worker.ExecutionEvidenceRequired || worker.SessionKey == "" || worker.TaskContractSHA256 == "" {
		t.Fatalf("explicit officer task is not independently executable: %#v", worker)
	}
}

func TestImageProviderOverridesAndFallback(t *testing.T) {
	items := []Manifest{{
		ID: "image", Triggers: []string{"生图"}, Status: "callable",
		Providers: []Provider{
			{ID: "agnes-image", Default: true, Fallback: "default-gpt-image"},
			{ID: "xiaobai-image2", Triggers: []string{"小白"}, Fallback: "default-gpt-image"},
		},
	}}
	defaultRoute := Route("请生图", items)
	if defaultRoute.Provider != "agnes-image" || defaultRoute.ProviderFallback != "default-gpt-image" {
		t.Fatalf("wrong default provider: %#v", defaultRoute)
	}
	override := Route("用小白生图", items)
	if override.Provider != "xiaobai-image2" {
		t.Fatalf("explicit provider override lost: %#v", override)
	}
}

func TestNaturalEnglishImageTrigger(t *testing.T) {
	items := []Manifest{{ID: "image", Triggers: []string{"generate an image"}, Status: "callable", PrimarySkill: "wuji-image-suite", Providers: []Provider{{ID: "agnes-image", Default: true}}}}
	got := Route("generate an image for this slide", items)
	if got.Capability != "image" || got.PrimarySkill != "wuji-image-suite" || got.Provider != "agnes-image" {
		t.Fatalf("natural English image request missed route: %#v", got)
	}
}

func TestMostSpecificProviderTriggerWins(t *testing.T) {
	items := []Manifest{{
		ID: "image", Triggers: []string{"image"}, Status: "callable",
		Providers: []Provider{
			{ID: "generic", Default: true, Triggers: []string{"image"}},
			{ID: "image2", Triggers: []string{"image2"}},
		},
	}}
	got := Route("use image2", items)
	if got.Provider != "image2" {
		t.Fatalf("specific provider trigger lost to generic trigger: %#v", got)
	}
}

func TestCapabilitySubEngineSelection(t *testing.T) {
	items := []Manifest{{
		ID: "presentation", Triggers: []string{"ppt", "slidev", "流体背景"}, Status: "primary",
		Engines: []Engine{
			{ID: "editable-pptx", Default: true, PrimarySkill: "wuji-editable-deck"},
			{ID: "web-deck", PrimarySkill: "wuji-web-deck", Triggers: []string{"slidev", "流体背景"}},
		},
	}}
	if got := Route("用 Slidev 做开发者演示", items); got.Engine != "web-deck" || got.PrimarySkill != "wuji-web-deck" {
		t.Fatalf("wrong Slidev engine: %#v", got)
	}
	if got := Route("给PPT增加流体背景", items); got.Engine != "web-deck" {
		t.Fatalf("wrong fluid engine: %#v", got)
	}
	if got := Route("做一个普通PPT", items); got.Engine != "editable-pptx" || got.PrimarySkill != "wuji-editable-deck" {
		t.Fatalf("wrong default engine: %#v", got)
	}
	web := Route("用 Slidev 做开发者演示", items)
	var prefix struct {
		PrimarySkill string `json:"primary_skill"`
		Engine       string `json:"engine"`
	}
	if err := json.Unmarshal([]byte(web.Workers[0].StableCapabilityPrefix), &prefix); err != nil || prefix.PrimarySkill != "wuji-web-deck" || prefix.Engine != "web-deck" {
		t.Fatalf("worker prefix did not preserve selected engine Skill: prefix=%#v err=%v", prefix, err)
	}
}

func TestSemanticAtomCanSelectItsContainingCapability(t *testing.T) {
	root := t.TempDir()
	dashi := filepath.Join(root, "dashiai")
	if err := os.MkdirAll(dashi, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dashi, "SKILL.md"), []byte("# Dashi"), 0o644); err != nil {
		t.Fatal(err)
	}
	items := []Manifest{{
		ID: "presentation", Triggers: []string{"ppt", "可编辑html演示"}, Status: "callable", PrimarySkill: "wuji-editable-deck", Root: root,
		Engines: []Engine{{ID: "editable-pptx", Default: true, PrimarySkill: "wuji-editable-deck"}, {ID: "web-deck", PrimarySkill: "wuji-web-deck", Triggers: []string{"可编辑html演示"}}},
		Sources: []Source{{ID: "dashiai-ppt-local", Engine: "web-deck", Priority: "secondary", Lifecycle: "callable", Activation: []string{"可编辑html演示"}, Entrypoint: "SKILL.md", Globs: []string{dashi}, Required: []string{"SKILL.md"}}},
	}}
	got := Route("做一个可编辑HTML演示", items)
	if got.Capability != "presentation" || got.Engine != "web-deck" || got.PrimarySkill != "wuji-web-deck" || len(got.MountedSources) != 1 || got.MountedSources[0].ID != "dashiai-ppt-local" {
		t.Fatalf("semantic atom did not reach and activate its containing route: %#v", got)
	}
	got = Route("create an editable HTML deck", []Manifest{{
		ID: "presentation", Triggers: []string{"ppt", "editable html deck"}, Status: "callable", PrimarySkill: "wuji-editable-deck", Root: root,
		Engines: []Engine{{ID: "editable-pptx", Default: true, PrimarySkill: "wuji-editable-deck"}, {ID: "web-deck", PrimarySkill: "wuji-web-deck", Triggers: []string{"editable html deck"}}},
		Sources: []Source{{ID: "dashiai-ppt-local", Engine: "web-deck", Priority: "secondary", Lifecycle: "callable", Activation: []string{"editable html deck"}, Entrypoint: "SKILL.md", Globs: []string{dashi}, Required: []string{"SKILL.md"}}},
	}})
	if got.Capability != "presentation" || got.Engine != "web-deck" || len(got.MountedSources) != 1 || got.MountedSources[0].ID != "dashiai-ppt-local" {
		t.Fatalf("English semantic atom did not reach and activate its containing route: %#v", got)
	}
}

func TestExplicitEngineWinsWhenDefaultIsDeclaredLast(t *testing.T) {
	items := []Manifest{{
		ID: "writing", Triggers: []string{"translate"}, Status: "callable", PrimarySkill: "wuji-writing-suite",
		Engines: []Engine{
			{ID: "translation", Triggers: []string{"translate"}},
			{ID: "long-form", Default: true},
		},
	}}
	got := Route("translate this article", items)
	if got.Engine != "translation" {
		t.Fatalf("explicit engine was overwritten by a later default: %#v", got)
	}
}

func TestUnverifiedCapabilityUsesFallback(t *testing.T) {
	items := []Manifest{{ID: "review", Triggers: []string{"review"}, Status: "assets-retained", PrimarySkill: "unverified-tool", Fallback: "native-review"}}
	got := Route("review this", items)
	if got.PrimarySkill != "native-review" || len(got.MountedSources) != 0 {
		t.Fatalf("unverified capability pretended to activate: %#v", got)
	}
}

func TestPullRequestRoutesToCodeReview(t *testing.T) {
	items := []Manifest{
		{ID: "code", Triggers: []string{"fix", "review", "pull request"}, Status: "callable", PrimarySkill: "native"},
		{ID: "code-review", Triggers: []string{"code review", "pull request", "pr review"}, Status: "callable", PrimarySkill: "open-code-review"},
	}
	got := Route("review this pull request", items)
	if got.Capability != "code-review" {
		t.Fatalf("expected code-review, got %#v", got)
	}
}

func TestDesignSystemRoutesToVisual(t *testing.T) {
	items := []Manifest{
		{ID: "core", Triggers: []string{"polish"}, Status: "primary", PrimarySkill: "wuji-legion-codex-3-0"},
		{ID: "frontend", Triggers: []string{"ui", "page", "polish"}, Status: "callable", PrimarySkill: "wuji-frontend-suite"},
		{ID: "visual", Triggers: []string{"design system", "polish design", "taste"}, Status: "callable", PrimarySkill: "wuji-visual-suite"},
	}
	got := Route("polish the design system", items)
	if got.Capability != "visual" {
		t.Fatalf("expected visual, got %#v", got)
	}
}

func TestTokenUsageRoutesToContextNotSearch(t *testing.T) {
	items := []Manifest{
		{ID: "search", Triggers: []string{"search", "token"}, Status: "callable", PrimarySkill: "wuji-research-suite"},
		{ID: "context", Triggers: []string{"context", "token", "repo map"}, Status: "primary", PrimarySkill: "wuji context-select"},
	}
	got := Route("search code for token usage", items)
	if got.Capability != "context" {
		t.Fatalf("expected context, got %#v", got)
	}
}

func TestPagePolishRoutesToFrontend(t *testing.T) {
	items := []Manifest{
		{ID: "frontend", Triggers: []string{"页面", "前端", "美化"}, Status: "callable", PrimarySkill: "wuji-frontend-suite"},
		{ID: "visual", Triggers: []string{"美化", "视觉", "设计系统"}, Status: "callable", PrimarySkill: "wuji-visual-suite"},
	}
	got := Route("美化这个页面", items)
	if got.Capability != "frontend" {
		t.Fatalf("expected frontend, got %#v", got)
	}
}

func TestDynamicBoardRoutesToVisual(t *testing.T) {
	items := []Manifest{
		{ID: "frontend", Triggers: []string{"页面", "电子看板", "dashboard"}, Status: "callable", PrimarySkill: "wuji-frontend-suite"},
		{ID: "visual", Triggers: []string{"动态看板", "电子看板", "hud", "视觉"}, Status: "callable", PrimarySkill: "wuji-visual-suite"},
	}
	got := Route("动态看板 HUD", items)
	if got.Capability != "visual" {
		t.Fatalf("expected visual, got %#v", got)
	}
}

func TestMultiIntentSecondaryCapabilities(t *testing.T) {
	items := []Manifest{
		{ID: "writing", Triggers: []string{"写文章", "文章"}, Status: "callable", PrimarySkill: "wuji-writing-suite"},
		{ID: "image", Triggers: []string{"生图", "配图"}, Status: "callable", PrimarySkill: "wuji-image-suite"},
		{ID: "video", Triggers: []string{"视频", "做成视频"}, Status: "callable", PrimarySkill: "wuji-video-suite"},
	}
	article := Route("写文章并配图", items)
	if article.Capability != "writing" || len(article.SecondaryCapabilities) != 1 || article.SecondaryCapabilities[0] != "image" {
		t.Fatalf("expected writing+image secondary: %#v", article)
	}
	clip := Route("生成图片并做成视频", items)
	if clip.Capability != "image" && clip.Capability != "video" {
		t.Fatalf("expected image or video primary: %#v", clip)
	}
	if clip.Capability == "video" {
		if len(clip.SecondaryCapabilities) == 0 || clip.SecondaryCapabilities[0] != "image" {
			t.Fatalf("expected image secondary for video primary: %#v", clip)
		}
	}
	if clip.Capability == "image" {
		if len(clip.SecondaryCapabilities) == 0 || clip.SecondaryCapabilities[0] != "video" {
			t.Fatalf("expected video secondary for image primary: %#v", clip)
		}
	}
}

func TestSparseSourceMount(t *testing.T) {
	root := t.TempDir()
	primaryDir := root + "/primary"
	secondaryDir := root + "/secondary"
	if err := os.MkdirAll(primaryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(secondaryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{primaryDir, secondaryDir} {
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	items := []Manifest{{
		ID: "presentation", Triggers: []string{"ppt"}, Status: "primary", PrimarySkill: "wuji-editable-deck", Root: root,
		Sources: []Source{
			{ID: "wuji-editable-deck-unified", Priority: "primary", Lifecycle: "callable", Entrypoint: "SKILL.md", Globs: []string{primaryDir}, Required: []string{"SKILL.md"}},
			{ID: "humanize-ppt-complete", Priority: "secondary", Lifecycle: "callable", Activation: []string{"叙事结构"}, Entrypoint: "SKILL.md", Globs: []string{secondaryDir}, Required: []string{"SKILL.md"}},
			{ID: "optional-atom", Priority: "optional", Globs: []string{secondaryDir}, Required: []string{"SKILL.md"}},
		},
	}}
	sparse := Route("做一个PPT", items)
	if len(sparse.MountedSources) != 1 || sparse.MountedSources[0].ID != "wuji-editable-deck-unified" {
		t.Fatalf("sparse mount failed: %#v", sparse.MountedSources)
	}
	automatic := Route("做一个有叙事结构的PPT", items)
	if len(automatic.MountedSources) != 2 || automatic.MountedSources[1].ID != "humanize-ppt-complete" || automatic.MountedSources[1].Entrypoint != "SKILL.md" || automatic.MountedSources[1].ActivationReason != "semantic-trigger:叙事结构" {
		t.Fatalf("semantic source activation failed: %#v", automatic.MountedSources)
	}
	named := Route("做一个PPT 使用 humanize-ppt", items)
	if len(named.MountedSources) != 2 || named.MountedSources[1].ActivationReason != "explicit-source-request" {
		t.Fatalf("named secondary should mount: %#v", named.MountedSources)
	}
	full := Route("做一个PPT 完整能力", items)
	if len(full.MountedSources) != 2 {
		t.Fatalf("full mount must not promote cold optional material: %#v", full.MountedSources)
	}
}

func TestRouteDoesNotMountIncompleteSource(t *testing.T) {
	root := t.TempDir()
	incomplete := filepath.Join(root, "incomplete")
	if err := os.MkdirAll(incomplete, 0o755); err != nil {
		t.Fatal(err)
	}
	items := []Manifest{{
		ID: "code", Triggers: []string{"code"}, Status: "callable", PrimarySkill: "native", Root: root,
		Sources: []Source{{ID: "broken", Priority: "primary", Globs: []string{incomplete}, Required: []string{"SKILL.md"}}},
	}}
	if got := Route("code task", items); len(got.MountedSources) != 0 {
		t.Fatalf("incomplete source was mounted: %#v", got.MountedSources)
	}
}

func TestHighRiskRoutingRequiresStrictChangeCapsule(t *testing.T) {
	items := []Manifest{{ID: "code", Triggers: []string{"路由", "代码"}, Status: "callable", PrimarySkill: "native"}}
	highRisk := Route("迁移路由架构并调整模型策略", items)
	if !highRisk.ChangeCapsule.Required || !highRisk.ChangeCapsule.Strict || highRisk.ChangeCapsule.Reason != "high-risk-change-boundary" {
		t.Fatalf("high-risk route omitted strict change capsule: %#v", highRisk.ChangeCapsule)
	}
	routine := Route("修复一个代码拼写", items)
	if routine.ChangeCapsule.Required {
		t.Fatalf("routine change unnecessarily required a capsule: %#v", routine.ChangeCapsule)
	}
}

func TestPresentationDelegationRequiresExplicitSelfContainedHandoff(t *testing.T) {
	items := []Manifest{{
		ID: "presentation", Triggers: []string{"ppt"}, Status: "primary",
		Experts: []Expert{
			{ID: "narrative", Purpose: "plan", Independent: true, ModelClass: "sol"},
			{ID: "visual", Purpose: "visual", Independent: true, ModelClass: "sol"},
			{ID: "qa", Purpose: "qa", Independent: false, ModelClass: "luna"},
		},
	}}
	direct := Route("做一个PPT", items)
	if direct.Parallel || len(direct.Workers) != 1 || direct.Workers[0].Model != "gpt-5.6-terra" || direct.DelegationDecision.Reason != "default-task-reasoning" {
		t.Fatalf("presentation did not receive its bounded default task route: %#v", direct)
	}
	missingHandoff := Route("做一个PPT parallel", items)
	if missingHandoff.Parallel || len(missingHandoff.Workers) != 1 || missingHandoff.Workers[0].ID != "task-judgment" || missingHandoff.DelegationDecision.Reason != "self-contained-handoff-required" {
		t.Fatalf("parallel presentation did not fall back to one bounded task judgment: %#v", missingHandoff)
	}
	parallel := RouteWithContext("做一个PPT parallel", items, DelegationContext{SelfContained: true})
	if !parallel.Parallel || len(parallel.Workers) != 2 {
		t.Fatalf("explicit self-contained presentation did not fan out: %#v", parallel.Workers)
	}
	parentDependent := RouteWithContext("create a PPT from the preceding meeting transcript parallel", items, DelegationContext{SelfContained: true})
	if !parentDependent.Parallel || len(parentDependent.Workers) != 2 {
		t.Fatalf("self-contained parent-dependent presentation did not route its independent branches: %#v", parentDependent)
	}
	serial := Route("做一个PPT 串行", items)
	if serial.Parallel || len(serial.Workers) != 1 || serial.Workers[0].ID != "task-judgment" {
		t.Fatalf("serial keyword should retain one sequential task route: %#v", serial.Workers)
	}
}

func TestPresentationRouteInjectsEngineBoundFusionAsset(t *testing.T) {
	root := t.TempDir()
	sourceRoot := filepath.Join(root, "presentation")
	for path, content := range map[string]string{
		"scripts/invoke-presentation.ps1": "param()\n",
		"assets/editable.json":            "editable catalog",
		"assets/web.md":                   "web deck",
		"assets/fluid.js":                 "fluid deck",
	} {
		full := filepath.Join(sourceRoot, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	manifest := Manifest{
		Root: root, ID: "presentation", Description: "presentation", Triggers: []string{"ppt", "slidev"}, Status: "behavior-verified", PrimarySkill: "wuji-editable-deck", DirectMount: true, Fallback: "fallback",
		Probe:   &Probe{Command: "success", Kind: "behavior", Fixture: "presentation-fixture", RequiredEvidence: []string{"assertions"}, ComparisonEvidence: "assertions"},
		Engines: []Engine{{ID: "editable-pptx", Default: true, PrimarySkill: "wuji-editable-deck", Triggers: []string{"pptx"}}, {ID: "web-deck", PrimarySkill: "wuji-web-deck", Triggers: []string{"slidev", "stage fluid"}}},
		Sources: []Source{
			{ID: "presentation-fusion-runtime", Engine: "editable-pptx", Lifecycle: "callable", Entrypoint: "scripts/invoke-presentation.ps1", Globs: []string{"${ROOT}/presentation"}, Required: []string{"scripts/invoke-presentation.ps1"}},
			{ID: "presentation-web-runtime", Engine: "web-deck", Lifecycle: "callable", Entrypoint: "scripts/invoke-presentation.ps1", Globs: []string{"${ROOT}/presentation"}, Required: []string{"scripts/invoke-presentation.ps1"}},
		},
		Genome: &FusionGenome{SchemaVersion: 1, Species: "presentation-fusion", Revision: "presentation-2-1", Adapters: []FusionAdapter{
			{ID: "editable-pptx", Domain: "editable-pptx", Source: "presentation-fusion-runtime", Entrypoint: "scripts/invoke-presentation.ps1", AssetContracts: []FusionAsset{{ID: "editable-default", Path: "assets/editable.json", Compatibility: []string{"editable-pptx", "default"}}}},
			{ID: "web-deck", Domain: "web-deck", Source: "presentation-fusion-runtime", Entrypoint: "scripts/invoke-presentation.ps1", AssetContracts: []FusionAsset{{ID: "web-default", Path: "assets/web.md", Compatibility: []string{"web-deck", "default"}}, {ID: "web-fluid", Path: "assets/fluid.js", Compatibility: []string{"web-deck", "stage-fluid"}}}},
		}},
	}
	if err := ValidateManifest(manifest); err != nil {
		t.Fatalf("presentation fusion fixture is invalid: %v", err)
	}
	editable := Route("create an editable PPTX", []Manifest{manifest})
	if len(editable.AssetContracts) != 1 || editable.AssetContracts[0].AssetID != "presentation:editable-default" || len(editable.Workers) != 1 {
		t.Fatalf("editable presentation route omitted its trusted asset: %#v", editable)
	}
	worker := editable.Workers[0]
	if len(worker.AssetContracts) != 1 || worker.AssetContracts[0].AssetID != editable.AssetContracts[0].AssetID || !strings.Contains(worker.StableCapabilityPrefix, "presentation:editable-default") || len(worker.SourceExecution) != 1 {
		t.Fatalf("worker did not receive its asset contract and invocation entrypoint: %#v", worker)
	}
	fluid := Route("build a Slidev web presentation with stage fluid", []Manifest{manifest})
	if fluid.Engine != "web-deck" || len(fluid.AssetContracts) != 1 || fluid.AssetContracts[0].AssetID != "presentation:web-fluid" {
		t.Fatalf("fluid route leaked the default or editable asset: %#v", fluid)
	}
	if _, err := SelectFusionAsset([]Manifest{manifest}, AssetSelectionRequest{Capability: "presentation", Domain: "web-deck", Compatibility: []string{"editable-pptx"}}); err == nil {
		t.Fatal("incompatible presentation asset selection was accepted")
	}
}
