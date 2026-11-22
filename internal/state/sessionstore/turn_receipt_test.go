package sessionstore

import (
	"testing"

	"reasonix/internal/contract/provider"
)

// receipt is a decision receipt with the identity the dedupe keys on.
func receipt(id string) *provider.DecisionReceipt {
	return &provider.DecisionReceipt{ID: id, Kind: "plan", Tool: "exit_plan_mode", Outcome: "start_execution"}
}

// A receipt held against a turn must never land on the reply that came before
// it: in an ongoing chat that reply is the previous turn's, and the plan card is
// answered before this turn's user message exists.
func TestAttachTurnReceiptsNeverCrossesTheFloor(t *testing.T) {
	s := NewSession("sys")
	s.Add(provider.Message{Role: provider.RoleUser, Content: "first ask"})
	s.Add(provider.Message{Role: provider.RoleAssistant, Content: "first answer"})
	// The next turn starts here; its user message has not landed yet.
	floor := s.Len()

	if s.AttachTurnReceipts(floor, []*provider.DecisionReceipt{receipt("r1")}) {
		t.Fatal("attached a receipt with no message in this turn to own it")
	}
	msgs := s.Snapshot()
	if len(msgs[2].DecisionReceipts) != 0 {
		t.Fatalf("previous turn's reply took the receipt: %+v", msgs[2].DecisionReceipts)
	}
	if len(msgs) != floor {
		t.Fatalf("attach appended %d messages; the announced slot must stay empty", len(msgs)-floor)
	}

	// The turn's own message and reply arrive.
	s.Add(provider.Message{Role: provider.RoleUser, Content: "second ask"})
	idx := s.AddIndexed(provider.Message{Role: provider.RoleAssistant, Content: "second answer"})
	if !s.AttachTurnReceipts(floor, []*provider.DecisionReceipt{receipt("r1")}) {
		t.Fatal("attach failed once this turn had a reply")
	}
	msgs = s.Snapshot()
	if len(msgs[2].DecisionReceipts) != 0 {
		t.Fatal("previous turn's reply was modified by a later attach")
	}
	if len(msgs[idx].DecisionReceipts) != 1 || msgs[idx].DecisionReceipts[0].ID != "r1" {
		t.Fatalf("this turn's reply = %+v, want the receipt", msgs[idx].DecisionReceipts)
	}
}

// A retried attach — after a failed save, or with a marker whose pending list
// outlived its transcript — must not duplicate the receipt.
func TestAttachTurnReceiptsDeduplicatesByID(t *testing.T) {
	s := NewSession("sys")
	s.Add(provider.Message{Role: provider.RoleUser, Content: "ask"})
	idx := s.AddIndexed(provider.Message{Role: provider.RoleAssistant, Content: "answer"})

	if !s.AttachTurnReceipts(0, []*provider.DecisionReceipt{receipt("r1")}) {
		t.Fatal("first attach failed")
	}
	if !s.AttachTurnReceipts(0, []*provider.DecisionReceipt{receipt("r1"), receipt("r2")}) {
		t.Fatal("second attach failed")
	}
	got := s.Snapshot()[idx].DecisionReceipts
	if len(got) != 2 || got[0].ID != "r1" || got[1].ID != "r2" {
		t.Fatalf("receipts = %+v, want r1 then r2 exactly once each", got)
	}
}

// Holding a receipt writes it to the turn's marker and to nothing else, so an
// early save leaves the reserved slot empty and the transcript untouched.
func TestHoldTurnReceiptKeepsTheAnnouncedSlotEmpty(t *testing.T) {
	d := newDurabilityRun(t)
	run, _ := d.buildSaved(1)
	path := d.path

	// Turn 2 begins; its user message has not landed.
	floor := run.Len()
	marker, err := BeginSessionInFlightTurn(path, floor, false)
	if err != nil {
		t.Fatalf("begin turn: %v", err)
	}
	held, err := HoldTurnReceipt(path, marker.ID, receipt("r1"))
	if err != nil || !held {
		t.Fatalf("hold = %v err = %v, want the marker to carry it", held, err)
	}

	// An early save must not move the transcript.
	if err := run.SaveSnapshot(path); err != nil {
		t.Fatalf("early save: %v", err)
	}
	reloaded, err := LoadSession(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(reloaded.Messages) != floor {
		t.Fatalf("early save wrote %d messages, want the transcript unchanged at %d", len(reloaded.Messages), floor)
	}
	if pending := PendingTurnReceipts(path, marker.ID); len(pending) != 1 || pending[0].ID != "r1" {
		t.Fatalf("pending receipts = %+v, want r1 on the turn", pending)
	}

	// A stale approval must not write to a newer turn's marker.
	fresh, err := BeginSessionInFlightTurn(path, floor+2, false)
	if err != nil {
		t.Fatalf("begin fresh turn: %v", err)
	}
	if stale, err := HoldTurnReceipt(path, marker.ID, receipt("r2")); err != nil || stale {
		t.Fatalf("stale hold = %v err = %v, want no marker matched", stale, err)
	}
	if pending := PendingTurnReceipts(path, fresh.ID); len(pending) != 0 {
		t.Fatalf("fresh turn inherited %+v", pending)
	}
}

// Clearing the pending copy is what makes the receipt exactly-once after a
// commit, and it must be callable only for the matching marker.
func TestClearPendingReceiptsKeepsTheMarkerUntilItMatches(t *testing.T) {
	d := newDurabilityRun(t)
	_, _ = d.buildSaved(1)
	marker, err := BeginSessionInFlightTurn(d.path, 3, false)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := HoldTurnReceipt(d.path, marker.ID, receipt("r1")); err != nil {
		t.Fatalf("hold: %v", err)
	}
	if cleared, err := ClearPendingReceiptsIfDurable(d.path, "not-this-turn"); err != nil || cleared {
		t.Fatalf("mismatched clear = %v err = %v, want no change", cleared, err)
	}
	if pending := PendingTurnReceipts(d.path, marker.ID); len(pending) != 1 {
		t.Fatalf("pending after mismatched clear = %+v, want r1 kept", pending)
	}
	if cleared, err := ClearPendingReceiptsIfDurable(d.path, marker.ID); err != nil || !cleared {
		t.Fatalf("clear = %v err = %v, want the pending copy dropped", cleared, err)
	}
	if pending := PendingTurnReceipts(d.path, marker.ID); len(pending) != 0 {
		t.Fatalf("pending after clear = %+v, want none", pending)
	}
}
