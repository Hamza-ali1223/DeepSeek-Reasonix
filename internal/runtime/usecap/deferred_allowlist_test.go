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

// Dispatch re-enters ResolveCall, so the allowlist is enforced on the path that
// actually executes and not only on the one that describes itself. A grant that
// gated only resolution would let the same call run under a later intent.
func TestPlannerAllowlistIsEnforcedOnTheExecutingPath(t *testing.T) {
	host := tool.NewRegistry()
	allowed := &allowlistFixtureTool{name: "mcp__github__search", server: "github", raw: "search"}
	denied := &allowlistFixtureTool{name: "mcp__github__delete", server: "github", raw: "delete"}
	host.Add(allowed)
	host.Add(denied)
	proxy := NewUseCapabilityTool(context.Background(), nil, []plugin.Spec{{Name: "github", Authorized: true}}, host, capability.NewLedger(), nil, nil)
	proxy.SetPlannerToolAllowlist([]string{"mcp-tool:github/search"})

	if _, err := proxy.Execute(context.Background(), json.RawMessage(
		`{"action":"call","capability_id":"mcp-tool:github/delete"}`,
	)); err == nil || !strings.Contains(err.Error(), "research allowlist") {
		t.Fatalf("dispatch of a non-allowlisted tool = %v, want the planner allowlist refusal", err)
	}
	if denied.calls != 0 {
		t.Fatalf("the non-allowlisted tool ran %d times", denied.calls)
	}

	if _, err := proxy.Execute(context.Background(), json.RawMessage(
		`{"action":"call","capability_id":"mcp-tool:github/search"}`,
	)); err != nil {
		t.Fatalf("dispatch of an allowlisted tool: %v", err)
	}
	if allowed.calls != 1 {
		t.Fatalf("the allowlisted tool ran %d times, want 1", allowed.calls)
	}
}

type allowlistFixtureTool struct {
	name   string
	server string
	raw    string
	calls  int
}

func (f *allowlistFixtureTool) Name() string        { return f.name }
func (f *allowlistFixtureTool) Description() string { return "fixture" }
func (f *allowlistFixtureTool) ReadOnly() bool      { return true }
func (f *allowlistFixtureTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object"}`)
}
func (f *allowlistFixtureTool) Execute(context.Context, json.RawMessage) (string, error) {
	f.calls++
	return "", nil
}
func (f *allowlistFixtureTool) MCPServerName() string     { return f.server }
func (f *allowlistFixtureTool) MCPRawToolName() string    { return f.raw }
func (f *allowlistFixtureTool) MCPServerAuthorized() bool { return true }
func (f *allowlistFixtureTool) MCPDestructiveHint() bool  { return false }

var _ tool.Tool = (*allowlistFixtureTool)(nil)
