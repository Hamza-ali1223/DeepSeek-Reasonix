---
owner: @Hamza-ali1223
backup: @SivanCola
status: active
reviewed: 2026-11-22
---

# Agent + Planner mode

## Purpose

1. Define the workflow, permissions, settings, and acceptance checks for two-model collaboration in Studio. Track delivery in [issue #33](https://github.com/Hamza-ali1223/DeepSeek-Reasonix/issues/33).
2. The planner researches and proposes a plan or gives advice. The executor carries out the work under the session's existing permissions. The host owns routing, approvals, and consultation limits.

## Work modes

| Choice | Required behavior |
| --- | --- |
| Agent | MUST preserve the existing executor workflow and routing. When consultation is enabled, the executor MAY ask the planner for advice during a turn. |
| Plan | MUST preserve the existing interactive plan-first workflow and model selection. |
| Agent + Planner | MUST route eligible user tasks to the effective dedicated planner. The planner MAY return a brief plan, a detailed structured plan, or a conclusion that no edits are needed. |
| Unavailable or failed planner | MUST identify the failure and leave execution stopped. The user MAY retry or explicitly continue with the executor alone. |
| Synthetic or control input | MUST preserve existing routing for background continuations, approval responses, and control commands. |

1. The controller MUST capture the selected mode as trusted turn state when it accepts a submission. Queue retry and reload MUST retain that mode. Older submissions without a mode MUST retain their legacy behavior.
2. The submit and inbox boundaries MUST carry planner intent as a typed field. Unsupported values MUST receive a typed refusal. Pasted text MUST NOT create or override that intent.
3. For an eligible Agent + Planner turn, the host MUST route to the dedicated planner first and validate its structured plan before handoff. A detailed plan SHOULD cover intent, evidence, steps, affected areas, approvals, risks, and verification.
4. A no-change conclusion MAY end the turn without executor work. If the planner is unavailable or fails, the host MUST report the failure and stop before execution.
5. A planning turn that spends its research rounds MUST still receive one finalization round. That round MUST refuse research calls and MUST accept exactly one validated `submit_plan` or `conclude_no_changes`.
6. A finalization round that takes neither exit MUST leave execution stopped. A rejected exit MUST NOT be reported as a delivered plan.

## Plan approval and receipt recovery

1. The host MUST default to asking the user before executing a valid plan. The user MAY choose automatic plan execution. Tool confirmation, including Allow all, MUST remain independent and retain its hard blocks.
2. Approval MUST bind to the plan revision the host showed. With plan approval enabled, the host MUST obtain fresh approval before work expands beyond the approved scope.
3. The host announces the user message's transcript index before the planner runs. A decision receipt records the user's answer. It MUST belong to that turn even when the session has no disk marker, and MUST NOT take the announced index or attach to a previous turn.
4. If approval arrives before the message lands, the host MUST hold the receipt separately until its own turn can carry it. The receipt is local metadata and MUST stay out of provider requests.
5. If restart recovery retains the user message, it MUST retain exactly one copy of its pending receipt. When no assistant reply was saved, the recovered user message MAY carry the receipt as metadata.
6. A matching commit digest means the transcript matches the recorded save. Recovery MUST verify that the receipt itself is durable before clearing the pending copy, including when the digest matches.
7. Reopening the recovered session again MUST keep the same receipt without creating a duplicate.

## Settings and providers

| Setting | Required behavior |
| --- | --- |
| Planner and executor | MUST use existing role and model settings in Studio and TOML. Studio MUST show each effective model, provider, protocol, reasoning effort, and any active project override. |
| API adapter | MUST follow each selected model's provider configuration, including Responses, Chat Completions, and Anthropic. |
| Collaboration defaults | MUST round-trip through Studio and TOML. Studio MUST show when a setting takes effect. |
| `agent.planner_consultation` | MUST default off. Enabling it opts the next eligible session into the fixed consultation tool schema. Initial explicit planning remains available without the tool. |
| `agent.planner_consultation_limit` | MUST default to one consultation per executor turn. Supported values are 1 through 8. |
| `agent.planner_tools` | MUST accept research tool names, `mcp-server:<server>`, or `mcp-tool:<server>/<tool>` in the user configuration. An empty list MUST keep the existing planner policy. |

1. Each role MUST keep separate history, provider client, and continuation state. A handoff MUST transfer a structured plan or answer with the evidence needed by the recipient.
2. Existing system prompts, tool order, and schemas MUST stay stable within their declared lifetimes. Route, plans, and consultation evidence MUST travel as turn state. Dynamic restrictions MUST be enforced at dispatch.
3. Consultation enablement, its limit, and the planner research scope apply when a session opens. Saving settings MUST NOT silently change the tool schema or limit of an already open session.
4. The planner MUST reuse existing read tools and the authorized, non-destructive MCP route. Authorization and explicit deny MUST apply to discovery and dispatch. Planner restrictions MUST preserve executor access and MUST NOT add shell write authority.

## Consultation

1. The executor MAY call [`consult_planner`](../TOOL_CONTRACT.md#conditional-executor-tool-consult_planner) for targeted advice or a plan revision during its active turn. Consultation is available independently of initial Agent + Planner routing.
2. The coordinator MUST resume the existing planner session and return the result to the executor in the same turn. Calls MUST run serially and count against the host-owned limit, which resets for each executor turn.
3. The planner MAY use the research tools available to its role while answering. The host MUST validate structured revisions, preserve plan identity and revision order, and enforce approval and permissions before expanded work takes effect.
4. The executor SHOULD send a focused question and include only the evidence needed to answer it. The executor MUST NOT choose the planner model or claim approval through tool input.

## Studio and acceptance

1. Studio MUST distinguish a planner request from actual model activity. It MUST show planning, approval, execution, consultation, and failure with runtime model attribution.
2. Studio MUST show requests, input and output tokens, and reported cached tokens by role. Missing data MUST appear as unavailable. Observed cache ratios MAY vary by provider.
3. Provider-boundary tests MUST verify configured models and endpoints, continuation, stable prefixes, and separate histories across initial planning and consultation.
4. Regressions MUST cover Agent and Plan, missing planner, approval and rejection, plan revision, cancellation, settings precedence, MCP restrictions, and wire compatibility.
5. Windows Studio verification MUST use Computer Use with an isolated Reasonix home. Live checks use the existing Responses and Chat Completions connections. Anthropic coverage uses a local test endpoint.
6. Receipt regressions MUST cover approval with no disk marker, recovery with a pending receipt and no saved reply, and two successive reopens. Recovery coverage MUST include both a matching commit digest and a save without a matching digest.
