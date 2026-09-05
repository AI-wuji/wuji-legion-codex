package core

import "testing"

func TestSelectOfficerRecommendationsLimitsIndependentReviewToOneCompositeCall(t *testing.T) {
	if got := SelectOfficerRecommendations("rename a local variable"); len(got) != 0 {
		t.Fatalf("ordinary task selected an officer: %#v", got)
	}
	medium := SelectOfficerRecommendations("review and test a bounded feature")
	if len(medium) != 1 || medium[0].Role != "internal-quality-check" || medium[0].Decision != "internal-quality-check" || medium[0].Contract.RequiresUserConfirmation {
		t.Fatalf("medium task did not remain an internal, non-independent quality check: %#v", medium)
	}
	got := SelectOfficerRecommendations("migrate routing architecture and publish a release")
	if len(got) != 1 || got[0].Role != "composite-moe-officer" || got[0].Decision != "independent-composite-quality-inspection-with-audit" || len(got[0].Contract.Stages) != 3 || got[0].Contract.Stages[2] != "audit" {
		t.Fatalf("governance risk did not remain a single composite review with audit section: %#v", got)
	}
}

func TestRouteStartsOneCompositeOfficerForLargeOrHighRiskTask(t *testing.T) {
	route := Route("migrate routing architecture", nil)
	if len(route.OfficerRecommendations) != 1 || len(route.OfficerWorkers) != 1 || len(route.Officers) != 1 || route.Officers[0] != "composite-moe" {
		t.Fatalf("large or high-risk route did not start exactly one composite officer: %#v", route)
	}
	worker := route.OfficerWorkers[0]
	if worker.ID != "officer-composite-moe" || worker.Model != "gpt-5.6-sol" || worker.Writes {
		t.Fatalf("composite officer contract is invalid: %#v", worker)
	}
}
