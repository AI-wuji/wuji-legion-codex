package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoutePresentationAndNoNuwa(t *testing.T) {
	items := []Manifest{{ID: "presentation", Triggers: []string{"ppt"}, Status: "primary", PrimarySkill: "presentations:Presentations"}}
	got := Route("做一个高级PPT", items)
	if got.Capability != "presentation" || got.Nuwa || got.WriteAuthority != "aji-only" {
		t.Fatalf("unexpected route: %#v", got)
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

func TestSpecializedExtractionDoesNotFanOut(t *testing.T) {
	items := []Manifest{{
		ID: "search", Triggers: []string{"youtube字幕"}, Status: "callable",
		Engines: []Engine{{ID: "web-research", Default: true}, {ID: "content-extraction", Triggers: []string{"youtube字幕"}}},
	}}
	got := Route("提取youtube字幕", items)
	if got.Engine != "content-extraction" || got.Parallel || len(got.Workers) != 0 {
		t.Fatalf("specialized extraction should stay direct: %#v", got)
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
		{ID: "core", Triggers: []string{"polish"}, Status: "primary", PrimarySkill: "wuji-legion-codex-2-0"},
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
			{ID: "wuji-editable-deck-unified", Priority: "primary", Globs: []string{primaryDir}, Required: []string{"SKILL.md"}},
			{ID: "humanize-ppt-complete", Priority: "secondary", Globs: []string{secondaryDir}, Required: []string{"SKILL.md"}},
			{ID: "optional-atom", Priority: "optional", Globs: []string{secondaryDir}, Required: []string{"SKILL.md"}},
		},
	}}
	sparse := Route("做一个PPT", items)
	if len(sparse.MountedSources) != 1 || sparse.MountedSources[0].ID != "wuji-editable-deck-unified" {
		t.Fatalf("sparse mount failed: %#v", sparse.MountedSources)
	}
	named := Route("做一个PPT 使用 humanize-ppt", items)
	if len(named.MountedSources) != 2 {
		t.Fatalf("named secondary should mount: %#v", named.MountedSources)
	}
	full := Route("做一个PPT 完整能力", items)
	if len(full.MountedSources) != 3 {
		t.Fatalf("full mount should include optional: %#v", full.MountedSources)
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

func TestPresentationDelegationRequiresExplicitSelfContainedHandoff(t *testing.T) {
	items := []Manifest{{
		ID: "presentation", Triggers: []string{"ppt"}, Status: "primary",
		Experts: []Expert{
			{ID: "narrative", Purpose: "plan", Independent: true, ModelClass: "terra"},
			{ID: "visual", Purpose: "visual", Independent: true, ModelClass: "terra"},
			{ID: "qa", Purpose: "qa", Independent: false, ModelClass: "terra"},
		},
	}}
	direct := Route("做一个PPT", items)
	if direct.Parallel || len(direct.Workers) != 0 || direct.DelegationDecision.Reason != "direct-route-by-default" {
		t.Fatalf("presentation delegated by default: %#v", direct)
	}
	missingHandoff := Route("做一个PPT parallel", items)
	if missingHandoff.Parallel || len(missingHandoff.Workers) != 0 || missingHandoff.DelegationDecision.Reason != "self-contained-handoff-required" {
		t.Fatalf("parallel presentation accepted an implicit handoff: %#v", missingHandoff)
	}
	parallel := RouteWithContext("做一个PPT parallel", items, DelegationContext{SelfContained: true})
	if !parallel.Parallel || len(parallel.Workers) != 2 {
		t.Fatalf("explicit self-contained presentation did not fan out: %#v", parallel.Workers)
	}
	parentDependent := RouteWithContext("create a PPT from the preceding meeting transcript parallel", items, DelegationContext{SelfContained: true})
	if parentDependent.Parallel || len(parentDependent.Workers) != 0 || parentDependent.DelegationDecision.Reason != "parent-context-affinity-requires-Aji" {
		t.Fatalf("parent-dependent presentation escaped to Terra: %#v", parentDependent)
	}
	serial := Route("做一个PPT 串行", items)
	if serial.Parallel || len(serial.Workers) != 0 {
		t.Fatalf("serial keyword should disable fanout: %#v", serial.Workers)
	}
}
