---
owner: @Hamza-ali1223
backup: @SivanCola
status: active
reviewed: 2026-11-26
---

# Planner consultation limits

## Scope

This contract covers what bounds one `consult_planner` call and one executor turn's consultation count. It changes no
planning route and no planner research budget.

## Rules

| ID | Rule |
| --- | --- |
| C1 | The per-turn consultation limit MUST accept 1 through 20. |
| C2 | The consultation deadline MUST be user-configurable, and MUST keep 90 seconds as its default. |
| C3 | A save MUST apply to a session opened after it, and MUST NOT change a session already open. |
| C4 | The executor MUST stay cancellable while it waits for a consultation. |
| C5 | The three failure causes MUST stay distinguishable: a provider timeout, a parent cancellation, and the consultation's own deadline. |
| C6 | A timeout MUST name the setting that would allow a longer window. |

## Current bounds

| Bound | Value | Where |
| --- | --- | --- |
| Deadline | 90 seconds, a literal | `internal/runtime/coordinator/consultation.go` |
| Per-turn limit, host clamp | 8 | `SetConsultationLimit` |
| Per-turn limit, save endpoint | 8 | `internal/frontend/serve/planner_settings.go` |
| Per-turn limit, Studio input | 8 | `desktop/frontend-next/src/ui/PlannerSettings.tsx` |

## Fences

| Fence | Protects | Disposition |
| --- | --- | --- |
| The per-turn limit clamp | Cost control on a call that runs a second model | Keep; widen 8 to 20 |
| The deadline | No unbounded wait inside a turn | Keep; make it a setting |
| The typed refusal codes | A caller can tell the three causes apart | Keep |
| Parent cancellation | The user can always stop | Keep; a longer window MUST NOT defeat it |

## Enforcement

1. A limit of 20 MUST be accepted by the save endpoint, the host, and Studio alike.
2. A refused consultation MUST report one of the three typed causes, not a combined message.
3. Cancelling an executor turn MUST end a consultation in flight.
