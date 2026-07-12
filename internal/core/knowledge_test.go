package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func knowledgeFixture(t *testing.T) (string, string) {
	return knowledgeFixtureAt(t, time.Now().UTC())
}

func knowledgeFixtureAt(t *testing.T, verifiedAt time.Time) (string, string) {
	t.Helper()
	store := t.TempDir()
	evidence := filepath.Join(t.TempDir(), "verification.json")
	receipt := `{"schema_version":1,"type":"wuji-verification-receipt","passed":true,"verifier":"go-test","verified_at":"` + verifiedAt.UTC().Format(time.RFC3339) + `"}`
	if err := os.WriteFile(evidence, []byte(receipt), 0o600); err != nil {
		t.Fatal(err)
	}
	return store, evidence
}

func TestKnowledgeExactQueryUsesNodeIndex(t *testing.T) {
	store, evidence := knowledgeFixture(t)
	record, err := recordKnowledgeAt(store, KnowledgeRecordInput{
		Kind: "solution", Key: "browser timeout", Scope: "global", Summary: "Use the bounded browser wait helper.",
		Location: "https://example.test/solution/browser-timeout", Verification: evidence, Tags: []string{"browser", "timeout"},
	}, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	result, err := QueryKnowledge(store, KnowledgeQuery{Trigger: "explicit-reuse", Kind: "solution", Key: "browser timeout", Scope: "global"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.ExactMatch || result.FullScan || result.CandidateRecords != 1 || len(result.Matches) != 1 || result.Matches[0].ID != record.ID {
		t.Fatalf("unexpected exact query result: %#v", result)
	}
	if result.MaxIndexLookups != knowledgeMaxLookups || result.MaxCandidateRecords != knowledgeMaxCandidates || result.MaxResults != knowledgeMaxResults || result.MaxRefsPerIndex != knowledgeMaxRefsPerIndex {
		t.Fatalf("knowledge query did not expose its hard budgets: %#v", result)
	}
}

func TestKnowledgeIndexedCandidateAndRelationQueries(t *testing.T) {
	store, evidence := knowledgeFixture(t)
	_, err := RecordKnowledge(store, KnowledgeRecordInput{
		Kind: "failure", Key: "browser wait timeout", Scope: "global", Summary: "Use the known wait helper.", RootCause: "The page was queried before the navigation settled.",
		Location: filepath.Join(t.TempDir(), "solution.md"), Verification: evidence, Tags: []string{"browser", "wait"},
		Relations: []KnowledgeRelation{{Predicate: "caused-by", Target: "navigation-race"}},
	})
	// The solution location must exist when it is local.
	if err == nil {
		t.Fatalf("expected missing solution location to be rejected, got %v", err)
	}
	solution := filepath.Join(t.TempDir(), "solution.md")
	if err := os.WriteFile(solution, []byte("verified"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := RecordKnowledge(store, KnowledgeRecordInput{
		Kind: "failure", Key: "browser wait timeout", Scope: "global", Summary: "Use the known wait helper.", RootCause: "The page was queried before the navigation settled.",
		Location: solution, Verification: evidence, Tags: []string{"browser", "wait"}, Relations: []KnowledgeRelation{{Predicate: "caused-by", Target: "navigation-race"}},
	}); err != nil {
		t.Fatal(err)
	}
	result, err := QueryKnowledge(store, KnowledgeQuery{Trigger: "failure", Key: "browser wait", Scope: "global", Tags: []string{"browser"}})
	if err != nil {
		t.Fatal(err)
	}
	if result.FullScan || result.IndexLookups == 0 || len(result.Matches) != 1 {
		t.Fatalf("indexed candidate query did not stay bounded: %#v", result)
	}
	relationResult, err := QueryKnowledge(store, KnowledgeQuery{Trigger: "explicit-reuse", Scope: "global", RelatedTo: "navigation-race", Relation: "caused-by"})
	if err != nil || len(relationResult.Matches) != 1 {
		t.Fatalf("relation query failed: result=%#v err=%v", relationResult, err)
	}
}

func TestKnowledgeRequiresRootCauseAndRejectsSecrets(t *testing.T) {
	store, evidence := knowledgeFixture(t)
	location := filepath.Join(t.TempDir(), "solution.md")
	if err := os.WriteFile(location, []byte("verified"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := RecordKnowledge(store, KnowledgeRecordInput{Kind: "failure", Key: "failure", Scope: "global", Summary: "summary", Location: location, Verification: evidence}); err == nil {
		t.Fatal("failure without root cause was accepted")
	}
	secret := "sk-" + strings.Repeat("a", 20)
	if _, err := RecordKnowledge(store, KnowledgeRecordInput{Kind: "solution", Key: "api", Scope: "global", Summary: "authorization: " + secret, Location: location, Verification: evidence}); err == nil {
		t.Fatal("credential-like content was accepted")
	}
}

func TestKnowledgeUpdateIncrementsRevisionAndTamperingIsRejected(t *testing.T) {
	store, evidence := knowledgeFixture(t)
	location := filepath.Join(t.TempDir(), "solution.md")
	if err := os.WriteFile(location, []byte("verified"), 0o600); err != nil {
		t.Fatal(err)
	}
	first, err := RecordKnowledge(store, KnowledgeRecordInput{Kind: "solution", Key: "cache miss", Scope: "global", Summary: "first", Location: location, Verification: evidence, Tags: []string{"old-tag"}})
	if err != nil {
		t.Fatal(err)
	}
	second, err := RecordKnowledge(store, KnowledgeRecordInput{Kind: "solution", Key: "cache miss", Scope: "global", Summary: "updated", Location: location, Verification: evidence, Tags: []string{"new-tag"}})
	if err != nil || second.Revision != first.Revision+1 {
		t.Fatalf("revision was not advanced: first=%#v second=%#v err=%v", first, second, err)
	}
	oldTagResult, err := QueryKnowledge(store, KnowledgeQuery{Trigger: "explicit-reuse", Scope: "global", Tags: []string{"old-tag"}})
	if err != nil || oldTagResult.CandidateRecords != 0 {
		t.Fatalf("stale tag index remained after update: result=%#v err=%v", oldTagResult, err)
	}
	path := KnowledgeRecordPath(store, second.Kind, second.ID)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.Replace(string(data), `"updated"`, `"tampered"`, 1))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := QueryKnowledge(store, KnowledgeQuery{Trigger: "explicit-reuse", Kind: "solution", Key: "cache miss", Scope: "global"}); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("tampered record was not rejected: %v", err)
	}
}

func TestKnowledgeVerificationEvidenceHashIsRechecked(t *testing.T) {
	store, evidence := knowledgeFixture(t)
	location := filepath.Join(t.TempDir(), "solution.md")
	if err := os.WriteFile(location, []byte("verified"), 0o600); err != nil {
		t.Fatal(err)
	}
	record, err := RecordKnowledge(store, KnowledgeRecordInput{Kind: "solution", Key: "evidence hash", Scope: "global", Summary: "verified", Location: location, Verification: evidence})
	if err != nil {
		t.Fatal(err)
	}
	objectPath, err := knowledgeObjectPath(store, record.Verification)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(objectPath, []byte(`{"passed":false}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := QueryKnowledge(store, KnowledgeQuery{Trigger: "verification-trace", Kind: "solution", Key: "evidence hash", Scope: "global"}); err == nil || !strings.Contains(err.Error(), "evidence hash mismatch") {
		t.Fatalf("modified evidence remained trusted: %v", err)
	}
}

func TestKnowledgeRejectsOversizedInputsBeforeIndexing(t *testing.T) {
	store, evidence := knowledgeFixture(t)
	location := filepath.Join(t.TempDir(), "solution.md")
	if err := os.WriteFile(location, []byte("verified"), 0o600); err != nil {
		t.Fatal(err)
	}
	oversizedKey := strings.Repeat("k", knowledgeMaxKeyBytes+1)
	if _, err := RecordKnowledge(store, KnowledgeRecordInput{
		Kind: "solution", Key: oversizedKey, Scope: "global", Summary: "summary", Location: location, Verification: evidence,
	}); err == nil || !strings.Contains(err.Error(), "key exceeds") {
		t.Fatalf("oversized record key was accepted: %v", err)
	}
	if _, err := QueryKnowledge(store, KnowledgeQuery{Trigger: "explicit-reuse", Key: oversizedKey, Scope: "global"}); err == nil || !strings.Contains(err.Error(), "key exceeds") {
		t.Fatalf("oversized query key was accepted: %v", err)
	}
}

func TestKnowledgeScopeIsStronglyIsolatedAndRequired(t *testing.T) {
	store, evidence := knowledgeFixture(t)
	solution := filepath.Join(t.TempDir(), "solution.md")
	if err := os.WriteFile(solution, []byte("verified"), 0o600); err != nil {
		t.Fatal(err)
	}
	scopeA, err := KnowledgeWorkspaceScope(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	scopeB, err := KnowledgeWorkspaceScope(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct{ scope, summary string }{{scopeA, "workspace A"}, {scopeB, "workspace B"}} {
		if _, err := RecordKnowledge(store, KnowledgeRecordInput{Kind: "solution", Key: "same key", Scope: item.scope, Summary: item.summary, Location: solution, Verification: evidence}); err != nil {
			t.Fatal(err)
		}
	}
	resultA, err := QueryKnowledge(store, KnowledgeQuery{Trigger: "explicit-reuse", Kind: "solution", Key: "same key", Scope: scopeA})
	if err != nil || len(resultA.Matches) != 1 || resultA.Matches[0].Summary != "workspace A" {
		t.Fatalf("workspace A scope leaked or missed: result=%#v err=%v", resultA, err)
	}
	resultB, err := QueryKnowledge(store, KnowledgeQuery{Trigger: "explicit-reuse", Kind: "solution", Key: "same key", Scope: scopeB})
	if err != nil || len(resultB.Matches) != 1 || resultB.Matches[0].Summary != "workspace B" {
		t.Fatalf("workspace B scope leaked or missed: result=%#v err=%v", resultB, err)
	}
	if _, err := QueryKnowledge(store, KnowledgeQuery{Trigger: "explicit-reuse", Kind: "solution", Key: "same key"}); err == nil {
		t.Fatal("scope-free knowledge query was accepted")
	}
}

func TestKnowledgeCapacityEvictsExpiredOldestAndByteOverflow(t *testing.T) {
	oldNow := time.Now().UTC().Add(-knowledgeRetention - 24*time.Hour)
	store, evidence := knowledgeFixtureAt(t, oldNow)
	solution := filepath.Join(t.TempDir(), "old-solution.md")
	if err := os.WriteFile(solution, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := recordKnowledgeAt(store, KnowledgeRecordInput{Kind: "solution", Key: "expired", Scope: "global", Summary: "expired", Location: solution, Verification: evidence}, oldNow); err != nil {
		t.Fatal(err)
	}
	capacity, err := enforceKnowledgeStoreCapacity(store, time.Now().UTC(), knowledgeMaxNodes, knowledgeMaxStoreBytes)
	if err != nil || capacity.NodeCount != 0 || capacity.EvictedRecords != 1 {
		t.Fatalf("expired record was not collected: capacity=%#v err=%v", capacity, err)
	}

	store, evidence = knowledgeFixture(t)
	base := time.Now().UTC()
	for index := 0; index < 3; index++ {
		location := filepath.Join(t.TempDir(), fmt.Sprintf("solution-%d.md", index))
		if err := os.WriteFile(location, []byte(strings.Repeat(fmt.Sprintf("%d", index), 12*1024)), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := recordKnowledgeAt(store, KnowledgeRecordInput{Kind: "solution", Key: fmt.Sprintf("quota-%d", index), Scope: "global", Summary: "quota", Location: location, Verification: evidence}, base.Add(time.Duration(index)*time.Second)); err != nil {
			t.Fatal(err)
		}
	}
	capacity, err = enforceKnowledgeStoreCapacity(store, base.Add(time.Minute), 2, knowledgeMaxStoreBytes)
	if err != nil || capacity.NodeCount != 2 || capacity.EvictedRecords != 1 {
		t.Fatalf("node quota did not evict the oldest record: capacity=%#v err=%v", capacity, err)
	}
	oldest, err := QueryKnowledge(store, KnowledgeQuery{Trigger: "explicit-reuse", Kind: "solution", Key: "quota-0", Scope: "global"})
	if err != nil || len(oldest.Matches) != 0 {
		t.Fatalf("oldest record survived node quota: result=%#v err=%v", oldest, err)
	}
	current, err := knowledgeStoreCapacity(store)
	if err != nil {
		t.Fatal(err)
	}
	capacity, err = enforceKnowledgeStoreCapacity(store, base.Add(2*time.Minute), knowledgeMaxNodes, current.StoreBytes-1)
	if err != nil || capacity.StoreBytes > current.StoreBytes-1 || capacity.EvictedRecords == 0 {
		t.Fatalf("byte quota did not collect records: capacity=%#v err=%v", capacity, err)
	}
}

func TestKnowledgeTransactionRepairRebuildsIndexesAndSweepsObjects(t *testing.T) {
	store, evidence := knowledgeFixture(t)
	solution := filepath.Join(t.TempDir(), "solution.md")
	if err := os.WriteFile(solution, []byte("verified"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := RecordKnowledge(store, KnowledgeRecordInput{Kind: "solution", Key: "repair", Scope: "global", Summary: "repair", Location: solution, Verification: evidence}); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(store, "v1", "indexes")); err != nil {
		t.Fatal(err)
	}
	orphanRef, err := storeKnowledgeObject(store, []byte("orphan"))
	if err != nil {
		t.Fatal(err)
	}
	orphanPath, err := knowledgeObjectPath(store, orphanRef)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeKnowledgeFile(knowledgeTransactionPath(store), []byte("interrupted\n")); err != nil {
		t.Fatal(err)
	}
	result, err := QueryKnowledge(store, KnowledgeQuery{Trigger: "explicit-reuse", Kind: "solution", Key: "repair", Scope: "global"})
	if err != nil || len(result.Matches) != 1 {
		t.Fatalf("interrupted transaction was not repaired: result=%#v err=%v", result, err)
	}
	if _, err := os.Stat(orphanPath); !os.IsNotExist(err) {
		t.Fatalf("orphan object survived transaction repair: %v", err)
	}
	if _, err := os.Stat(knowledgeTransactionPath(store)); !os.IsNotExist(err) {
		t.Fatalf("transaction marker survived repair: %v", err)
	}
}

func TestKnowledgeRejectsSelfAttestedReceiptAndLocalPathRelations(t *testing.T) {
	store := t.TempDir()
	evidence := filepath.Join(t.TempDir(), "fake.json")
	if err := os.WriteFile(evidence, []byte(`{"passed":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	solution := filepath.Join(t.TempDir(), "solution.md")
	if err := os.WriteFile(solution, []byte("verified"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := RecordKnowledge(store, KnowledgeRecordInput{Kind: "solution", Key: "fake", Scope: "global", Summary: "fake", Location: solution, Verification: evidence}); err == nil {
		t.Fatal("self-attested passed JSON was accepted as verification")
	}
	_, validEvidence := knowledgeFixture(t)
	if _, err := RecordKnowledge(store, KnowledgeRecordInput{
		Kind: "solution", Key: "path", Scope: "global", Summary: "path", Location: solution, Verification: validEvidence,
		Relations: []KnowledgeRelation{{Predicate: "related-to", Target: solution}},
	}); err == nil {
		t.Fatal("local absolute path was accepted in a graph relation")
	}
}

func TestKnowledgeConcurrentUpdatesProduceOneRevisionChain(t *testing.T) {
	store, evidence := knowledgeFixture(t)
	solution := filepath.Join(t.TempDir(), "solution.md")
	if err := os.WriteFile(solution, []byte("verified"), 0o600); err != nil {
		t.Fatal(err)
	}
	const updates = 8
	errs := make(chan error, updates)
	var group sync.WaitGroup
	for index := 0; index < updates; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			_, err := RecordKnowledge(store, KnowledgeRecordInput{Kind: "solution", Key: "concurrent", Scope: "global", Summary: fmt.Sprintf("update-%d", index), Location: solution, Verification: evidence})
			errs <- err
		}(index)
	}
	group.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	result, err := QueryKnowledge(store, KnowledgeQuery{Trigger: "explicit-reuse", Kind: "solution", Key: "concurrent", Scope: "global"})
	if err != nil || len(result.Matches) != 1 || result.Matches[0].Revision != updates {
		t.Fatalf("concurrent updates did not serialize revisions: result=%#v err=%v", result, err)
	}
}
