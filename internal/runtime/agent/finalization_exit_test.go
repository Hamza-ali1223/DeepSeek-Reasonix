package agent

import (
	"context"
	"strings"
	"testing"

	"reasonix/internal/contract/event"
	"reasonix/internal/contract/provider"
	"reasonix/internal/contract/tool"
	"reasonix/internal/state/sessionstore"
)

// finalizationExitStreams scripts one research round, then whatever the grace
// round answers with, so the budget lands exactly on the second round.
func finalizationExitStreams(grace []provider.Chunk) [][]provider.Chunk {
	return [][]provider.Chunk{
		{
			{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "read-1", Name: "read_file", Arguments: `{}`}},
			{Type: provider.ChunkDone},
		},
		grace,
	}
}

// TestFinalizationRoundAdmitsOnlyTheStructuredExit is the boundary the
// screenshot hit: a planning turn out of research rounds must still deliver its
// plan, and must still refuse to research further.
func TestFinalizationRoundAdmitsOnlyTheStructuredExit(t *testing.T) {
	prov := &mockProvider{name: "planner", streams: finalizationExitStreams([]provider.Chunk{
		{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "read-2", Name: "read_file", Arguments: `{}`}},
		{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "plan-1", Name: "submit_plan", Arguments: `{"objective":"apply the fix","steps":[{"title":"edit the file"}]}`}},
		{Type: provider.ChunkDone},
	})}
	reg := tool.NewRegistry()
	reg.Add(readProbe{})
	reg.Add(NewSubmitPlanTool())

	a := New(prov, reg, sessionstore.NewSession("sys"), Options{MaxSteps: 1}, event.Discard)
	ctx, submission := WithPlanSubmission(context.Background())

	if err := a.Run(ctx, "plan the fix"); err != nil {
		t.Fatalf("a delivered plan must end the turn cleanly: %v", err)
	}
	if _, ok := submission.Plan(); !ok {
		t.Fatal("the plan submitted in the finalization round was not recorded")
	}
	msgs := a.sess.conversation.Snapshot()
	var refusedResearch bool
	for _, m := range msgs {
		if m.Role != provider.RoleTool {
			continue
		}
		switch m.ToolCallID {
		case "read-2":
			refusedResearch = true
			if !strings.Contains(m.Content, CodeResearchClosed) {
				t.Fatalf("research result = %q, want refusal %s", m.Content, CodeResearchClosed)
			}
		case "plan-1":
			if strings.Contains(m.Content, "blocked") || strings.Contains(m.Content, "error:") {
				t.Fatalf("submit_plan was refused in the finalization round: %q", m.Content)
			}
		}
	}
	if !refusedResearch {
		t.Fatal("no research refusal recorded; the round did not reach the gate")
	}
}

// TestFinalizationRoundWithoutExitPauses keeps the executor's contract: a turn
// with no structured exit still refuses every tool in its grace round.
func TestFinalizationRoundWithoutExitPauses(t *testing.T) {
	prov := &mockProvider{name: "executor", streams: finalizationExitStreams([]provider.Chunk{
		{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "read-2", Name: "read_file", Arguments: `{}`}},
		{Type: provider.ChunkDone},
	})}
	reg := tool.NewRegistry()
	reg.Add(readProbe{})

	a := New(prov, reg, sessionstore.NewSession("sys"), Options{MaxSteps: 1}, event.Discard)
	err := a.Run(context.Background(), "read everything")
	info, ok := InspectRunPause(err)
	if !ok || info.Kind != "max_steps" {
		t.Fatalf("Run error = %v, want a max_steps pause at the round budget", err)
	}
	// The call is still paired, and the pairing is the refusal: a turn with no
	// structured exit never reaches the tool.
	for _, m := range a.sess.conversation.Snapshot() {
		if m.Role == provider.RoleTool && m.ToolCallID == "read-2" {
			if !strings.Contains(m.Content, "no more tools will run") {
				t.Fatalf("grace-round result = %q, want the whole-round refusal", m.Content)
			}
			if strings.Contains(m.Content, "package main") {
				t.Fatal("the grace round ran the tool")
			}
		}
	}
}

// TestFinalizationRoundRefusedExitPauses checks that an exit the host rejects
// is not delivered: the turn pauses instead of reporting a plan it never took.
func TestFinalizationRoundRefusedExitPauses(t *testing.T) {
	prov := &mockProvider{name: "planner", streams: finalizationExitStreams([]provider.Chunk{
		{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "plan-1", Name: "submit_plan", Arguments: `{"objective":"","steps":[]}`}},
		{Type: provider.ChunkDone},
	})}
	reg := tool.NewRegistry()
	reg.Add(readProbe{})
	reg.Add(NewSubmitPlanTool())

	a := New(prov, reg, sessionstore.NewSession("sys"), Options{MaxSteps: 1}, event.Discard)
	ctx, submission := WithPlanSubmission(context.Background())

	err := a.Run(ctx, "plan the fix")
	if _, ok := InspectRunPause(err); !ok {
		t.Fatalf("Run error = %v, want a resumable pause for a rejected plan", err)
	}
	if _, ok := submission.Plan(); ok {
		t.Fatal("a rejected plan was recorded as delivered")
	}
}
