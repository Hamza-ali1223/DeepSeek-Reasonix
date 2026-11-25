---
owner: @Hamza-ali1223
backup: @SivanCola
status: active
reviewed: 2026-11-26
---

# Planner research budget

## Scope

This contract covers what bounds a planning turn's research in Agent + Planner, what the planner is told about that
bound, and what it may still do with its last round. It changes no executor budget and no permission.

## Three controls, three questions

A round cap is not a spend cap. Each control below answers a different question, and none substitutes for another.

| Control | Answers | Default |
| --- | --- | --- |
| Research rounds | How many times the planner may research and reconsider | 12 |
| Task budget | What the whole task may spend | Off |
| Progress watch | Whether a run looks stuck | Executor only |

One round MAY carry many tool calls, so the cap bounds how often the planner changes course. It does not bound batch
size, prompt growth inside a round, retries, or elapsed time. A research-round setting MUST NOT be presented to a user
as a spend limit.

## Rules

| ID | Rule |
| --- | --- |
| B1 | A research round MUST mean one assistant reply that issues a batch of research calls, whatever the batch size. |
| B2 | The cap MUST be user-configurable, default 12, accepted range 1 through 12. |
| B3 | The default and the ceiling are host-owned. The value a session uses is the user's choice inside that range. |
| B4 | `0` MUST mean the default and MUST NOT mean unlimited. A save carrying `0` MUST be refused with a message. |
| B5 | No accepted value MAY produce an unbounded planner. Unlimited rounds are not an option. |
| B6 | The planner MUST be told the exact cap in a host-authored turn block, outside the cache-stable prefix. |
| B7 | The cap MUST be accompanied by a target two rounds below it, stated as guidance. For caps below 3 the target MUST be omitted, and the block MUST say the cap carries no reserve. |
| B8 | The host MUST NOT refuse research at the target. Only the cap closes research. |
| B9 | An accepted `submit_plan` or `conclude_no_changes` MUST end the planning turn. No research round MAY follow it. |
| B10 | At the cap, research MUST close and exactly one validated exit MUST remain callable. No research or repair round MAY follow. |
| B11 | A plan submitted before the cap MUST NOT be discarded by the pause at the cap. |
| B12 | A save MUST apply to a session opened after it, and MUST NOT change a session already open. |
| B13 | The executor's round budget MUST stay unbounded. |

## What the planner is told

The block replaces the `depth` and `route` fields. `depth` is removed because no shipped policy can produce a
non-`full` value, so it conveyed nothing.

```text
<planner-turn>
route: plan_for_approval
max_research_rounds: 12
aim_by_research_round: 10
One research round is one reply that issues a batch of research calls,
regardless of how many calls are in that batch.
Aim to finalize using the results of at most 10 research rounds.
Use rounds 11 and 12 only if an important question remains unresolved.
Submit earlier if you have enough evidence.
</planner-turn>
```

Two milestone updates are appended at the tail. The opening block MUST NOT be rewritten.

| When | Appended state |
| --- | --- |
| The batch that leaves two rounds, for caps of 3 or more | `research_rounds_remaining: 2`, target reached, remaining rounds reserved for the most important open question |
| At the cap | `research_rounds_remaining: 0`, research closed, one structured exit permitted |

A round counter field MUST NOT be sent. The planner MUST NOT be asked to count its own transcript, because a
mistaken count reads as a real one.

## Notice identities

| Situation | Means | Identity |
| --- | --- | --- |
| The cap is reached and one finalization reply is allowed | The turn may still deliver | The existing tool-budget notice |
| The turn ended with no validated plan | Nothing was delivered | A distinct terminal planner outcome |

A caller MUST NOT classify research exhaustion by reading notice text. A caller MUST NOT treat every tool-loop pause
as research exhaustion.

The terminal outcome reaches two readers separately.

| Reader | Receives |
| --- | --- |
| The user | The cause, that execution did not start, and the next action: retry planning or continue in Agent mode |
| The executor | Only on an authorized handoff: the original task, that research was incomplete, that no validated plan exists, and whether `consult_planner` is available |

On a plan-for-approval route the executor MUST NOT start merely to deliver that warning.

## Fences

| Fence | Protects | Disposition |
| --- | --- | --- |
| The loop round condition | The hard stop | Keep; the number becomes a setting |
| The finalization round | No silent cut mid-thought | Keep |
| The whole-round refusal | Makes the last round actually last | Narrow: refuse research, admit one exit |
| The planner-turn rollback | No half-turn poisons the next planner request | Keep, but it MUST NOT discard a delivered plan |
| The research-boundary error | No execution without an approved plan | Keep for plan-only and plan-for-approval |
| The subagent tool exclusion | Cost and concurrency | Keep; revisit only when planner delegation is designed |

## Enforcement

1. A cap of 1 MUST admit one reply that batches several reads, then admit a plan in the next reply, then refuse
   further research.
2. A plan submitted before the cap MUST reach the executor, and a plan submitted at the cap MUST reach approval.
3. An invalid or absent exit at the cap MUST leave execution stopped on an approval route.
4. An executor turn with no structured exit MUST still refuse every call in its finalization round.
5. A saved cap MUST change a newly opened session and MUST NOT change an open one.
