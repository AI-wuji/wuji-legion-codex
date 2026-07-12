package core

import "testing"

func TestRouteEmitsExecutableModelPolicy(t *testing.T) {
	items := []Manifest{{
		ID: "code", Triggers: []string{"code"}, Status: "callable", PrimarySkill: "native",
		Experts: []Expert{
			{ID: "implementation", Purpose: "implement", Independent: true, ModelClass: "terra"},
			{ID: "verification", Purpose: "verify", Independent: true, ModelClass: "terra"},
		},
	}}
	got := Route("code task parallel", items)
	if got.MainModel != "gpt-5.6-sol" || got.ModelPolicy.ClassModels["terra"] != "gpt-5.6-terra" {
		t.Fatalf("main model policy is incomplete: %#v", got.ModelPolicy)
	}
	if len(got.Workers) != 2 || got.Workers[0].Model != "gpt-5.6-terra" || !equalStrings(got.Workers[0].FallbackModels, []string{"gpt-5.6-luna", "gpt-5.6-sol"}) {
		t.Fatalf("route did not emit an executable Terra policy: %#v", got.Workers)
	}
}

func TestSearchWorkersUseConcreteLunaModel(t *testing.T) {
	items := []Manifest{{
		ID: "search", Triggers: []string{"research"}, Status: "callable",
		Engines: []Engine{{ID: "web-research", Default: true}},
	}}
	got := Route("research the web", items)
	if len(got.Workers) != 3 {
		t.Fatalf("expected three research workers: %#v", got.Workers)
	}
	for _, worker := range got.Workers {
		if worker.Model != "gpt-5.6-luna" || !equalStrings(worker.FallbackModels, []string{"gpt-5.6-terra", "gpt-5.6-sol"}) {
			t.Fatalf("research worker did not receive an executable Luna policy: %#v", worker)
		}
	}
}

func TestUnknownModelClassIsNotSilentlyRoutedToSol(t *testing.T) {
	model, fallbacks := modelSpec("unknown")
	if model != "" || fallbacks != nil {
		t.Fatalf("unknown model class must not silently consume Sol: model=%s fallbacks=%#v", model, fallbacks)
	}
	if err := validateExperts([]Expert{{ID: "bad", Purpose: "invalid model", ModelClass: "terra-"}}); err == nil {
		t.Fatal("manifest validation accepted an unsupported model_class")
	}
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range want {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
