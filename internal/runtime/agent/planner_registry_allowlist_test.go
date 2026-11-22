package agent

import (
	"context"
	"encoding/json"
	"testing"

	"reasonix/internal/contract/tool"
)

type plannerAllowlistProxy struct {
	called string
}

func (*plannerAllowlistProxy) Name() string        { return "use_capability" }
func (*plannerAllowlistProxy) Description() string { return "fixed proxy" }
func (*plannerAllowlistProxy) ReadOnly() bool      { return true }
func (*plannerAllowlistProxy) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"action":{"type":"string"}}}`)
}
func (p *plannerAllowlistProxy) Execute(_ context.Context, args json.RawMessage) (string, error) {
	p.called = string(args)
	return "", nil
}
func (p *plannerAllowlistProxy) ResolveCall(_ context.Context, args json.RawMessage) (tool.ResolvedCall, error) {
	var request struct {
		Action string `json:"action"`
	}
	_ = json.Unmarshal(args, &request)
	return tool.ResolvedCall{DisplayName: p.Name(), ProxyAction: request.Action, SkipExecute: true}, nil
}

func TestPlannerToolRegistryAllowlistKeepsControlToolsAndRestrictsMCP(t *testing.T) {
	parent := tool.NewRegistry()
	parent.Add(fakeTool{name: "read_file", readOnly: true})
	parent.Add(fakeTool{name: "grep", readOnly: true})
	parent.Add(fakeTool{name: "write_file", readOnly: false})
	parent.Add(fakeTool{name: "mcp__github__search", readOnly: true})
	proxy := &plannerAllowlistProxy{}
	parent.Add(proxy)

	planner := PlannerToolRegistry(parent, nil, []string{"grep", "mcp-tool:github/search"})
	if _, ok := planner.Get("grep"); !ok {
		t.Fatal("allowlisted builtin grep missing")
	}
	for _, name := range []string{"read_file", "write_file", "mcp__github__search"} {
		if _, ok := planner.Get(name); ok {
			t.Errorf("non-allowlisted or direct tool %q exposed", name)
		}
	}
	for _, name := range []string{"submit_plan", "conclude_no_changes"} {
		if _, ok := planner.Get(name); !ok {
			t.Errorf("planner control tool %q missing from narrowed registry", name)
		}
	}
	proxyTool, ok := planner.Get("use_capability")
	if !ok {
		t.Fatal("canonical MCP allowlist did not expose restricted proxy")
	}
	restricted, ok := proxyTool.(*plannerCapabilityProxy)
	if !ok {
		t.Fatalf("planner proxy type = %T, want plannerCapabilityProxy", proxyTool)
	}
	if !restricted.allowed["mcp-tool:github/search"] || len(restricted.allowed) != 1 {
		t.Fatalf("restricted dispatch IDs = %v, want only github/search", restricted.allowed)
	}
	if _, ok := restricted.servers["github"]; !ok || len(restricted.servers) != 1 {
		t.Fatalf("discovery servers = %v, want only github", restricted.servers)
	}

	if _, err := restricted.ResolveCall(context.Background(), json.RawMessage(`{"action":"call","capability_id":"mcp-tool:github/search"}`)); err != nil {
		t.Fatalf("allowlisted MCP dispatch was blocked: %v", err)
	}
	if _, err := restricted.ResolveCall(context.Background(), json.RawMessage(`{"action":"call","capability_id":"mcp-tool:slack/search"}`)); err == nil {
		t.Fatal("non-allowlisted MCP dispatch was accepted")
	}

	serverPlanner := PlannerToolRegistry(parent, nil, []string{"mcp-server:github"})
	serverTool, _ := serverPlanner.Get("use_capability")
	serverRestricted, ok := serverTool.(*plannerCapabilityProxy)
	if !ok {
		t.Fatalf("server-grant proxy type = %T, want plannerCapabilityProxy", serverTool)
	}
	if err := serverRestricted.check(json.RawMessage(`{"action":"call","capability_id":"mcp-tool:github/search"}`)); err != nil {
		t.Fatalf("server grant did not cover eligible server tool: %v", err)
	}
	if err := serverRestricted.check(json.RawMessage(`{"action":"call","capability_id":"mcp-tool:slack/search"}`)); err == nil {
		t.Fatal("server grant widened dispatch to another server")
	}
	if _, err := serverRestricted.ResolveCall(context.Background(), json.RawMessage(`{"action":"call","capability_id":"mcp-tool:github/search"}`)); err != nil {
		t.Fatalf("server-grant dispatch through proxy failed: %v", err)
	}

	exactPlanner := PlannerToolRegistry(parent, nil, []string{"mcp-tool:github/search"})
	exactTool, _ := exactPlanner.Get("use_capability")
	exactRestricted, ok := exactTool.(*plannerCapabilityProxy)
	if !ok {
		t.Fatalf("exact-tool proxy type = %T, want plannerCapabilityProxy", exactTool)
	}
	if err := exactRestricted.check(json.RawMessage(`{"action":"call","capability_id":"mcp-tool:github/search"}`)); err != nil {
		t.Fatalf("exact grant blocked selected tool: %v", err)
	}
	for _, id := range []string{"mcp-tool:github/other", "mcp-server:github"} {
		if err := exactRestricted.check(json.RawMessage(`{"action":"call","capability_id":"` + id + `"}`)); err == nil {
			t.Errorf("exact tool selection allowed %s", id)
		}
	}

	defaultPlanner := PlannerToolRegistry(parent, nil, nil)
	if _, ok := defaultPlanner.Get("read_file"); !ok {
		t.Fatal("empty allowlist changed the default planner research surface")
	}
	if _, ok := defaultPlanner.Get("use_capability"); !ok {
		t.Fatal("empty allowlist removed the default MCP proxy")
	}

	noMCP := PlannerToolRegistry(parent, nil, []string{"use_capability"})
	if _, ok := noMCP.Get("use_capability"); ok {
		t.Fatal("explicit use_capability entry widened to the unrestricted proxy")
	}
}
