package usecap

import (
	"context"
	"encoding/json"
	"fmt"

	"reasonix/internal/contract/tool"
	"reasonix/internal/ext/plugin"
	"reasonix/internal/runtime/capability"
)

// inspect renders one capability entry. For MCP entries it lists the server's
// tools without side effects: live tools when connected, cached schema
// otherwise, so resolution never starts a process or opens a connection.
func (t *UseCapabilityTool) inspect(ctx context.Context, id string) (string, error) {
	if err := t.plannerScopeError(id); err != nil {
		return "", err
	}
	cat := t.currentCatalog()
	if e, ok := cat.Lookup(id); ok {
		if !t.plannerAllowsCapabilityEntry(e) {
			return "", fmt.Errorf("capability %q is outside the planner's configured research allowlist", id)
		}
		info := map[string]any{
			"id":          e.ID,
			"kind":        e.Kind,
			"name":        e.Name,
			"description": e.Description,
			"status":      e.Status,
			"read_only":   e.ReadOnly,
			"destructive": e.Destructive,
			"auto_use":    e.AutoUse,
			"requires":    e.Requires,
			"profiles":    e.Profiles,
			"tool_name":   e.ToolName,
			"auto_start":  e.AutoStart,
		}
		if len(e.Aliases) > 0 {
			info["aliases"] = e.Aliases
		}
		// A registry tool answers with its own schema. Without it the model is
		// left guessing argument names for a tool the provider schema hides.
		if e.Kind == capability.KindTool && t.registry != nil {
			if tl, ok := t.registry.Get(e.ToolName); ok {
				info["input_schema"] = json.RawMessage(tl.Schema())
			}
		}
		b, _ := json.MarshalIndent(info, "", "  ")
		if e.Kind == capability.KindMCPServer || e.Kind == capability.KindMCPTool {
			server := e.Source
			if server == "" {
				server = e.ConnectName
			}
			toolFilter := ""
			if e.Kind == capability.KindMCPTool {
				parsedServer, raw, err := parseMCPCapabilityID(e.ID)
				if err != nil {
					return string(b), nil
				}
				if server == "" {
					server = parsedServer
				} else if parsedServer != server {
					return string(b), nil
				}
				toolFilter = raw
			}
			if server != "" {
				if !t.serverEnabled(server) {
					return string(b) + "\n\nServer is disabled in this session.", nil
				}
				if t.host != nil && t.host.HasClient(server) {
					// serverTools refreshes the snapshot too: inspecting a
					// server another tab connected restores tool routing here.
					tools, err := t.serverTools(ctx, server)
					if err != nil {
						return string(b) + "\n\nTool listing failed: " + err.Error(), nil
					}
					tools = t.filterPlannerServerTools(server, tools)
					return string(b) + "\n\nTools:\n" + inspectToolListJSON(server, filterInspectTools(tools, toolFilter)), nil
				}
				if spec, ok := t.specFor(server); ok {
					if cs, ok := plugin.LoadCachedSchemaForSpec(spec); ok && len(cs.Tools) > 0 {
						var list []inspectToolInfo
						for _, ct := range cs.Tools {
							if toolFilter != "" && ct.Name != toolFilter {
								continue
							}
							if !t.plannerAllowsTool(server, ct.Name) || !t.plannerMCPToolEligible(server, ct.Name, ct.Destructive) {
								continue
							}
							list = append(list, inspectToolInfo{
								ID:          "mcp-tool:" + server + "/" + ct.Name,
								Name:        plugin.ModelToolName(server, ct.Name),
								Description: ct.Description,
								ReadOnly:    ct.ReadOnly,
								Destructive: ct.Destructive,
								Schema:      ct.Schema,
							})
						}
						extra, _ := json.MarshalIndent(list, "", "  ")
						return string(b) + "\n\nTools (from cached schema; server not started):\n" + string(extra), nil
					}
					return string(b) + "\n\nServer not connected and no cached tool schema; call use_capability(action=\"call\", capability_id=\"mcp-server:" + server + "\") to connect (after approval) and list its tools.", nil
				}
			}
		}
		return string(b), nil
	}
	return "", fmt.Errorf("unknown capability_id %q", id)
}

// filterInspectTools narrows concrete mcp-tool inspection to that exact tool.
// Server inspection intentionally keeps the full directory. This prevents a
// restricted sub-agent allowed one tool from discovering sibling tool schemas
// through action=inspect on its allowed capability ID.
func filterInspectTools(tools []tool.Tool, raw string) []tool.Tool {
	if raw == "" {
		return tools
	}
	filtered := make([]tool.Tool, 0, 1)
	for _, tl := range tools {
		if m, ok := tl.(tool.MCPMetadata); ok && m.MCPRawToolName() == raw {
			filtered = append(filtered, tl)
			break
		}
	}
	return filtered
}

type inspectToolInfo struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	ReadOnly    bool            `json:"read_only"`
	Destructive bool            `json:"destructive,omitempty"`
	Schema      json.RawMessage `json:"input_schema,omitempty"`
}

// inspectToolListJSON renders a server's live tools as the capability-id
// directory shared by inspect and the first-discovery connect result.
func inspectToolListJSON(server string, tools []tool.Tool) string {
	var list []inspectToolInfo
	for _, tl := range tools {
		raw := ""
		if m, ok := tl.(tool.MCPMetadata); ok {
			raw = m.MCPRawToolName()
		}
		list = append(list, inspectToolInfo{
			ID:          "mcp-tool:" + server + "/" + raw,
			Name:        tl.Name(),
			Description: tl.Description(),
			ReadOnly:    tl.ReadOnly(),
			Destructive: tool.HasMCPDestructiveHint(tl),
			Schema:      tl.Schema(),
		})
	}
	extra, _ := json.MarshalIndent(list, "", "  ")
	return string(extra)
}
