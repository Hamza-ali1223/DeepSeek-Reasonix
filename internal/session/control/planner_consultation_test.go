package control

import (
	"context"
	"strings"
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

func TestAgentConsultationReturnsSubmittedPlanWithoutApproval(t *testing.T) {
	for _, tc := range []struct {
		name, approvalField string
	}{
		{name: "ordinary submission"},
		{name: "planner requests approval", approvalField: `,"requires_approval":true`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("REASONIX_HOME", testenv.TempDir(t))
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			const consultID = "consult-1"
			const objective = "Fix the cache key"
			const step = "Include the namespace"
			planner := &recordingProvider{name: "planner", streams: [][]provider.Chunk{
				toolCallTurn("submit-1", "submit_plan", `{"objective":"`+objective+`","steps":[{"title":"`+step+`"}]`+tc.approvalField+`}`),
				textTurn("Plan submitted"),
			}}
			executor := &recordingProvider{name: "executor", streams: [][]provider.Chunk{
				toolCallTurn(consultID, coordinator.ConsultationToolName, `{"question":"How should I fix the cache key?"}`),
				textTurn("I have the plan"),
			}}
			registry := tool.NewRegistry()
			exec := agent.New(executor, registry, sessionstore.NewSession("executor system"), agent.Options{}, event.Discard)
			coord := coordinator.NewCoordinatorWithPlannerPolicy(planner, sessionstore.NewSession("planner system"), nil, agent.PlannerToolRegistry(tool.NewRegistry(), nil, nil), agent.Options{}, exec, 0, event.Discard, NewPlannerPolicy(0))
			registry.Add(coordinator.NewConsultationTool(coord))
			var approval *event.Approval
			c := New(Options{Runner: coord, Executor: exec, Sink: event.FuncSink(func(e event.Event) {
				if e.Kind == event.ApprovalRequest {
					approval = &e.Approval
					cancel()
				}
			})})
			t.Cleanup(c.Close)
			c.EnableInteractiveApproval()
			c.SetToolApprovalMode(ToolApprovalAuto)
			if c.PlanApprovalSettings().Auto {
				t.Fatal("expected plan approval to be set to Ask")
			}

			err := c.RunTurn(ctx, "Fix the cache key with the planner's advice")
			if approval != nil {
				t.Fatalf("consultation requested approval: %+v", approval)
			}
			if err != nil {
				t.Fatal(err)
			}
			if decisions := c.Decisions(); len(decisions) != 0 {
				t.Fatalf("unexpected pending decisions: %+v", decisions)
			}
			if contract := exec.PlanContract(); contract != nil {
				t.Fatalf("consultation installed a plan contract: %+v", contract)
			}
			if len(executor.requests) < 2 {
				t.Fatalf("executor made %d requests, want a request after consultation", len(executor.requests))
			}
			for _, msg := range executor.requests[1].Messages {
				if msg.Role == provider.RoleTool && msg.ToolCallID == consultID && msg.Name == coordinator.ConsultationToolName {
					if !strings.Contains(msg.Content, objective) || !strings.Contains(msg.Content, step) {
						t.Fatalf("executor received %q, want the submitted objective and step", msg.Content)
					}
					return
				}
			}
			t.Fatal("executor did not receive the consultation tool result")
		})
	}
}
