package core

import (
	"strings"
	"testing"
)

func TestRoleGraphCoversDirectAndDelegatedRoles(t *testing.T) {
	direct := Route("无极军团是什么？", nil)
	if direct.RoleGraph.RoleID != "aji" || len(direct.RoleGraphs) != 1 ||
		direct.RoleGraph.MoE != "sparse-role-moe" || direct.RoleGraph.PonyTail != ponyTailDoctrine {
		t.Fatalf("direct route lacks the minimal Aji role graph: %#v", direct)
	}
	if err := validateRoleTaskGraph(direct.RoleGraph); err != nil {
		t.Fatal(err)
	}

	items := []Manifest{{
		ID: "presentation", Status: "callable", Triggers: []string{"ppt"},
		Experts: []Expert{
			{ID: "narrative", Purpose: "build narrative", Independent: true, ModelClass: "sol"},
			{ID: "visual", Purpose: "build visuals", Independent: true, ModelClass: "sol"},
		},
	}}
	routed := RouteWithContext("做一个PPT parallel", items, DelegationContext{SelfContained: true})
	for _, graph := range routed.RoleGraphs {
		if err := validateRoleTaskGraph(graph); err != nil {
			t.Fatalf("invalid role graph %s: %v", graph.RoleID, err)
		}
	}
	for _, worker := range routed.Workers {
		if worker.TaskGraphID == "" || worker.TaskGraphSHA256 == "" || worker.PonyTail != ponyTailDoctrine {
			t.Fatalf("worker lacks graph identity or PonyTail: %#v", worker)
		}
		if worker.AllocatedTaskContractBytes > maxTaskContractBytes {
			t.Fatalf("graph metadata exceeded worker contract budget: %d", worker.AllocatedTaskContractBytes)
		}
	}
	if routed.RoleGraphs["general-staff"].RoleKind != "staff" ||
		routed.RoleGraphs["commander:presentation"].RoleKind != "commander" ||
		routed.RoleGraphs["verification"].RoleKind != "verification" ||
		routed.RoleGraphs["aji-report"].RoleKind != "report" {
		t.Fatalf("delegated route did not create every logical role graph: %#v", routed.RoleGraphs)
	}
	for _, node := range routed.RoleGraph.Nodes {
		if node.State == "complete" || node.State == "verified" || node.State == "active" {
			t.Fatalf("route preparation claimed role execution without host evidence: %#v", node)
		}
	}
}

func TestRoleGraphRejectsCyclesAndTamperedHashes(t *testing.T) {
	graph, err := makeRoleTaskGraph("role", "expert", "test", "3.0", "", "a", []RoleTaskNode{
		roleGraphNode("a", "role", "intake", "start", []string{"b"}, "", false),
		roleGraphNode("b", "role", "execute", "finish", []string{"a"}, "", false),
	}, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("cyclic role graph was accepted: %v", err)
	}

	graph, err = makeRoleTaskGraph("role", "expert", "test", "3.0", "", "a", []RoleTaskNode{
		roleGraphNode("a", "role", "intake", "start", nil, "", false),
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	graph.Nodes[0].Objective = "tampered"
	if err := validateRoleTaskGraph(graph); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("tampered role graph was accepted: %v", err)
	}
}

func TestRoleWorkerBatchesRespectDependenciesAndWrites(t *testing.T) {
	graph, err := makeRoleTaskGraph("aji", "aji", "composed", "3.0", "", "root", []RoleTaskNode{
		roleGraphNode("root", "aji", "intake", "start", nil, "", false),
		roleGraphNode("worker-a", "a", "worker", "a", []string{"root"}, "a", true),
		roleGraphNode("worker-b", "b", "worker", "b", []string{"worker-a"}, "b", true),
		roleGraphNode("worker-c", "c", "worker", "c", []string{"root"}, "c", true),
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	workers := []WorkerTask{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	batches, err := roleWorkerBatches(graph, workers, "worker-")
	if err != nil {
		t.Fatal(err)
	}
	if len(batches) != 2 || len(batches[0]) != 2 || len(batches[1]) != 1 || batches[1][0].ID != "b" {
		t.Fatalf("worker dependency batches are wrong: %#v", batches)
	}
}

func TestExplicitCompositeRouteCreatesSecondaryCommanderGraphs(t *testing.T) {
	manifests, err := LoadManifests(expertTestRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	route := Route("写文章并配图", manifests)
	if route.Capability != "writing" || !containsString(route.SecondaryCapabilities, "image") {
		t.Fatalf("test did not select a writing plus image composition: %#v", route)
	}
	if len(route.Workers) < 2 || route.RoleGraphs["commander:writing"].RoleKind != "commander" ||
		route.RoleGraphs["commander:image"].RoleKind != "commander" {
		t.Fatalf("secondary commander was not executable: %#v", route)
	}
	if route.RoleGraphs["commander:writing"].GraphID == route.RoleGraphs["commander:image"].GraphID {
		t.Fatal("secondary commander reused the primary graph identity")
	}
}
