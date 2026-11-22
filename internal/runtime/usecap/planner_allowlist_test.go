package usecap

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"reasonix/internal/contract/tool"
	"reasonix/internal/ext/plugin"
	"reasonix/internal/runtime/capability"
)

type plannerPolicyTool struct {
	server      string
	raw         string
	authorized  bool
	destructive bool
	readOnly    bool
}

func (t plannerPolicyTool) Name() string                                           { return "mcp__" + t.server + "__" + t.raw }
func (plannerPolicyTool) Description() string                                      { return "fixture" }
func (t plannerPolicyTool) ReadOnly() bool                                         { return t.readOnly }
func (plannerPolicyTool) Schema() json.RawMessage                                  { return json.RawMessage(`{"type":"object"}`) }
func (plannerPolicyTool) Execute(context.Context, json.RawMessage) (string, error) { return "", nil }
func (t plannerPolicyTool) MCPServerName() string                                  { return t.server }
func (t plannerPolicyTool) MCPRawToolName() string                                 { return t.raw }
func (t plannerPolicyTool) MCPServerAuthorized() bool                              { return t.authorized }
func (t plannerPolicyTool) MCPDestructiveHint() bool                               { return t.destructive }

func TestPlannerMCPAllowlistPrunesCatalogAndTypedServerTools(t *testing.T) {
	entries := []capability.Entry{
		{ID: "mcp-server:github", Kind: capability.KindMCPServer, Name: "github", Source: "github"},
		{ID: "mcp-tool:github/search", Kind: capability.KindMCPTool, Name: "github/search", Source: "github", ReadOnly: false},
		{ID: "mcp-tool:github/delete", Kind: capability.KindMCPTool, Name: "github/delete", Source: "github", ReadOnly: false, Destructive: true},
		{ID: "mcp-tool:slack/search", Kind: capability.KindMCPTool, Name: "slack/search", Source: "slack", ReadOnly: false},
		{ID: "tool:grep", Kind: capability.KindTool, Name: "grep", ToolName: "grep", ReadOnly: true},
		{ID: "skill:review", Kind: capability.KindSkill, Name: "review"},
	}
	proxy := NewUseCapabilityTool(context.Background(), nil, []plugin.Spec{
		{Name: "github", Authorized: true},
		{Name: "slack", Authorized: true},
	}, nil, nil, nil, func() capability.Catalog { return capability.Catalog{Entries: entries} })
	proxy.SetPlannerToolAllowlist([]string{"grep", "mcp-server:github"})

	listed, err := proxy.listCapabilities()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"mcp-server:github", "mcp-tool:github/search", "tool:grep"} {
		if !strings.Contains(listed, want) {
			t.Errorf("planner allowlist list omitted %q: %s", want, listed)
		}
	}
	for _, denied := range []string{"mcp-tool:github/delete", "mcp-tool:slack/search", "skill:review"} {
		if strings.Contains(listed, denied) {
			t.Errorf("planner allowlist list exposed %q: %s", denied, listed)
		}
	}

	visible := proxy.filterPlannerServerTools("github", []tool.Tool{
		plannerPolicyTool{server: "github", raw: "search", authorized: true, readOnly: false},
		plannerPolicyTool{server: "github", raw: "delete", authorized: true, destructive: true},
		plannerPolicyTool{server: "github", raw: "unauthorized", destructive: false},
		plannerPolicyTool{server: "slack", raw: "search", authorized: true},
	})
	if len(visible) != 1 || visible[0].Name() != "mcp__github__search" {
		t.Fatalf("typed server directory = %v, want only authorized non-destructive github search (even without readOnlyHint)", toolNames(visible))
	}
}

func TestPlannerExactMCPAllowlistDoesNotGrantServerOrSiblingTools(t *testing.T) {
	proxy := NewUseCapabilityTool(context.Background(), nil, []plugin.Spec{{Name: "github", Authorized: true}}, nil, nil, nil, func() capability.Catalog {
		return capability.Catalog{Entries: []capability.Entry{
			{ID: "mcp-server:github", Kind: capability.KindMCPServer, Name: "github", Source: "github"},
			{ID: "mcp-tool:github/search", Kind: capability.KindMCPTool, Name: "github/search", Source: "github"},
			{ID: "mcp-tool:github/issues", Kind: capability.KindMCPTool, Name: "github/issues", Source: "github"},
		}}
	})
	proxy.SetPlannerToolAllowlist([]string{"mcp-tool:github/search"})
	listed, err := proxy.listCapabilities()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(listed, "mcp-tool:github/search") || strings.Contains(listed, "mcp-tool:github/issues") {
		t.Fatalf("exact-tool catalog = %s, want only selected tool", listed)
	}
	if proxy.plannerAllowsServerInspect("github") {
		t.Fatal("exact tool selection granted mcp-server inspect/connect")
	}
	if !proxy.plannerAllowsTool("github", "search") || proxy.plannerAllowsTool("github", "issues") {
		t.Fatal("exact tool dispatch predicate did not isolate selected tool")
	}
}

func toolNames(tools []tool.Tool) []string {
	names := make([]string, 0, len(tools))
	for _, tl := range tools {
		names = append(names, tl.Name())
	}
	return names
}
