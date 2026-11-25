package control

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"reasonix/internal/base/testenv"
	"reasonix/internal/contract/event"
	"reasonix/internal/contract/provider"
	"reasonix/internal/contract/tool"
	"reasonix/internal/runtime/agent"
	"reasonix/internal/runtime/coordinator"
	"reasonix/internal/state/sessionstore"
)

func TestPlanReceiptWithoutDiskMarker(t *testing.T) {
	t.Setenv("REASONIX_HOME", testenv.TempDir(t))
	const request = "implement the next feature"
	sess := sessionstore.NewSession("executor system")
	sess.Add(provider.Message{Role: provider.RoleUser, Content: "first request"})
	previous := sess.AddIndexed(provider.Message{Role: provider.RoleAssistant, Content: "first reply"})
	floor := sess.Len()
	exec := agent.New(&scriptedTurns{turns: [][]provider.Chunk{textTurn("new reply")}}, tool.NewRegistry(), sess, agent.Options{}, event.Discard)
	planner := &recordingProvider{name: "planner", streams: [][]provider.Chunk{
		toolCallTurn("submit-1", "submit_plan", `{"objective":"implement the next feature","steps":[{"title":"Implement the feature"}]}`),
		textTurn("plan submitted"),
	}}
	coord := coordinator.NewCoordinatorWithPlannerPolicy(planner, sessionstore.NewSession("planner system"), nil, agent.PlannerToolRegistry(tool.NewRegistry(), nil, nil), agent.Options{}, exec, 0, event.Discard, NewPlannerPolicy(0))
	cards := make(chan string, 1)
	started := make(chan int, 1)
	mismatches := make(chan string, 1)
	done := make(chan event.Event, 1)
	c := New(Options{Runner: coord, Executor: exec, Sink: event.FuncSink(func(e event.Event) {
		switch {
		case e.Kind == event.TurnStarted && e.MsgIndex != nil:
			started <- *e.MsgIndex
		case e.Kind == event.ApprovalRequest:
			cards <- e.Approval.ID
		case e.Code == event.NoticeCodeTurnIdentityMismatch:
			mismatches <- e.Detail
		case e.Kind == event.TurnDone:
			done <- e
		}
	})})
	t.Cleanup(c.Close)
	c.EnableInteractiveApproval()
	c.SubmitHTTPOptions(request, SubmitOptions{WorkMode: WorkModeAgentPlanner})
	var id string
	select {
	case id = <-cards:
	case e := <-done:
		t.Fatalf("turn finished before approval: %v", e.Err)
	case <-time.After(5 * time.Second):
		t.Fatal("approval card did not open")
	}
	select {
	case index := <-started:
		if index != floor {
			t.Fatalf("announced index = %d, want %d", index, floor)
		}
	default:
		t.Fatal("the controller did not announce the user message")
	}
	if c.SessionPath() != "" || sess.Len() != floor {
		t.Fatal("expected approval in memory before the user message lands")
	}
	if err := c.ResolvePlanDecision(id, PlanDecisionStartExecution); err != nil {
		t.Fatal(err)
	}
	select {
	case e := <-done:
		if e.Err != nil {
			t.Fatal(e.Err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the approved turn did not finish")
	}
	select {
	case detail := <-mismatches:
		t.Fatalf("turn identity mismatch: %s", detail)
	default:
	}
	msgs := sess.Snapshot()
	if len(msgs) <= floor || msgs[floor].Role != provider.RoleUser || sessionstore.UserMessageText(msgs[floor]) != request {
		t.Fatalf("user message did not land at index %d: %+v", floor, msgs)
	}
	if got := msgs[previous].DecisionReceipts; len(got) != 0 {
		t.Fatalf("previous reply carries %d receipts, want none", len(got))
	}
	var count int
	for _, m := range msgs[floor+1:] {
		for _, receipt := range m.DecisionReceipts {
			if receipt.ID != id || m.Role != provider.RoleAssistant || m.Content != "new reply" {
				t.Fatalf("unexpected receipt on %q: %+v", m.Content, receipt)
			}
			count++
		}
	}
	if count != 1 {
		t.Fatalf("new reply carries %d receipts, want one", count)
	}
}

func TestPlanReceiptSurvivesInterruptedRecovery(t *testing.T) {
	t.Run("without_commit_digest", func(t *testing.T) {
		checkPlanReceiptRecovery(t, false, false)
	})
	t.Run("matching_commit_digest", func(t *testing.T) {
		checkPlanReceiptRecovery(t, true, false)
	})
	t.Run("normal_finalization_without_reply", func(t *testing.T) {
		checkPlanReceiptRecovery(t, false, true)
	})
}

func checkPlanReceiptRecovery(t *testing.T, prepareCommit, finishTurn bool) {
	t.Helper()
	t.Setenv("REASONIX_HOME", testenv.TempDir(t))
	c, path := newReceiptController(t)
	sess := c.executor.Session()
	sess.Add(provider.Message{Role: provider.RoleUser, Content: "previous request"})
	previous := sess.AddIndexed(provider.Message{Role: provider.RoleAssistant, Content: "previous reply"})
	if err := c.snapshot(false, false, false); err != nil {
		t.Fatal(err)
	}
	floor := sess.Len()
	_, marker := c.beginTurn(t.Context(), floor, true)
	if marker.ID == "" {
		t.Fatal("missing disk marker")
	}
	receipt := &provider.DecisionReceipt{ID: "approved-plan", Kind: "plan", Tool: planApprovalTool, Outcome: string(PlanDecisionStartExecution)}
	if held, err := sessionstore.HoldTurnReceipt(path, marker.ID, receipt); err != nil || !held {
		t.Fatalf("hold receipt: held=%v, err=%v", held, err)
	}
	const request = "next request"
	sess.Add(provider.Message{Role: provider.RoleUser, Content: request})
	if finishTurn {
		c.finishInFlightTurn(floor, marker)
		meta, ok, err := sessionstore.LoadBranchMeta(path)
		if err != nil || !ok || meta.InFlightTurn != nil {
			t.Fatalf("finalization did not clear the marker: ok=%v, err=%v", ok, err)
		}
	} else {
		if prepareCommit {
			digest, err := sess.ContentDigest()
			if err != nil {
				t.Fatal(err)
			}
			if _, matched, err := sessionstore.PrepareSessionInFlightTurnCommit(path, marker.InFlightTurnMeta, digest); err != nil || !matched {
				t.Fatalf("prepare commit: matched=%v, err=%v", matched, err)
			}
		}
		if err := c.snapshot(false, false, false); err != nil {
			t.Fatal(err)
		}
	}
	for reopen := 1; reopen <= 2; reopen++ {
		loaded, err := sessionstore.LoadSession(path)
		if err != nil {
			t.Fatal(err)
		}
		exec := agent.New(nil, nil, sessionstore.NewSession("sys"), agent.Options{}, event.Discard)
		resumed := New(Options{Executor: exec, SessionDir: filepath.Dir(path), SessionPath: path})
		t.Cleanup(resumed.Close)
		if err := resumed.Resume(loaded, path); err != nil {
			t.Fatal(err)
		}
		msgs := exec.Session().Snapshot()
		if len(msgs) <= floor || msgs[floor].Role != provider.RoleUser || msgs[floor].Content != request {
			t.Fatalf("reopen %d lost the request at index %d: %+v", reopen, floor, msgs)
		}
		if len(msgs[previous].DecisionReceipts) != 0 {
			t.Fatalf("reopen %d attached a receipt to the previous reply", reopen)
		}
		var count int
		for _, m := range msgs[floor:] {
			for _, got := range m.DecisionReceipts {
				if got.ID == receipt.ID {
					count++
				}
			}
		}
		if count != 1 {
			t.Errorf("reopen %d has %d copies of the approval, want one", reopen, count)
		}
		for _, m := range provider.ModelMessages(msgs) {
			if len(m.DecisionReceipts) != 0 || m.DecisionReceipt != nil {
				t.Fatalf("reopen %d exposed approval metadata to the model", reopen)
			}
		}
		meta, ok, err := sessionstore.LoadBranchMeta(path)
		if err != nil || !ok || meta.InFlightTurn != nil {
			t.Fatalf("reopen %d did not clear the recovered marker: ok=%v, err=%v", reopen, ok, err)
		}
	}
}

// newReceiptController is a controller with a real session path and an executor,
// so receipt ownership can be exercised through the save and reload boundary.
func newReceiptController(t *testing.T) (*Controller, string) {
	t.Helper()
	dir := testenv.TempDir(t)
	path := filepath.Join(dir, "session.jsonl")
	prov := &scriptedTurns{turns: [][]provider.Chunk{textTurn("answer")}}
	ag := agent.New(prov, tool.NewRegistry(), sessionstore.NewSession("sys"), agent.Options{}, event.Discard)
	c := New(Options{Runner: ag, Executor: ag, Sink: event.Discard})
	t.Cleanup(c.Close)
	c.SetFreshSessionPath(path)
	return c, path
}

// A plan decision recorded before the turn's user message lands must not touch
// the previous turn's reply, must leave the announced slot empty, and must
// survive a save/load onto the turn that actually owns it.
func TestPlanReceiptWaitsForItsOwnTurnAcrossAnEarlySave(t *testing.T) {
	c, path := newReceiptController(t)
	sess := c.executor.Session()
	sess.Add(provider.Message{Role: provider.RoleUser, Content: "first ask"})
	sess.Add(provider.Message{Role: provider.RoleAssistant, Content: "first answer"})

	floor := sess.Len()
	marker, err := sessionstore.BeginSessionInFlightTurn(path, floor, false)
	if err != nil {
		t.Fatalf("begin turn: %v", err)
	}
	c.holdTurnReceipt(context.Background(), turnStamp{id: marker.ID, start: floor},
		&provider.DecisionReceipt{ID: "plan-1", Kind: "plan", Tool: planApprovalTool, Outcome: string(PlanDecisionStartExecution)})

	// An early save must not write the receipt into the reserved slot.
	if err := sess.SaveSnapshot(path); err != nil {
		t.Fatalf("early save: %v", err)
	}
	early, err := sessionstore.LoadSession(path)
	if err != nil {
		t.Fatalf("reload early: %v", err)
	}
	if len(early.Messages) != floor {
		t.Fatalf("early save wrote %d messages, want %d (the announced slot stays empty)", len(early.Messages), floor)
	}
	if got := early.Messages[2].DecisionReceipts; len(got) != 0 {
		t.Fatalf("previous reply took the receipt: %+v", got)
	}

	// The turn's own user message and reply arrive.
	sess.Add(provider.Message{Role: provider.RoleUser, Content: "second ask"})
	idx := sess.AddIndexed(provider.Message{Role: provider.RoleAssistant, Content: "second answer"})
	if !c.settleTurnReceipts(marker) {
		t.Fatal("settle could not attach the receipt to its own turn")
	}
	if err := sess.SaveSnapshot(path); err != nil {
		t.Fatalf("final save: %v", err)
	}
	if _, err := sessionstore.ClearPendingReceiptsIfDurable(path, marker.ID); err != nil {
		t.Fatalf("clear pending: %v", err)
	}

	reloaded, err := sessionstore.LoadSession(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got := reloaded.Messages[2].DecisionReceipts; len(got) != 0 {
		t.Fatalf("previous reply carries %+v after reload", got)
	}
	got := reloaded.Messages[idx].DecisionReceipts
	if len(got) != 1 || got[0].ID != "plan-1" {
		t.Fatalf("owning reply = %+v, want exactly one plan-1", got)
	}
}

// The transcript saved and the process died before marker cleanup: reopening
// must show exactly one receipt, not zero and not two.
func TestPlanReceiptSurvivesASaveWithoutMarkerCleanup(t *testing.T) {
	c, path := newReceiptController(t)
	sess := c.executor.Session()
	floor := sess.Len()
	marker, err := sessionstore.BeginSessionInFlightTurn(path, floor, false)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	c.holdTurnReceipt(context.Background(), turnStamp{id: marker.ID, start: floor},
		&provider.DecisionReceipt{ID: "plan-1", Kind: "plan", Tool: planApprovalTool, Outcome: string(PlanDecisionStartExecution)})

	idx := sess.AddIndexed(provider.Message{Role: provider.RoleAssistant, Content: "planned"})
	if !c.settleTurnReceipts(marker) {
		t.Fatal("settle could not attach")
	}
	digest, err := sess.ContentDigest()
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	prepared, matched, err := sessionstore.PrepareSessionInFlightTurnCommit(path, marker, digest)
	if err != nil || !matched {
		t.Fatalf("prepare commit = %v matched=%v err=%v", prepared, matched, err)
	}
	if err := sess.SaveSnapshot(path); err != nil {
		t.Fatalf("save: %v", err)
	}
	// Crash here: the marker still exists and still lists the receipt.

	// Recovery attaches idempotently before clearing.
	c.settleTurnReceipts(prepared)
	if _, err := sessionstore.ClearSessionInFlightTurnIfMatch(path, prepared); err != nil {
		t.Fatalf("clear: %v", err)
	}

	reloaded, err := sessionstore.LoadSession(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	var ids []string
	for _, m := range reloaded.Messages {
		for _, r := range m.DecisionReceipts {
			ids = append(ids, r.ID)
		}
		if r := m.DecisionReceipt; r != nil {
			ids = append(ids, r.ID)
		}
	}
	if len(ids) != 1 || ids[0] != "plan-1" {
		t.Fatalf("receipts after recovery = %v, want exactly one plan-1", ids)
	}
	if got := reloaded.Messages[idx].DecisionReceipts; len(got) != 1 {
		t.Fatalf("owning reply = %+v, want the receipt", got)
	}
}

// A rejection is the ordinary path: the host saves the user message and the
// plan reply itself, so the receipt lands on that saved reply and survives.
func TestPlanRejectionReceiptSurvivesReload(t *testing.T) {
	c, path := newReceiptController(t)
	sess := c.executor.Session()
	floor := sess.Len()
	marker, err := sessionstore.BeginSessionInFlightTurn(path, floor, false)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	c.holdTurnReceipt(context.Background(), turnStamp{id: marker.ID, start: floor},
		&provider.DecisionReceipt{ID: "plan-1", Kind: "plan", Tool: planApprovalTool, Outcome: string(PlanDecisionRevisePlan)})

	// persistExecutorNoOp's shape: the turn's user message, then the plan reply.
	sess.Add(provider.Message{Role: provider.RoleUser, Content: "implement it"})
	idx := sess.AddIndexed(provider.Message{Role: provider.RoleAssistant, Content: "plan\n\n(The user did not approve this plan; execution was not started.)"})
	if !c.settleTurnReceipts(marker) {
		t.Fatal("settle could not attach to the saved plan reply")
	}
	if err := sess.SaveSnapshot(path); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := sessionstore.ClearPendingReceiptsIfDurable(path, marker.ID); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if _, err := sessionstore.ClearSessionInFlightTurnIfMatch(path, marker); err != nil {
		t.Fatalf("clear marker: %v", err)
	}

	reloaded, err := sessionstore.LoadSession(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	got := reloaded.Messages[idx].DecisionReceipts
	if len(got) != 1 || got[0].ID != "plan-1" || got[0].Outcome != string(PlanDecisionRevisePlan) {
		t.Fatalf("rejected plan reply = %+v, want the revise receipt", got)
	}
	if reloaded.Messages[idx].Role != provider.RoleAssistant {
		t.Fatalf("carrying message is %q, want the plan reply", reloaded.Messages[idx].Role)
	}
}

// A failed final save must keep the receipt pending so the next attempt still
// attaches it to the same turn — never to a later one.
func TestPlanReceiptSurvivesAFailedSave(t *testing.T) {
	c, path := newReceiptController(t)
	sess := c.executor.Session()
	floor := sess.Len()
	marker, err := sessionstore.BeginSessionInFlightTurn(path, floor, false)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	c.holdTurnReceipt(context.Background(), turnStamp{id: marker.ID, start: floor},
		&provider.DecisionReceipt{ID: "plan-1", Kind: "plan", Tool: planApprovalTool, Outcome: string(PlanDecisionRevisePlan)})

	// No carrier yet: settle keeps it pending rather than forcing it into a slot.
	if c.settleTurnReceipts(marker) {
		t.Fatal("settle claimed the receipt was owned with no message to hold it")
	}
	pending := sessionstore.PendingTurnReceipts(path, marker.ID)
	if len(pending) != 1 || pending[0].ID != "plan-1" {
		t.Fatalf("pending = %+v, want plan-1 still on the turn", pending)
	}
	if got := sessionstore.PendingTurnReceipts(path, "some-later-turn"); len(got) != 0 {
		t.Fatalf("a later turn inherited %+v", got)
	}

	// The reply arrives; the retry attaches it exactly once.
	idx := sess.AddIndexed(provider.Message{Role: provider.RoleAssistant, Content: "revised"})
	if !c.settleTurnReceipts(marker) {
		t.Fatal("retry could not attach the receipt")
	}
	if got := sess.Snapshot()[idx].DecisionReceipts; len(got) != 1 || got[0].ID != "plan-1" {
		t.Fatalf("retry attached %+v, want exactly one plan-1", got)
	}
}
