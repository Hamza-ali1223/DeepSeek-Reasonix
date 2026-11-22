package config

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestPlannerToolsUserConfigRoundTripsAndProjectCannotChangeIt(t *testing.T) {
	want := []string{"read_file", "mcp-tool:github/search"}
	user := "[agent]\nplanner_tools = [\"read_file\", \"mcp-tool:github/search\"]\n"
	project := "[agent]\nplanner_tools = [\"write_file\"]\n"
	cfg, _ := loadScoped(t, user, project)
	if !reflect.DeepEqual(cfg.Agent.PlannerTools, want) {
		t.Fatalf("planner_tools after project merge = %v, want user choice %v", cfg.Agent.PlannerTools, want)
	}
	if !slices.Contains(ignoredKeys(cfg), "agent.planner_tools") {
		t.Fatalf("ignored project keys = %v, want agent.planner_tools", ignoredKeys(cfg))
	}

	rendered := RenderTOMLForScope(cfg, RenderScopeUser)
	if !strings.Contains(rendered, `planner_tools = ["read_file", "mcp-tool:github/search"]`) {
		t.Fatalf("user planner tools did not render:\n%s", rendered)
	}
}

func TestProjectCannotSetPlannerToolsWithoutUserChoice(t *testing.T) {
	cfg, _ := loadScoped(t, "", "[agent]\nplanner_tools = [\"mcp-server:github\"]\n")
	if len(cfg.Agent.PlannerTools) != 0 {
		t.Fatalf("project configured planner tools = %v, want default empty allowlist", cfg.Agent.PlannerTools)
	}
	if !slices.Contains(ignoredKeys(cfg), "agent.planner_tools") {
		t.Fatalf("ignored project keys = %v, want agent.planner_tools", ignoredKeys(cfg))
	}
}
