package boot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"reasonix/internal/contract/event"
	"reasonix/internal/runtime/coordinator"
	"reasonix/internal/session/control"
)

// TestEffectPlannerModeBaselineUsesConfiguredAdaptersAndKeepsRoleHistories
// exercises the explicit plan-first marker and an ordinary follow-up turn at
// the provider wire boundary. Planner and executor share an HTTP fixture, but
// use distinct configured models and protocol endpoints.
func TestEffectPlannerModeBaselineUsesConfiguredAdaptersAndKeepsRoleHistories(t *testing.T) {
	home, root := compositionTestWorkspace(t, "planner-wire")
	t.Setenv("PLANNER_WIRE_TEST_KEY", "test-key")
	type capture struct {
		path string
		body map[string]json.RawMessage
	}
	var mu sync.Mutex
	var requests []capture
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode %s request: %v", r.URL.Path, err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		mu.Lock()
		requests = append(requests, capture{path: r.URL.Path, body: body})
		mu.Unlock()
		w.Header().Set("Content-Type", "text/event-stream")
		if r.URL.Path == "/v1/responses" {
			_, _ = io.WriteString(w, answeredResponsesSSE)
			return
		}
		_, _ = fmt.Fprintln(w, `data: {"choices":[{"index":0,"delta":{"content":"Executor response."},"finish_reason":null}]}`)
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, `data: {"choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`)
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, "data: [DONE]")
		_, _ = fmt.Fprintln(w)
	}))
	t.Cleanup(server.Close)
	writeFile(t, home, "config.toml", "[agent]\nplanner_consultation = true\nplanner_consultation_limit = 2\n")
	writeFile(t, root, "reasonix.toml", fmt.Sprintf(`
default_model = "executor"

[agent]
planner_model = "planner"
planner_consultation = true
system_prompt = "WIRE PREFIX"

[codegraph]
enabled = false

[[providers]]
name = "executor"
kind = "openai"
base_url = %q
model = "executor-wire-model"
api_key_env = "PLANNER_WIRE_TEST_KEY"

[[providers]]
name = "planner"
kind = "responses"
base_url = %q
model = "planner-wire-model"
api_key_env = "PLANNER_WIRE_TEST_KEY"
`, server.URL, server.URL))
	approveWorkspaceAt(t, home, root)
	ctrl, err := Build(t.Context(), Options{Home: home, WorkspaceRoot: root, Sink: event.Discard})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	t.Cleanup(ctrl.Close)
	ctrl.EnsureSessionPath()

	if err := ctrl.RunTurn(t.Context(), control.PlannerRouteMarker+" inspect the first fixture"); err != nil {
		t.Fatalf("explicit planner turn: %v", err)
	}
	if err := ctrl.RunTurn(t.Context(), "ordinary follow-up work"); err != nil {
		t.Fatalf("ordinary turn: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	var planner, executor []capture
	for _, req := range requests {
		switch req.path {
		case "/responses":
			planner = append(planner, req)
		case "/chat/completions":
			executor = append(executor, req)
		default:
			t.Errorf("unexpected provider endpoint %q", req.path)
		}
	}
	if len(planner) == 0 {
		t.Fatal("Responses planner received no requests for the explicit plan-first turn")
	}
	if len(executor) < 2 {
		t.Fatalf("Chat Completions executor requests = %d, want handoff and ordinary work", len(executor))
	}
	var plannerModel string
	for i, req := range planner {
		if err := json.Unmarshal(req.body["model"], &plannerModel); err != nil || plannerModel != "planner-wire-model" {
			t.Fatalf("planner request %d model = %q (decode error %v), want planner-wire-model", i+1, plannerModel, err)
		}
	}
	var executorModel string
	for i, req := range executor {
		if err := json.Unmarshal(req.body["model"], &executorModel); err != nil || executorModel != "executor-wire-model" {
			t.Fatalf("executor request %d model = %q (decode error %v), want executor-wire-model", i+1, executorModel, err)
		}
		if !strings.Contains(string(req.body["tools"]), coordinator.ConsultationToolName) {
			t.Fatalf("opted-in executor request %d omitted the fixed consultation schema: %s", i+1, req.body["tools"])
		}
	}

	var plannerPrefix, plannerTools string
	sawExplicitTask := false
	for i, req := range planner {
		var instructions string
		if err := json.Unmarshal(req.body["instructions"], &instructions); err != nil || !strings.Contains(instructions, "You are the planner in a two-model coding agent.") {
			t.Fatalf("planner request %d instructions = %q (decode error %v), want planner role prefix", i+1, instructions, err)
		}
		if i == 0 {
			plannerPrefix, plannerTools = instructions, string(req.body["tools"])
		} else if instructions != plannerPrefix || string(req.body["tools"]) != plannerTools {
			t.Fatalf("planner cache prefix changed across its requests: instructions=%t tools=%t", instructions == plannerPrefix, string(req.body["tools"]) == plannerTools)
		}
		var plannerInput any
		if err := json.Unmarshal(req.body["input"], &plannerInput); err != nil {
			t.Fatalf("decode planner input: %v", err)
		}
		plannerWire, _ := json.Marshal(plannerInput)
		if strings.Contains(string(plannerWire), "ordinary follow-up work") {
			t.Errorf("ordinary turn leaked into planner history: %s", plannerWire)
		}
		sawExplicitTask = sawExplicitTask || strings.Contains(string(plannerWire), "inspect the first fixture")
	}
	if !sawExplicitTask {
		t.Fatal("explicit plan-first task absent from every planner request")
	}

	var firstMessages, secondMessages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(executor[0].body["messages"], &firstMessages); err != nil {
		t.Fatalf("decode first executor messages: %v", err)
	}
	if err := json.Unmarshal(executor[1].body["messages"], &secondMessages); err != nil {
		t.Fatalf("decode second executor messages: %v", err)
	}
	if len(firstMessages) == 0 || firstMessages[0].Role != "system" || !strings.Contains(firstMessages[0].Content, "WIRE PREFIX") {
		t.Fatalf("executor system prefix missing: %+v", firstMessages)
	}
	if len(secondMessages) == 0 || secondMessages[0].Role != "system" || secondMessages[0].Content != firstMessages[0].Content {
		t.Fatalf("executor prefix changed across turns: first=%q second=%+v", firstMessages[0].Content, secondMessages)
	}
	secondWire, _ := json.Marshal(secondMessages)
	if strings.Contains(string(secondWire), "ordinary follow-up work") == false {
		t.Fatalf("ordinary user turn absent from executor history: %s", secondWire)
	}
	t.Logf("wire captures: planner Responses=%d model=%s; executor Chat Completions=%d model=%s", len(planner), plannerModel, len(executor), executorModel)
}
