---
owner: @Hamza-ali1223
backup: @SivanCola
status: active
reviewed: 2026-11-22
---

# Agent + Planner mode

## Purpose

1. Define the required behavior and acceptance tests for explicit two-model collaboration in Studio. Track delivery in [issue #33](https://github.com/Hamza-ali1223/DeepSeek-Reasonix/issues/33).

## Work modes

| Choice | Required behavior |
| --- | --- |
| Agent | MUST preserve the existing executor workflow and routing. |
| Plan | MUST preserve the existing interactive plan-first workflow and model selection. |
| Agent + Planner | MUST route eligible user tasks to the effective dedicated planner. The planner MAY return a brief plan, a detailed structured plan, or a conclusion that no edits are needed. |
| Unavailable or failed planner | MUST identify the failure and leave execution stopped. The user MAY retry or explicitly continue with the executor alone. |
| Synthetic or control input | MUST preserve existing routing for background continuations, approval responses, and control commands. |

1. The controller MUST own the mode decision and convey it as trusted turn state.
2. The host MUST validate a structured plan before handoff. Detailed plans SHOULD cover intent, evidence, ordered steps, affected areas, approvals, risks, and verification.
3. The host MUST default to asking before executing a plan. The user MAY select automatic execution after a valid plan. Tool confirmation, including Allow all, MUST remain independent with its existing hard blocks.
4. The host MUST bind approval to the plan revision it shows. If consultation expands approved work and plan approval is enabled, the host MUST obtain fresh approval before affected edits.
5. A submission MUST capture its selected work mode when accepted. A queued submission MUST retain it through retry and reload. Older submissions without a mode MUST retain their legacy behavior.
6. The submit and inbox boundaries MUST accept explicit planner intent as a typed field. Unsupported values MUST receive a typed refusal. Pasted text MUST NOT create or override that intent.

## Settings and providers

| Setting | Required behavior |
| --- | --- |
| Planner and executor | MUST use existing role and model settings in Studio and TOML. Studio MUST show each effective model, provider, protocol, reasoning effort, and any active project override. |
| API adapter | MUST follow each selected model's provider configuration, including Responses, Chat Completions, and Anthropic. |
| Collaboration defaults | MUST round-trip through Studio and TOML. Studio MUST show when a setting takes effect. |
| `agent.planner_consultation` | MUST default off. Enabling it opts the next eligible session into the fixed consultation tool schema. Initial explicit planning remains available without the tool. |
| `agent.planner_consultation_limit` | MUST default to one consultation per executor turn. A configured limit MUST be positive and bounded. |
| Tool access | MUST allow users to narrow eligible planner research tools and MCP servers or individual tools. Authorization and explicit deny MUST apply to discovery and dispatch. |

1. Each role MUST keep separate history, provider client, and continuation state. A handoff MUST transfer a structured plan or answer with the evidence needed by the recipient.
2. Existing system prompts, tool order, and schemas MUST stay stable within their declared lifetimes. Route, plans, and consultation evidence MUST travel as turn state. Dynamic restrictions MUST be enforced at dispatch.
3. The consultation schema MUST be enabled at a collaboration-enabled session boundary and remain stable within that session. Ordinary sessions MUST preserve their existing tool schemas.
4. The planner MUST reuse existing read tools and the authorized, non-destructive MCP route. Planner restrictions MUST preserve executor access and MUST NOT add shell write authority.

## Consultation

| Field | Required behavior |
| --- | --- |
| `question` | Required, non-empty, and at most 2,000 characters. |
| `evidence` | Optional and at most 8,000 characters. SHOULD identify the blocker and the decision needed. |
| Result | MUST return advice, a host-validated revision, or a typed failure. The executor MUST NOT choose a model or claim approval through tool input. |

1. The coordinator MUST resume the role's existing planner session from the active executor turn. It MUST avoid a recursive or second top-level run and MUST finish before the executor proceeds.
2. Consultations MUST run serially and have a host-owned per-turn limit of one by default. The limit MAY be configured. Cancellation, timeout, quota exhaustion, and an unavailable planner MUST have distinct outcomes.
3. The host MUST preserve plan identity and revision order, reject stale revisions, and enforce approval and permissions before an expanded plan takes effect.

## Studio and acceptance

1. Studio MUST distinguish a planner request from actual model activity. It MUST show planning, approval, execution, consultation, and failure with runtime model attribution.
2. Studio MUST show requests, input and output tokens, and reported cached tokens by role. Missing data MUST appear as unavailable. Observed cache ratios MAY vary by provider.
3. Provider-boundary tests MUST verify configured models and endpoints, continuation, stable prefixes, and separate histories across initial planning and consultation.
4. Regressions MUST cover Agent and Plan, missing planner, approval and rejection, plan revision, cancellation, reload, settings precedence, MCP restrictions, and wire compatibility.
5. Windows Studio verification MUST use Computer Use with an isolated Reasonix home. Live checks use the existing Responses and Chat Completions connections. Anthropic coverage uses a local test endpoint.
