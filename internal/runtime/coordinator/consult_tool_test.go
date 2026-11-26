package coordinator

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"reasonix/internal/contract/event"
	"reasonix/internal/contract/provider"
	"reasonix/internal/contract/tool"
	"reasonix/internal/runtime/agent"
	"reasonix/internal/state/sessionstore"
)

func TestConsultationToolValidatesBoundsAndCallsPlanner(t *testing.T) {
	var gotQuestion, gotEvidence string
	tl := consultationTool{consult: func(_ context.Context, q, e string) (string, error) {
		gotQuestion, gotEvidence = q, e
		return "advice", nil
	}}
	got, err := tl.Execute(context.Background(), json.RawMessage(`{"question":"  inspect cache  ","evidence":" blocker "}`))
	if err != nil || got != "advice" || gotQuestion != "inspect cache" || gotEvidence != "blocker" {
		t.Fatalf("Execute = %q, %v, question=%q evidence=%q", got, err, gotQuestion, gotEvidence)
	}
	for _, raw := range []string{`{}`, `{"question":"   "}`, `{"question":"` + strings.Repeat("x", consultationQuestionMax+1) + `"}`, `{"question":"ok","evidence":"` + strings.Repeat("x", consultationEvidenceMax+1) + `"}`} {
		if _, err := tl.Execute(context.Background(), json.RawMessage(raw)); err == nil {
			t.Fatalf("Execute accepted out-of-contract args (%d bytes)", len(raw))
		}
	}
}

func TestCoordinatorConsultationResumesPlannerSessionWithinExecutorTurn(t *testing.T) {
	planner := &mockProvider{name: "planner", streams: [][]provider.Chunk{
		{{Type: provider.ChunkText, Text: "initial plan"}, {Type: provider.ChunkDone}},
		{{Type: provider.ChunkText, Text: "Use the existing cache key helper."}, {Type: provider.ChunkDone}},
	}}
	exec := &mockProvider{name: "executor", streams: [][]provider.Chunk{
		{{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "consult", Name: ConsultationToolName, Arguments: `{"question":"where is the cache key?","evidence":"agent is blocked"}`}}, {Type: provider.ChunkDone}},
		{{Type: provider.ChunkText, Text: "Done"}, {Type: provider.ChunkDone}},
	}}
	reg := tool.NewRegistry()
	executor := agent.New(exec, reg, sessionstore.NewSession("executor"), agent.Options{}, event.Discard)
	plannerSess := sessionstore.NewSession("planner")
	coord := NewCoordinator(planner, plannerSess, nil, tool.NewRegistry(), agent.Options{}, executor, 0, event.Discard, nil)
	if strings.Contains(executorToolHandoffContext(executor), ConsultationToolName) {
		t.Fatal("the handoff advertised a tool absent from the executor")
	}
	reg.Add(NewConsultationTool(coord))
	if err := coord.Run(context.Background(), "fix the issue"); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(planner.requests) != 2 {
		t.Fatalf("planner requests = %d, want initial and consultation", len(planner.requests))
	}
	if !strings.Contains(lastUser(exec.requests[0]), "ask a focused question with consult_planner") {
		t.Fatal("enabled consultation is missing from the handoff")
	}
	if !strings.Contains(lastUser(planner.requests[1]), "where is the cache key?") || !strings.Contains(lastUser(planner.requests[1]), "agent is blocked") {
		t.Fatalf("consultation request = %q", lastUser(planner.requests[1]))
	}
	if len(plannerSess.Snapshot()) != 5 {
		t.Fatalf("planner session messages = %d, want system plus two user/assistant pairs", len(plannerSess.Snapshot()))
	}
	found := false
	for _, request := range exec.requests {
		for _, message := range request.Messages {
			if strings.Contains(message.Content, "Use the existing cache key helper.") {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("executor did not receive consultation result in its tool history")
	}
}

func TestCoordinatorConsultationHasHostOwnedPerTurnLimit(t *testing.T) {
	planner := &mockProvider{name: "planner", chunks: []provider.Chunk{{Type: provider.ChunkText, Text: "initial plan"}, {Type: provider.ChunkDone}}}
	exec := &mockProvider{name: "executor", streams: [][]provider.Chunk{
		{{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "consult-1", Name: ConsultationToolName, Arguments: `{"question":"first"}`}}, {Type: provider.ChunkDone}},
		{{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "consult-2", Name: ConsultationToolName, Arguments: `{"question":"second"}`}}, {Type: provider.ChunkDone}},
		{{Type: provider.ChunkText, Text: "Done"}, {Type: provider.ChunkDone}},
	}}
	reg := tool.NewRegistry()
	executor := agent.New(exec, reg, sessionstore.NewSession("executor"), agent.Options{}, event.Discard)
	coord := NewCoordinator(planner, sessionstore.NewSession("planner"), nil, tool.NewRegistry(), agent.Options{}, executor, 0, event.Discard, nil)
	reg.Add(NewConsultationTool(coord))
	if err := coord.Run(context.Background(), "task"); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(planner.requests) != 2 {
		t.Fatalf("planner requests = %d, want initial plus one consultation", len(planner.requests))
	}
	found := false
	for _, m := range executor.Session().Snapshot() {
		if strings.Contains(m.Content, "planner consultation limit reached") {
			found = true
		}
	}
	if !found {
		t.Fatal("executor history did not receive the quota failure")
	}
}

func TestConsultedPlanNeedsHostApprovalAndKeepsHostIdentity(t *testing.T) {
	planner := &mockProvider{name: "planner", streams: append([][]provider.Chunk{{{Type: provider.ChunkText, Text: "initial advice"}, {Type: provider.ChunkDone}}}, submitPlanCall(e2ePlanArgs)...)}
	exec := &mockProvider{name: "executor", streams: [][]provider.Chunk{
		{{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{ID: "consult", Name: ConsultationToolName, Arguments: `{"question":"propose a revised plan"}`}}, {Type: provider.ChunkDone}},
		{{Type: provider.ChunkText, Text: "Done"}, {Type: provider.ChunkDone}},
	}}
	reg := tool.NewRegistry()
	executor := agent.New(exec, reg, sessionstore.NewSession("executor"), agent.Options{}, event.Discard)
	coord := NewCoordinator(planner, sessionstore.NewSession("planner"), nil, agent.PlannerToolRegistry(tool.NewRegistry(), nil, nil), agent.Options{MaxSteps: 4}, executor, 0, event.Discard, nil)
	reg.Add(NewConsultationTool(coord))
	approver := &recordingPlanApprover{allow: true}
	coord.SetPlannerPlanApprover(approver)
	if err := coord.Run(context.Background(), "task"); err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := executor.PlanContract()
	if !approver.called || got == nil || got.ID == "" || got.Revision != 1 || !got.ApprovedByUser {
		t.Fatalf("approval=%+v plan=%+v, want approved host-identified revision 1", approver, got)
	}
}
