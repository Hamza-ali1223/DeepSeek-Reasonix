package usecap

import (
	"fmt"
	"strings"

	"reasonix/internal/contract/tool"
	"reasonix/internal/ext/plugin"
	"reasonix/internal/runtime/capability"
)

type plannerMCPAllowlist struct {
	tools    map[string]bool
	servers  map[string]bool
	builtins map[string]bool
}

// SetPlannerToolAllowlist narrows this role-local frontend's discovery and
// dispatch catalog without changing the stable provider schema. Server grants
// include eligible tools; exact tool grants remain isolated to that tool.
func (t *UseCapabilityTool) SetPlannerToolAllowlist(names []string) {
	if t == nil || len(names) == 0 {
		return
	}
	allow := &plannerMCPAllowlist{tools: map[string]bool{}, servers: map[string]bool{}, builtins: map[string]bool{}}
	for _, name := range names {
		id := strings.TrimSpace(name)
		if id == "" {
			continue
		}
		if server, ok := ParseMCPServerCapabilityID(id); ok && !strings.Contains(server, "/") {
			allow.servers[server] = true
			continue
		}
		if rest, ok := strings.CutPrefix(id, "mcp-tool:"); ok {
			server, raw, cut := strings.Cut(rest, "/")
			server, raw = strings.TrimSpace(server), strings.TrimSpace(raw)
			if cut && server != "" && raw != "" && !strings.Contains(raw, "/") {
				allow.tools["mcp-tool:"+server+"/"+raw] = true
			}
			continue
		}
		allow.builtins[id] = true
	}
	if len(allow.tools)+len(allow.servers)+len(allow.builtins) > 0 {
		t.plannerMCP = allow
	}
}

func clonePlannerMCPAllowlist(in *plannerMCPAllowlist) *plannerMCPAllowlist {
	if in == nil {
		return nil
	}
	out := &plannerMCPAllowlist{tools: map[string]bool{}, servers: map[string]bool{}, builtins: map[string]bool{}}
	for id := range in.tools {
		out.tools[id] = true
	}
	for server := range in.servers {
		out.servers[server] = true
	}
	for name := range in.builtins {
		out.builtins[name] = true
	}
	return out
}

func (t *UseCapabilityTool) plannerAllowsServer(server string) bool {
	return t.plannerMCP == nil || t.plannerMCP.servers[server]
}

func (t *UseCapabilityTool) plannerAllowsTool(server, raw string) bool {
	if t.plannerMCP == nil {
		return true
	}
	return t.plannerMCP.servers[server] || t.plannerMCP.tools["mcp-tool:"+server+"/"+raw]
}

func (t *UseCapabilityTool) plannerAllowsServerInspect(server string) bool {
	return t.plannerMCP == nil || t.plannerMCP.servers[server]
}

// plannerScopeError reports the allowlist refusal for an MCP capability id this
// role may not reach, and nil for an id the gate does not decide. Discovery and
// dispatch ask this one question, so the narrowing cannot drift between what
// inspect shows and what call accepts.
func (t *UseCapabilityTool) plannerScopeError(id string) error {
	if t.plannerMCP == nil {
		return nil
	}
	if server, ok := ParseMCPServerCapabilityID(id); ok {
		if t.plannerAllowsServerInspect(server) {
			return nil
		}
	} else if server, raw, err := parseMCPCapabilityID(id); err == nil {
		if t.plannerAllowsTool(server, raw) && t.plannerMCPToolEligible(server, raw, false) {
			return nil
		}
	} else {
		return nil
	}
	return fmt.Errorf("capability %q is outside the planner's configured research allowlist", id)
}

// plannerAllowsCatalogDispatch decides the catalog third of the dispatch gate:
// an id that names no MCP capability must exist and pass the entry predicates.
// Server and tool ids were already decided by plannerScopeError.
func (t *UseCapabilityTool) plannerAllowsCatalogDispatch(id string) bool {
	if t.plannerMCP == nil {
		return true
	}
	if _, ok := ParseMCPServerCapabilityID(id); ok {
		return true
	}
	if _, _, err := parseMCPCapabilityID(id); err == nil {
		return true
	}
	entry, ok := t.currentCatalog().Lookup(id)
	return ok && t.plannerAllowsCapabilityEntry(entry)
}

// plannerDispatchError is the whole dispatch gate: the MCP half and the catalog
// half, asked once so a call cannot reach a capability discovery refused.
func (t *UseCapabilityTool) plannerDispatchError(id string) error {
	if err := t.plannerScopeError(id); err != nil {
		return err
	}
	if !t.plannerAllowsCatalogDispatch(id) {
		return fmt.Errorf("capability %q is outside the planner's configured research allowlist", id)
	}
	return nil
}

func (t *UseCapabilityTool) plannerAllowsCapabilityEntry(entry capability.Entry) bool {
	if t.plannerMCP == nil {
		return true
	}
	switch entry.Kind {
	case capability.KindMCPServer:
		return t.plannerMCP.servers[entry.Name]
	case capability.KindMCPTool:
		server, raw, err := parseMCPCapabilityID(entry.ID)
		return err == nil && t.plannerAllowsTool(server, raw) && t.plannerMCPToolEligible(server, raw, entry.Destructive)
	case capability.KindTool:
		return entry.ReadOnly && t.plannerMCP.builtins[entry.ToolName]
	default:
		return false
	}
}

func (t *UseCapabilityTool) plannerMCPToolEligible(server, raw string, destructive bool) bool {
	if destructive {
		return false
	}
	spec, ok := t.specFor(server)
	if !ok || !spec.ServerAuthorized() {
		return false
	}
	if cached, ok := plugin.LoadCachedSchemaForSpec(spec); ok {
		for _, candidate := range cached.Tools {
			if candidate.Name == raw && candidate.Destructive {
				return false
			}
		}
	}
	return true
}

func (t *UseCapabilityTool) plannerAllowsRuntimeMCPTool(tl tool.Tool, server, raw string) bool {
	if t.plannerMCP == nil {
		return true
	}
	meta, ok := tl.(tool.MCPMetadata)
	return ok && meta.MCPServerName() == server && meta.MCPRawToolName() == raw &&
		t.plannerAllowsTool(server, raw) && tool.IsMCPServerAuthorized(tl) && !tool.HasMCPDestructiveHint(tl)
}

func (t *UseCapabilityTool) filterPlannerServerTools(server string, tools []tool.Tool) []tool.Tool {
	if t.plannerMCP == nil {
		return tools
	}
	filtered := make([]tool.Tool, 0, len(tools))
	for _, tl := range tools {
		meta, ok := tl.(tool.MCPMetadata)
		if !ok || meta.MCPServerName() != server || !t.plannerAllowsTool(server, meta.MCPRawToolName()) {
			continue
		}
		if !tool.IsMCPServerAuthorized(tl) || tool.HasMCPDestructiveHint(tl) {
			continue
		}
		filtered = append(filtered, tl)
	}
	return filtered
}
