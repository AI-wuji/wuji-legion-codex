package core

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func TestCallableClaimRequiresExecutableProbe(t *testing.T) {
	manifest := validManifest("native", "callable")
	manifest.Probe = nil
	got := Verify(t.TempDir(), manifest)
	if got.Passed || got.Effective == "callable" {
		t.Fatalf("host_callable declaration promoted capability without a probe: %#v", got)
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
	testBinary, err := filepath.Abs(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Dir(testBinary)
	manifest := validBehaviorManifest("root-command", "root-command")
	manifest.Probe.Command = "${ROOT}/" + filepath.Base(testBinary)
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

func TestBehaviorProbeRequiresStructuredEvidence(t *testing.T) {
	t.Setenv("GO_WANT_WUJI_PROBE_HELPER", "1")
	manifest := validBehaviorManifest("plain-output", "plain")
	got := Verify(t.TempDir(), manifest)
	if got.Passed || got.Effective == "behavior-verified" {
		t.Fatalf("plain probe output was accepted as behavior evidence: %#v", got)
	}
}

func TestBehaviorProbeRejectsSelfAttestedReceipt(t *testing.T) {
	t.Setenv("GO_WANT_WUJI_PROBE_HELPER", "1")
	manifest := validBehaviorManifest("self-attested", "self-attested")
	got := Verify(t.TempDir(), manifest)
	if got.Passed || got.Effective == "behavior-verified" {
		t.Fatalf("self-attested receipt was accepted without a verified artifact: %#v", got)
	}
}

func TestPrimaryClaimRequiresVerifiedPromotionReceipt(t *testing.T) {
	t.Setenv("GO_WANT_WUJI_PROBE_HELPER", "1")
	manifest := validBehaviorManifest("unpromoted", "success")
	manifest.Status = "primary"
	got := Verify(t.TempDir(), manifest)
	if got.Passed || got.Effective == "primary" {
		t.Fatalf("primary claim passed without a promotion receipt: %#v", got)
	}
}

func TestLimitedOutputHandlesConcurrentWriters(t *testing.T) {
	output := &limitedOutput{}
	payload := []byte(strings.Repeat("x", 1024))
	var writers sync.WaitGroup
	for range 16 {
		writers.Add(1)
		go func() {
			defer writers.Done()
			for range 16 {
				if _, err := output.Write(payload); err != nil {
					t.Errorf("write failed: %v", err)
				}
			}
		}()
	}
	writers.Wait()
	got := output.String()
	if !strings.Contains(got, "output truncated") || len(output.data) != maxProbeOutputBytes {
		t.Fatalf("concurrent output was not bounded: bytes=%d", len(output.data))
	}
}

func TestProbeHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_WUJI_PROBE_HELPER") != "1" {
		return
	}
	mode := os.Args[len(os.Args)-2]
	fixture := os.Args[len(os.Args)-1]
	if mode == "hang" {
		time.Sleep(30 * time.Second)
	}
	if mode == "large-output" {
		path, digest := writeProbeArtifact(t, mode)
		fmt.Printf(`{"wuji_probe":"behavior","fixture":"%s","passed":true,"evidence":[{"id":"artifact","path":"%s","sha256":"%s"}],"signature":"shared-behavior-v1"}`+"\n", fixture, path, digest)
		fmt.Print(strings.Repeat("x", 128*1024))
		os.Exit(0)
	}
	if mode == "plain" {
		fmt.Print("probe-ok")
		os.Exit(0)
	}
	if mode == "self-attested" {
		fmt.Printf(`{"wuji_probe":"behavior","fixture":"%s","passed":true,"evidence":[{"id":"artifact","path":"missing.txt","sha256":"%s"}],"signature":"shared-behavior-v1"}`, fixture, strings.Repeat("0", 64))
		os.Exit(0)
	}
	path, digest := writeProbeArtifact(t, mode)
	signature := "shared-behavior-v1"
	if mode == "different-signature" {
		signature = "different-behavior-v1"
	}
	fmt.Printf(`{"wuji_probe":"behavior","fixture":"%s","passed":true,"evidence":[{"id":"artifact","path":"%s","sha256":"%s"}],"signature":"%s"}`, fixture, path, digest, signature)
	os.Exit(0)
}

func writeProbeArtifact(t *testing.T, mode string) (string, string) {
	t.Helper()
	dir := os.Getenv("WUJI_PROBE_EVIDENCE_DIR")
	if dir == "" {
		t.Fatal("probe evidence directory is missing")
	}
	name := "artifact.txt"
	payload := []byte("artifact:" + mode)
	if mode == "different-evidence" {
		payload = []byte("different artifact")
	}
	if err := os.WriteFile(filepath.Join(dir, name), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(payload))
	return name, digest
}

func validManifest(id, status string) Manifest {
	manifest := Manifest{
		ID:           id,
		Description:  "test capability",
		Triggers:     []string{"test"},
		Status:       status,
		PrimarySkill: "native",
		HostCallable: true,
		Fallback:     "native fallback",
	}
	if status == "callable" {
		manifest.Probe = &Probe{Command: "unavailable-test-probe", Kind: "smoke", Fixture: "callable-smoke-v1"}
	}
	return manifest
}

func validBehaviorManifest(id, mode string) Manifest {
	manifest := validManifest(id, "behavior-verified")
	manifest.Probe = &Probe{
		Command:            os.Args[0],
		Args:               []string{"-test.run=TestProbeHelperProcess", "--", mode, "shared-test-fixture-v1"},
		Fixture:            "shared-test-fixture-v1",
		Kind:               "behavior",
		RequiredEvidence:   []string{"artifact"},
		ComparisonEvidence: "artifact",
	}
	return manifest
}
