package core

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestBehaviorClaimRequiresProbe(t *testing.T) {
	manifest := validManifest("fake", "behavior-verified")
	got := Verify(t.TempDir(), manifest)
	if got.Passed {
		t.Fatal("behavior claim passed without a probe")
	}
}

func TestNativeHostCapabilityCanBeCallable(t *testing.T) {
	got := Verify(t.TempDir(), validManifest("native", "callable"))
	if !got.Passed || got.Effective != "callable" {
		t.Fatalf("native host capability should be callable: %#v", got)
	}
}

func TestEngineRequiresSourceCoverage(t *testing.T) {
	manifest := validManifest("bad-engine", "callable")
	manifest.Engines = []Engine{
		{ID: "one", Default: true, PrimarySkill: "native"},
		{ID: "two", PrimarySkill: "native", Triggers: []string{"two"}},
	}
	got := Verify(t.TempDir(), manifest)
	if got.Passed {
		t.Fatal("engine without a complete source package passed")
	}
}

func TestVerifyRejectsUnknownLifecycleStatus(t *testing.T) {
	manifest := validManifest("candidate", "callable")
	manifest.Status = "fused"
	got := Verify(t.TempDir(), manifest)
	if got.Passed {
		t.Fatal("unknown lifecycle status passed verification")
	}
}

func TestVerifyTimesOutHungProbe(t *testing.T) {
	t.Setenv("GO_WANT_WUJI_PROBE_HELPER", "1")
	manifest := validBehaviorManifest("timeout-probe", "hang")
	manifest.Probe.TimeoutSeconds = 1
	started := time.Now()
	got := Verify(t.TempDir(), manifest)
	if got.Passed || time.Since(started) > 5*time.Second {
		t.Fatalf("hung probe was not bounded: elapsed=%s result=%#v", time.Since(started), got)
	}
	if len(got.Errors) == 0 || !strings.Contains(got.Errors[0], "timed out") {
		t.Fatalf("timeout was not diagnosable: %#v", got.Errors)
	}
}

func TestProbeHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_WUJI_PROBE_HELPER") != "1" {
		return
	}
	mode := os.Args[len(os.Args)-1]
	if mode == "hang" {
		time.Sleep(30 * time.Second)
	}
	fmt.Print("probe-ok")
	os.Exit(0)
}

func validManifest(id, status string) Manifest {
	return Manifest{
		ID:           id,
		Description:  "test capability",
		Triggers:     []string{"test"},
		Status:       status,
		PrimarySkill: "native",
		HostCallable: true,
		Fallback:     "native fallback",
	}
}

func validBehaviorManifest(id, mode string) Manifest {
	manifest := validManifest(id, "behavior-verified")
	manifest.Probe = &Probe{
		Command: os.Args[0],
		Args:    []string{"-test.run=TestProbeHelperProcess", "--", mode},
		Fixture: "shared-test-fixture-v1",
		Kind:    "behavior",
	}
	return manifest
}
