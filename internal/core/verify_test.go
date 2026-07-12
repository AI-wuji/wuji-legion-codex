package core

import (
	"fmt"
	"os"
	"path/filepath"
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

func TestVerifyExpandsRootInProbeCommand(t *testing.T) {
	t.Setenv("GO_WANT_WUJI_PROBE_HELPER", "1")
	root, err := os.MkdirTemp(".", ".probe-root-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	root, err = filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	testBinary, err := os.ReadFile(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "probe.exe"), testBinary, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := validBehaviorManifest("root-command", "root-command")
	manifest.Probe.Command = "${ROOT}/probe.exe"
	manifest.Probe.TimeoutSeconds = 5
	if got := Verify(root, manifest); !got.Passed {
		t.Fatalf("root-relative probe command failed: %#v", got)
	}
}

func TestVerifyCapsProbeOutput(t *testing.T) {
	t.Setenv("GO_WANT_WUJI_PROBE_HELPER", "1")
	manifest := validBehaviorManifest("large-output", "large-output")
	got := Verify(t.TempDir(), manifest)
	if !got.Passed || len(got.Checks) == 0 || !strings.Contains(got.Checks[len(got.Checks)-1], "output truncated") {
		t.Fatalf("probe output was not capped: %#v", got)
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
	if mode == "large-output" {
		fmt.Print(strings.Repeat("x", 128*1024))
		os.Exit(0)
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
