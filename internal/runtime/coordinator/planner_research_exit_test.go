package coordinator

import (
	"context"
	"strings"
	"testing"

	"reasonix/internal/contract/event"
	"reasonix/internal/contract/provider"
	"reasonix/internal/contract/tool"
	"reasonix/internal/runtime/agent"
	"reasonix/internal/state/sessionstore"
)

// researchExitStreams scripts two research rounds and then whatever the grace
// round answers with, so the budget lands exactly on the final round.
func researchExitStreams(grace []provider.Chunk) [][]provider.Chunk {
	read := func(id, path string) []provider.Chunk {
		return []provider.Chunk{
			{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: id, Name: "read_file", Arguments: `{"path":"` + path + `"}`}},
			{Type: provider.ChunkDone},
		}
	}
	return [][]provider.Chunk{read("read-1", "a"), read("read-2", "b"), grace}
}

func researchExitCoordinator(t *testing.T, planner, exec *mockProvider, sink event.Sink, limit int) *Coordinator {
	t.Helper()
	parentReg := tool.NewRegistry()
	parentReg.Add(coordinatorTestTool{name: "read_file", readOnly: true, output: "package main"})
	policy := func(context.Context, string) agent.PlannerDecision {
		return agent.PlannerDecision{
			Route: agent.PlannerRoutePlanAndExecute, Depth: agent.PlannerDepthFull,
			Reason: "bounded_work", MaxResearchRounds: limit,
		}
	}
	executor := agent.New(exec, tool.NewRegistry(), sessionstore.NewSession("exec-sys"), agent.Options{}, sink)
	return NewCoordinatorWithPlannerPolicy(
		planner, sessionstore.NewSession("planner-sys"), nil, agent.PlannerToolRegistry(parentReg, nil, nil),
		agent.Options{MaxSteps: 0}, executor, 0, sink, policy,
	)
}

// TestCoordinatorPlannerSubmitsPlanOnFinalResearchRound is the screenshot's
// case: the planner spends its research budget and submits the plan on the
// final round. The host must take that plan instead of closing the turn with a
// finalization nudge and handing a planless task to the executor.
func TestCoordinatorPlannerSubmitsPlanOnFinalResearchRound(t *testing.T) {
	planner := &mockProvider{name: "planner", streams: researchExitStreams([]provider.Chunk{
		{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "plan-1", Name: "submit_plan", Arguments: `{"objective":"apply the narrow fix","steps":[{"title":"edit the file"}]}`}},
		{Type: provider.ChunkDone},
	})}
	exec := &mockProvider{name: "executor", chunks: []provider.Chunk{
		{Type: provider.ChunkText, Text: "Done."},
		{Type: provider.ChunkDone},
	}}
	sink := &recordSink{}
	coord := researchExitCoordinator(t, planner, exec, sink, 2)

	if err := coord.Run(context.Background(), "make the bounded change"); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := len(planner.requests); got != 3 {
		t.Fatalf("planner requests = %d, want two research rounds and the final round", got)
	}
	if got := lastUser(exec.requests[0]); !strings.Contains(got, "apply the narrow fix") {
		t.Fatalf("executor input = %q, want the submitted plan", got)
	}
	for _, n := range sink.kinds(event.Notice) {
		if n.Text == plannerResearchFallbackNotice {
			t.Fatalf("notices = %+v, want no research-limit fallback after a delivered plan", sink.kinds(event.Notice))
		}
	}
}

// TestCoordinatorPlannerRefusesResearchInFinalBatch keeps the boundary: the
// same round may deliver the plan, but the research call beside it must not run.
func TestCoordinatorPlannerRefusesResearchInFinalBatch(t *testing.T) {
	planner := &mockProvider{name: "planner", streams: researchExitStreams([]provider.Chunk{
		{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "read-3", Name: "read_file", Arguments: `{"path":"c"}`}},
		{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "plan-1", Name: "submit_plan", Arguments: `{"objective":"apply the narrow fix","steps":[{"title":"edit the file"}]}`}},
		{Type: provider.ChunkDone},
	})}
	exec := &mockProvider{name: "executor", chunks: []provider.Chunk{
		{Type: provider.ChunkText, Text: "Done."},
		{Type: provider.ChunkDone},
	}}
	sink := &recordSink{}
	coord := researchExitCoordinator(t, planner, exec, sink, 2)

	if err := coord.Run(context.Background(), "make the bounded change"); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// The refused read is reported to the planner, and the plan beside it lands.
	var refused bool
	for _, msg := range coord.plannerSess.Snapshot() {
		if msg.Role == provider.RoleTool && msg.ToolCallID == "read-3" {
			refused = true
			if !strings.Contains(msg.Content, "submit_plan") {
				t.Fatalf("refused research result = %q, want it to name the exit", msg.Content)
			}
		}
		if msg.Role == provider.RoleTool && msg.ToolCallID == "plan-1" && strings.Contains(msg.Content, "blocked") {
			t.Fatalf("submit_plan was refused in the final round: %q", msg.Content)
		}
	}
	if !refused {
		t.Fatal("the research call in the final batch was not refused; the boundary leaked")
	}
	if got := lastUser(exec.requests[0]); !strings.Contains(got, "apply the narrow fix") {
		t.Fatalf("executor input = %q, want the submitted plan", got)
	}
}

// TestCoordinatorPlannerNoChangesOnFinalResearchRound covers the other exit.
func TestCoordinatorPlannerNoChangesOnFinalResearchRound(t *testing.T) {
	planner := &mockProvider{name: "planner", streams: researchExitStreams([]provider.Chunk{
		{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "none-1", Name: "conclude_no_changes", Arguments: `{"reason":"the flag is already set"}`}},
		{Type: provider.ChunkDone},
	})}
	exec := &mockProvider{name: "executor"}
	sink := &recordSink{}
	coord := researchExitCoordinator(t, planner, exec, sink, 2)

	if err := coord.Run(context.Background(), "check the flag"); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := len(exec.requests); got != 0 {
		t.Fatalf("executor requests = %d, want none for a no-change conclusion", got)
	}
	var relayed string
	for _, e := range sink.kinds(event.Text) {
		if strings.Contains(e.Text, "the flag is already set") {
			relayed = e.Text
		}
	}
	if relayed == "" {
		t.Fatalf("texts = %+v, want the conclusion relayed to the user", sink.kinds(event.Text))
	}
}

// TestCoordinatorPlannerFinalRoundProseStillHandsOff keeps the graceful half
// for plan-and-execute: a planner that answers in prose instead of taking an
// exit is still advisory input, so the executor runs as it always did.
func TestCoordinatorPlannerFinalRoundProseStillHandsOff(t *testing.T) {
	planner := &mockProvider{name: "planner", streams: researchExitStreams([]provider.Chunk{
		{Type: provider.ChunkText, Text: "1. edit the narrow path\n2. run the focused test"},
		{Type: provider.ChunkDone},
	})}
	exec := &mockProvider{name: "executor", chunks: []provider.Chunk{
		{Type: provider.ChunkText, Text: "Done."},
		{Type: provider.ChunkDone},
	}}
	sink := &recordSink{}
	coord := researchExitCoordinator(t, planner, exec, sink, 2)

	if err := coord.Run(context.Background(), "make the bounded change"); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := len(exec.requests); got != 1 {
		t.Fatalf("executor requests = %d, want one handoff", got)
	}
	if got := lastUser(exec.requests[0]); !strings.Contains(got, "edit the narrow path") {
		t.Fatalf("executor input = %q, want the planner's prose", got)
	}
	for _, n := range sink.kinds(event.Notice) {
		if n.Text == plannerResearchFallbackNotice {
			t.Fatalf("notices = %+v, want no research-limit fallback after the planner answered", sink.kinds(event.Notice))
		}
	}
}

// TestCoordinatorPlannerInvalidExitOnFinalRoundFailsClosed checks that a
// rejected plan is not reported as a delivered one.
func TestCoordinatorPlannerInvalidExitOnFinalRoundFailsClosed(t *testing.T) {
	for _, route := range []agent.PlannerRoute{agent.PlannerRoutePlanOnly, agent.PlannerRoutePlanForApproval} {
		t.Run(string(route), func(t *testing.T) {
			planner := &mockProvider{name: "planner", streams: researchExitStreams([]provider.Chunk{
				{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "plan-1", Name: "submit_plan", Arguments: `{"objective":"","steps":[]}`}},
				{Type: provider.ChunkDone},
			})}
			exec := &mockProvider{name: "executor"}
			sink := &recordSink{}
			parentReg := tool.NewRegistry()
			parentReg.Add(coordinatorTestTool{name: "read_file", readOnly: true, output: "ok"})
			policy := func(context.Context, string) agent.PlannerDecision {
				return agent.PlannerDecision{Route: route, Depth: agent.PlannerDepthFull, Reason: "explicit_boundary", MaxResearchRounds: 2}
			}
			executor := agent.New(exec, tool.NewRegistry(), sessionstore.NewSession("exec-sys"), agent.Options{}, sink)
			coord := NewCoordinatorWithPlannerPolicy(
				planner, sessionstore.NewSession("planner-sys"), nil, agent.PlannerToolRegistry(parentReg, nil, nil),
				agent.Options{MaxSteps: 0}, executor, 0, sink, policy,
			)

			err := coord.Run(context.Background(), "make the bounded change")
			if err == nil || !strings.Contains(err.Error(), "research budget") {
				t.Fatalf("Run error = %v, want the research-boundary failure", err)
			}
			if got := len(exec.requests); got != 0 {
				t.Fatalf("executor requests = %d, want execution stopped", got)
			}
		})
	}
}

// TestCoordinatorPlannerApprovalRouteSubmitsOnFinalRound covers the route the
// product actually uses. Agent + Planner routes plan-for-approval, and that
// route fails closed on a research pause, so a plan delivered in the final round
// is the only way that mode can end with a plan at all.
func TestCoordinatorPlannerApprovalRouteSubmitsOnFinalRound(t *testing.T) {
	planner := &mockProvider{name: "planner", streams: researchExitStreams([]provider.Chunk{
		{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "plan-1", Name: "submit_plan", Arguments: `{"objective":"apply the narrow fix","steps":[{"title":"edit the file"}]}`}},
		{Type: provider.ChunkDone},
	})}
	exec := &mockProvider{name: "executor", chunks: []provider.Chunk{
		{Type: provider.ChunkText, Text: "Done."},
		{Type: provider.ChunkDone},
	}}
	parentReg := tool.NewRegistry()
	parentReg.Add(coordinatorTestTool{name: "read_file", readOnly: true, output: "ok"})
	policy := func(context.Context, string) agent.PlannerDecision {
		return agent.PlannerDecision{Route: agent.PlannerRoutePlanForApproval, Depth: agent.PlannerDepthFull, Reason: "agent_planner", MaxResearchRounds: 2}
	}
	sink := &recordSink{}
	executor := agent.New(exec, tool.NewRegistry(), sessionstore.NewSession("exec-sys"), agent.Options{}, sink)
	coord := NewCoordinatorWithPlannerPolicy(
		planner, sessionstore.NewSession("planner-sys"), nil, agent.PlannerToolRegistry(parentReg, nil, nil),
		agent.Options{MaxSteps: 0}, executor, 0, sink, policy,
	)
	gate := &coordinatorApprovalGate{allow: true}
	coord.SetPlannerPlanApprover(gate)

	if err := coord.Run(context.Background(), "make the bounded change"); err != nil {
		t.Fatalf("a plan delivered in the final round must reach approval, not fail closed: %v", err)
	}
	if gate.calls != 1 {
		t.Fatalf("approval gate calls = %d, want the plan to reach approval", gate.calls)
	}
}
