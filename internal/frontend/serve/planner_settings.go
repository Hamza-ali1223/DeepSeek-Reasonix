package serve

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"

	"reasonix/internal/contract/config"
	"reasonix/internal/contract/tool"
)

type plannerSettings struct {
	Consultation bool     `json:"consultation"`
	Limit        int      `json:"limit"`
	Tools        []string `json:"tools"`
}

type plannerSettingsEdit struct {
	Consultation *bool     `json:"consultation"`
	Limit        int       `json:"limit"`
	Tools        *[]string `json:"tools"`
}

func (s *Server) plannerSettings(w http.ResponseWriter, _ *http.Request) {
	agent := config.LoadForEdit(config.UserConfigPath()).Agent
	writeJSON(w, plannerSettings{
		Consultation: agent.PlannerConsultation,
		Limit:        min(max(agent.PlannerConsultationLimit, 1), 8),
		Tools:        append([]string{}, agent.PlannerTools...),
	})
}

func (s *Server) savePlannerSettings(w http.ResponseWriter, r *http.Request) {
	if !s.grants.at(r).providerEdit {
		refuse(w, http.StatusForbidden, "planner_settings.editing_disabled", "planner settings editing is not enabled on this server", nil)
		return
	}
	var edit plannerSettingsEdit
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&edit); err != nil {
		badBody(w)
		return
	}
	if edit.Consultation == nil || edit.Limit < 1 || edit.Limit > 8 {
		refuse(w, http.StatusBadRequest, "planner_settings.invalid", "consultation is required and limit must be between 1 and 8", nil)
		return
	}
	if edit.Tools != nil {
		tools := *edit.Tools
		if len(tools) > 128 {
			refuse(w, http.StatusBadRequest, "planner_settings.too_many_tools", "the planner tool list supports up to 128 entries", nil)
			return
		}
		seen := make(map[string]bool, len(tools))
		for i, value := range tools {
			name := strings.TrimSpace(value)
			if seen[name] || !validPlannerToolName(name) {
				refuse(w, http.StatusBadRequest, "planner_settings.invalid_tool", "expected a unique read tool name or MCP capability ID", map[string]any{"tool": name})
				return
			}
			tools[i], seen[name] = name, true
		}
	}
	path := config.UserConfigPath()
	unlock := config.LockUserConfigEdits()
	cfg := config.LoadForEdit(path)
	cfg.Agent.PlannerConsultation = *edit.Consultation
	cfg.Agent.PlannerConsultationLimit = edit.Limit
	if edit.Tools != nil {
		cfg.Agent.PlannerTools = slices.Clone(*edit.Tools)
	}
	err := cfg.SaveTo(path)
	unlock()
	if err != nil {
		saveFailed(w, http.StatusInternalServerError, "planner_settings.save_failed", err)
		return
	}
	s.plannerSettings(w, r)
}

func validPlannerToolName(name string) bool {
	if name == "" || strings.ContainsAny(name, " \r\n\t*?[") {
		return false
	}
	if id, ok := strings.CutPrefix(name, "mcp-server:"); ok {
		return id != "" && !strings.Contains(id, "/")
	}
	if id, ok := strings.CutPrefix(name, "mcp-tool:"); ok {
		server, name, ok := strings.Cut(id, "/")
		return ok && server != "" && name != "" && !strings.Contains(name, "/")
	}
	switch name {
	case "use_capability", "submit_plan", "conclude_no_changes":
		return false
	}
	builtin, exists := tool.LookupBuiltin(name)
	return !strings.ContainsAny(name, ":/") && (!exists || builtin.ReadOnly())
}
