package serve

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"reasonix/internal/contract/config"
)

func TestPlannerSettingsPersistsConsultationForTheNextSession(t *testing.T) {
	s := newProviderEditServer(t)
	s.AllowProviderEdit()
	srv := httptest.NewServer(operatorHandler(s))
	defer srv.Close()

	resp := postProvider(t, srv.URL, "/planner-settings", `{"consultation":true,"limit":2,"tools":["read_file","mcp-tool:docs/search"]}`)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /planner-settings = %d, want 200", resp.StatusCode)
	}
	agent := config.LoadForEdit(config.UserConfigPath()).Agent
	if !agent.PlannerConsultation || agent.PlannerConsultationLimit != 2 {
		t.Fatalf("consultation = %v, limit = %d", agent.PlannerConsultation, agent.PlannerConsultationLimit)
	}
	if !slices.Equal(agent.PlannerTools, []string{"read_file", "mcp-tool:docs/search"}) {
		t.Fatalf("planner tools = %v", agent.PlannerTools)
	}
	current := readJSON[plannerSettings](t, srv.URL, "/planner-settings")
	if !slices.Equal(current.Tools, agent.PlannerTools) {
		t.Fatalf("GET tools = %v, saved = %v", current.Tools, agent.PlannerTools)
	}
	resp = postProvider(t, srv.URL, "/planner-settings", `{"consultation":true,"limit":0}`)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("POST /planner-settings with zero limit = %d, want 400", resp.StatusCode)
	}
	resp = postProvider(t, srv.URL, "/planner-settings", `{"consultation":false,"limit":1,"tools":[]}`)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST default planner settings = %d, want 200", resp.StatusCode)
	}
	current = readJSON[plannerSettings](t, srv.URL, "/planner-settings")
	if current.Consultation || current.Limit != 1 || len(current.Tools) != 0 {
		t.Fatalf("planner settings did not return to defaults: %+v", current)
	}
}
