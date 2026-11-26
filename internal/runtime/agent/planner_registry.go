package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"reasonix/internal/contract/tool"
	"reasonix/internal/runtime/usecap"
	"reasonix/internal/tools/builtin"
)

// ask is deliberately absent: a planner that needs a user-owned decision asks
// for it with the real tool, so the answer shapes the plan instead of being
// stapled onto a finished one.
var plannerNonResearchTools = []string{
	"bash_output",
	"complete_step",
	"slash_command",
	"todo_write",
	"wait",
}

func plannerResearchRegistry(parent *tool.Registry) *tool.Registry {
	return FilterReadOnlyRegistry(parent, append(SubagentMetaTools(), plannerNonResearchTools...)...)
}

func isPlannerLocalResearchTool(name string) bool {
	return name != "use_capability" && !strings.HasPrefix(name, tool.MCPNamePrefix)
}

// PlannerResearchToolNames lists the local tools eligible for planner research.
func PlannerResearchToolNames(parent *tool.Registry) []string {
	var names []string
	if base := plannerResearchRegistry(parent); base != nil {
		for _, name := range base.Names() {
			if isPlannerLocalResearchTool(name) {
				names = append(names, name)
			}
		}
	}
	return names
}

// PlannerToolRegistry builds the planner's read-only research surface, its
// always-present plan control tools, and an isolated capability proxy. A
// non-empty names list narrows builtins and MCP capabilities; empty preserves
// the existing planner surface. A server ID grants eligible non-destructive
// tools on that server, while a tool ID grants only that exact MCP tool.
func PlannerToolRegistry(parent *tool.Registry, frontend *usecap.UseCapabilityTool, names []string) *tool.Registry {
	base := plannerResearchRegistry(parent)
	sub := tool.NewRegistry()
	sub.CopyDisabledMCPFrom(parent)
	sub.Add(NewSubmitPlanTool())
	sub.Add(NewConcludeNoChangesTool())
	allowedNames := map[string]bool{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" {
			allowedNames[name] = true
		}
	}
	if base != nil {
		for _, name := range base.Names() {
			if !isPlannerLocalResearchTool(name) {
				continue
			}
			if len(names) > 0 && !allowedNames[name] {
				continue
			}
			if tl, ok := base.Get(name); ok {
				sub.Add(tl)
			}
		}
	}
	proxy := tool.Tool(nil)
	if frontend != nil {
		proxy = frontend
	} else if parent != nil {
		proxy, _ = parent.Get("use_capability")
		if local, ok := proxy.(*usecap.UseCapabilityTool); ok {
			frontend = local.CloneForAgent(nil, nil)
			proxy = frontend
		}
	}
	if frontend != nil {
		frontend.SetPlannerToolAllowlist(names)
		bindPlannerLocalTools(sub, frontend)
	}
	if proxy != nil && plannerAllowlistHasMCP(names) {
		allowed := mcpCapabilityAllowlist(parent, names)
		if resolver, ok := proxy.(tool.CallResolver); ok && len(allowed) > 0 {
			baseProxy := &restrictedCapabilityProxy{
				Tool: proxy, resolver: resolver, allowed: allowed,
				servers: serversFromCapabilityAllowlist(allowed),
			}
			sub.Add(&plannerCapabilityProxy{restrictedCapabilityProxy: baseProxy, serverGrants: explicitPlannerServerGrants(allowed)})
		}
	} else if proxy != nil && len(names) == 0 {
		sub.Add(proxy)
	}
	return sub
}

type plannerCapabilityProxy struct {
	*restrictedCapabilityProxy
	serverGrants map[string]bool
}

func explicitPlannerServerGrants(allowed map[string]bool) map[string]bool {
	grants := map[string]bool{}
	for id := range allowed {
		if server, ok := validMCPServerCapabilityID(id); ok {
			grants[server] = true
		}
	}
	return grants
}

func (t *plannerCapabilityProxy) check(args json.RawMessage) error {
	var p struct {
		Action       string `json:"action"`
		CapabilityID string `json:"capability_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return fmt.Errorf("invalid args: %w", err)
	}
	action := strings.ToLower(strings.TrimSpace(p.Action))
	if usecap.IsIDlessCatalogAction(action) {
		return nil
	}
	id := strings.TrimSpace(p.CapabilityID)
	if id == "" {
		return fmt.Errorf("capability_id is required")
	}
	if server, _, ok := validMCPToolCapabilityID(id); ok && t.serverGrants[server] {
		return nil
	}
	if !t.allowed[id] {
		return fmt.Errorf("capability %q is outside the planner's allowed tools", id)
	}
	return nil
}

func (t *plannerCapabilityProxy) ResolveCall(ctx context.Context, args json.RawMessage) (tool.ResolvedCall, error) {
	if err := t.check(args); err != nil {
		return tool.ResolvedCall{}, err
	}
	return t.resolver.ResolveCall(ctx, args)
}

func (t *plannerCapabilityProxy) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	if err := t.check(args); err != nil {
		return "", err
	}
	return t.Tool.Execute(ctx, args)
}

func plannerAllowlistHasMCP(names []string) bool {
	for _, name := range names {
		if strings.HasPrefix(strings.TrimSpace(name), "mcp-server:") || strings.HasPrefix(strings.TrimSpace(name), "mcp-tool:") {
			return true
		}
	}
	return false
}

func bindPlannerLocalTools(reg *tool.Registry, frontend *usecap.UseCapabilityTool) {
	if reg != nil && frontend != nil {
		frontend.BindLocalTools(builtin.ForkFileViews(reg))
	}
}
