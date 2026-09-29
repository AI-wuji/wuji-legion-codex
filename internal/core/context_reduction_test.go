package core

import "testing"

func TestReduceContextArchivesSupersededToolPairAndKeepsMessages(t *testing.T) {
	got, err := ReduceContext(ContextReductionInput{
		Records: []ContextReductionRecord{
			{ID: "u1", Role: "user", Kind: "message", Text: "inspect the file"},
			{ID: "c1", Role: "assistant", Kind: "tool_call", Tool: "Read", Operation: "view", Footprint: "src/a.go", PairID: "p1"},
			{ID: "r1", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "src/a.go", PairID: "p1", Text: "same", Succeeded: true},
			{ID: "c2", Role: "assistant", Kind: "tool_call", Tool: "Read", Operation: "view", Footprint: "src/a.go", PairID: "p2"},
			{ID: "r2", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "src/a.go", PairID: "p2", Text: "same", Succeeded: true},
			{ID: "a1", Role: "assistant", Kind: "message", Text: "done"},
		},
		ArchiveDuplicates: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Visible) != 4 || len(got.Archived) != 2 {
		t.Fatalf("unexpected projection: visible=%#v archived=%#v", got.Visible, got.Archived)
	}
	if got.OriginalRecordBytes <= got.VisibleRecordBytes || got.SavedRecordBytes <= 0 {
		t.Fatalf("projection did not report measurable savings: %#v", got)
	}
	if got.Archived[0].ID != "c1" || got.Archived[1].ID != "r1" {
		t.Fatalf("old tool pair was not archived in order: %#v", got.Archived)
	}
	restored, err := RestoreContext(got)
	if err != nil || len(restored) != 6 || restored[2].Text != "same" {
		t.Fatalf("projection could not be restored: restored=%#v err=%v", restored, err)
	}
	got.Archived[1].Text = "tampered"
	if _, err := RestoreContext(got); err == nil {
		t.Fatal("archive tampering escaped the integrity check")
	}
}

func TestRecallArchivedContextReturnsOriginalEvidenceInRankOrder(t *testing.T) {
	projection, err := ReduceContext(ContextReductionInput{
		Records: []ContextReductionRecord{
			{ID: "old-a", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "src/a.go", Text: "alpha", Succeeded: true},
			{ID: "new-a", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "src/a.go", Text: "alpha", Succeeded: true},
			{ID: "old-b", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "src/b.go", Text: "beta", Succeeded: true},
			{ID: "new-b", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "src/b.go", Text: "beta", Succeeded: true},
		},
		ArchiveDuplicates: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := RecallArchivedContext(projection, "src/a.go", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Matches) != 1 || got.Matches[0].Record.ID != "old-a" || got.Matches[0].Record.Text != "alpha" {
		t.Fatalf("recall did not return the original archived record: %#v", got)
	}
	if got.Examined != 2 || !got.ArchivedOnly {
		t.Fatalf("recall metadata is incorrect: %#v", got)
	}
}

func TestRecallArchivedContextRejectsTamperedProjection(t *testing.T) {
	projection, err := ReduceContext(ContextReductionInput{
		Records: []ContextReductionRecord{
			{ID: "old", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "a", Text: "same", Succeeded: true},
			{ID: "new", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "a", Text: "same", Succeeded: true},
		},
		ArchiveDuplicates: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	projection.Archived[0].Text = "tampered"
	if _, err := RecallArchivedContext(projection, "a", 1); err == nil {
		t.Fatal("recall accepted a tampered projection")
	}
}

func TestRecallArchivedContextKeepsToolPairsTogether(t *testing.T) {
	projection, err := ReduceContext(ContextReductionInput{
		Records: []ContextReductionRecord{
			{ID: "old-call", Role: "assistant", Kind: "tool_call", Tool: "Read", Operation: "view", PairID: "p1"},
			{ID: "old-result", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "src/a.go", PairID: "p1", Text: "same", Succeeded: true},
			{ID: "new-call", Role: "assistant", Kind: "tool_call", Tool: "Read", Operation: "view", PairID: "p2"},
			{ID: "new-result", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "src/a.go", PairID: "p2", Text: "same", Succeeded: true},
		},
		ArchiveDuplicates: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := RecallArchivedContext(projection, "src/a.go", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Matches) != 2 || got.Matches[0].Record.ID != "old-call" || got.Matches[1].Record.ID != "old-result" {
		t.Fatalf("recall did not preserve the archived tool pair: %#v", got)
	}
}

func TestRestoreContextRejectsReclassifyingProtectedMessage(t *testing.T) {
	got, err := ReduceContext(ContextReductionInput{Records: []ContextReductionRecord{
		{ID: "u1", Role: "user", Kind: "message", Text: "never delete"},
		{ID: "r1", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "a", Text: "same", Succeeded: true},
		{ID: "r2", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "a", Text: "same", Succeeded: true},
	}, ArchiveDuplicates: true})
	if err != nil {
		t.Fatal(err)
	}
	got.Archived = append([]ContextReductionRecord{got.Visible[0]}, got.Archived...)
	got.Visible = got.Visible[1:]
	got.ArchivedCount++
	got.Decisions[0].Action = "archive"
	if _, err := RestoreContext(got); err == nil {
		t.Fatal("forged projection archived a protected user message")
	}
}

func TestReduceContextArchivesStaleReadAfterWrite(t *testing.T) {
	got, err := ReduceContext(ContextReductionInput{
		Records: []ContextReductionRecord{
			{ID: "r1", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "src/a.go", PairID: "p1", Text: "before", Succeeded: true},
			{ID: "w1", Role: "tool", Kind: "tool_result", Tool: "Write", Operation: "write", Footprint: "src/a.go", PairID: "p2", Text: "changed", Succeeded: true},
			{ID: "r2", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "src/a.go", PairID: "p3", Text: "after", Succeeded: true, Protected: true},
		},
		ArchiveStale: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Archived) != 1 || got.Archived[0].ID != "r1" {
		t.Fatalf("stale read was not archived: %#v", got.Archived)
	}
	if len(got.Visible) != 2 || got.Visible[0].ID != "w1" || got.Visible[1].ID != "r2" {
		t.Fatalf("write or protected read was lost: %#v", got.Visible)
	}
}

func TestReduceContextPreservesRecentAndCriticalRecords(t *testing.T) {
	got, err := ReduceContext(ContextReductionInput{
		Records: []ContextReductionRecord{
			{ID: "u1", Role: "user", Kind: "message", Text: "constraint"},
			{ID: "r1", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "a", PairID: "p1", Succeeded: true},
			{ID: "r2", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "b", PairID: "p2", Critical: true, Succeeded: true},
			{ID: "r3", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "c", PairID: "p3", Succeeded: true},
		},
		PreserveRecent:    1,
		ArchiveDuplicates: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Visible) != 4 {
		t.Fatalf("unrelated observations were archived: %#v", got)
	}
	ids := map[string]bool{}
	for _, record := range got.Visible {
		ids[record.ID] = true
	}
	if !ids["u1"] || !ids["r2"] || !ids["r3"] {
		t.Fatalf("protected or recent records were removed: %#v", got.Visible)
	}
}

func TestReduceContextRejectsDuplicateAndUnknownOperation(t *testing.T) {
	if _, err := ReduceContext(ContextReductionInput{Records: []ContextReductionRecord{
		{ID: "same", Role: "user", Kind: "message"},
		{ID: "same", Role: "tool", Kind: "tool_result"},
	}}); err == nil {
		t.Fatal("duplicate record id was accepted")
	}
	if _, err := ReduceContext(ContextReductionInput{Records: []ContextReductionRecord{
		{ID: "r1", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "delete"},
	}}); err == nil {
		t.Fatal("unknown context operation was accepted")
	}
}

func TestReduceContextKeepsFailureAndProtectedPair(t *testing.T) {
	got, err := ReduceContext(ContextReductionInput{
		Records: []ContextReductionRecord{
			{ID: "c1", Role: "assistant", Kind: "tool_call", Tool: "Read", Operation: "view", Footprint: "a", PairID: "p1"},
			{ID: "r1", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "a", PairID: "p1", Text: "same", IsError: true},
			{ID: "c2", Role: "assistant", Kind: "tool_call", Tool: "Read", Operation: "view", Footprint: "a", PairID: "p2"},
			{ID: "r2", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "a", PairID: "p2", Text: "same", Succeeded: true},
		}, ArchiveDuplicates: true,
	})
	if err != nil || got.ArchivedCount != 0 {
		t.Fatalf("failed result or its call was archived: result=%#v err=%v", got, err)
	}
}

func TestReduceContextNeedsIdenticalContentAndSuccessfulWitness(t *testing.T) {
	got, err := ReduceContext(ContextReductionInput{Records: []ContextReductionRecord{
		{ID: "old", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "a", Text: "before", Succeeded: true},
		{ID: "new", Role: "tool", Kind: "tool_result", Tool: "Read", Operation: "view", Footprint: "a", Text: "after", Succeeded: true},
		{ID: "failed-write", Role: "tool", Kind: "tool_result", Tool: "Write", Operation: "write", Footprint: "a", IsError: true},
	}, ArchiveDuplicates: true, ArchiveStale: true})
	if err != nil || got.ArchivedCount != 0 {
		t.Fatalf("changed view or failed write archived evidence: result=%#v err=%v", got, err)
	}
}
