package coordinator

import (
	"fmt"
	"strings"

	"reasonix/internal/contract/tool"
	"reasonix/internal/runtime/agent"
)

// executorToolHandoffContext names available consultation and MCP tools.
// The planner's tool list is narrower, so it cannot speak for the executor.
func executorToolHandoffContext(a *agent.Agent) string {
	schemas := a.ToolSchemas()
	if len(schemas) == 0 {
		return ""
	}
	toolNames := make([]string, 0, len(schemas))
	mcpNames := make([]string, 0)
	var b strings.Builder
	for _, schema := range schemas {
		name := strings.TrimSpace(schema.Name)
		if name == "" {
			continue
		}
		toolNames = append(toolNames, name)
		if name == ConsultationToolName {
			b.WriteString("- If you need the planner's advice while carrying out the plan, ask a focused question with consult_planner.")
		}
		if strings.HasPrefix(name, tool.MCPNamePrefix) {
			mcpNames = append(mcpNames, name)
		}
	}
	if len(mcpNames) == 0 {
		return b.String()
	}
	if b.Len() > 0 {
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "- The executor request includes the full tool schema (%d tools).", len(toolNames))
	fmt.Fprintf(&b, "\n- MCP tools are already registered for the executor in this request (%d MCP tools). MCP tool names include: %s.", len(mcpNames), boundedToolNames(mcpNames, 16))
	return b.String()
}

func boundedToolNames(names []string, max int) string {
	if len(names) == 0 {
		return "(none)"
	}
	if max <= 0 {
		max = 1
	}
	if len(names) <= max {
		return strings.Join(names, ", ")
	}
	return fmt.Sprintf("%s, ... +%d more", strings.Join(names[:max], ", "), len(names)-max)
}
