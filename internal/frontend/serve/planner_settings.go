package serve

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"reasonix/internal/contract/config"
	"reasonix/internal/contract/tool"
)

type plannerSettings struct {
	Consultation  bool     `json:"consultation"`
	Limit         int      `json:"limit"`
	Tools         []string `json:"tools"`
	ModelRef      string   `json:"modelRef,omitempty"`
	ResearchTools []string `json:"researchTools"`
	// ResearchRounds and ConsultationTimeout are the values a session opened
	// next will use. Both are pointer-free on the way out because a reader never
	// has to tell "unsaid" from "the default".
	ResearchRounds            int `json:"researchRounds"`
	ConsultationTimeoutSecond int `json:"consultationTimeoutSeconds"`
}

// plannerSettingsEdit reads the same fields on the way in, as pointers: an older
// client that does not send one MUST leave the saved value alone rather than
// reset it to a zero the user never chose.
type plannerSettingsEdit struct {
	Consultation              *bool     `json:"consultation"`
	Limit                     int       `json:"limit"`
	Tools                     *[]string `json:"tools"`
	ResearchRounds            *int      `json:"researchRounds"`
	ConsultationTimeoutSecond *int      `json:"consultationTimeoutSeconds"`
}

func (s *Server) plannerSettings(w http.ResponseWriter, _ *http.Request) {
	cfg := config.LoadForEdit(config.UserConfigPath())
	agent := cfg.Agent
	model, researchTools := s.ctl().PlannerSettingsContext()
	writeJSON(w, plannerSettings{
		Consultation:              agent.PlannerConsultation,
		Limit:                     min(max(agent.PlannerConsultationLimit, 1), config.MaxPlannerConsultationLimit),
		Tools:                     append([]string{}, agent.PlannerTools...),
		ModelRef:                  model,
		ResearchTools:             append([]string{}, researchTools...),
		ResearchRounds:            cfg.PlannerResearchRounds(),
		ConsultationTimeoutSecond: cfg.ConsultationTimeout(),
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
	if edit.Consultation == nil || edit.Limit < 1 || edit.Limit > config.MaxPlannerConsultationLimit {
		refuse(w, http.StatusBadRequest, "planner_settings.invalid",
			"consultation is required and limit must be between 1 and "+strconv.Itoa(config.MaxPlannerConsultationLimit), nil)
		return
	}
	if edit.ResearchRounds != nil &&
		(*edit.ResearchRounds < config.MinPlannerResearchRounds || *edit.ResearchRounds > config.MaxPlannerResearchRounds) {
		refuse(w, http.StatusBadRequest, "planner_settings.invalid_research_rounds",
			"research rounds must be between "+strconv.Itoa(config.MinPlannerResearchRounds)+" and "+strconv.Itoa(config.MaxPlannerResearchRounds),
			map[string]any{"min": config.MinPlannerResearchRounds, "max": config.MaxPlannerResearchRounds})
		return
	}
	if edit.ConsultationTimeoutSecond != nil &&
		(*edit.ConsultationTimeoutSecond < config.MinConsultationTimeoutSeconds || *edit.ConsultationTimeoutSecond > config.MaxConsultationTimeoutSeconds) {
		refuse(w, http.StatusBadRequest, "planner_settings.invalid_consultation_timeout",
			"the consultation timeout must be between "+strconv.Itoa(config.MinConsultationTimeoutSeconds)+" and "+strconv.Itoa(config.MaxConsultationTimeoutSeconds)+" seconds",
			map[string]any{"min": config.MinConsultationTimeoutSeconds, "max": config.MaxConsultationTimeoutSeconds})
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
	if edit.ResearchRounds != nil {
		cfg.Agent.PlannerResearchRounds = *edit.ResearchRounds
	}
	if edit.ConsultationTimeoutSecond != nil {
		cfg.Agent.PlannerConsultationTimeoutSeconds = *edit.ConsultationTimeoutSecond
	}
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
