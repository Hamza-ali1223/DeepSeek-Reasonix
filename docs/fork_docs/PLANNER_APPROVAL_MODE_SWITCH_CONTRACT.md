---
owner: @Hamza-ali1223
backup: @SivanCola
status: active
reviewed: 2026-11-26
---

# Planner approval and mode switch

## Scope

1. This contract governs approval of the initial Agent + Planner plan, the executor handoff, and the mode for later Studio submissions in that session.
2. Ordinary Plan and later consultations follow the [Agent + Planner spec](agent+planner-mode-spec.md) and the [tool contract](../TOOL_CONTRACT.md#conditional-executor-tool-consult_planner).

## Rules

| ID | Rule |
| --- | --- |
| A1 | The host MUST show the approval card for the initial valid Agent + Planner plan, even when automatic plan approval is enabled. Execution MUST wait for the user's approval. |
| A2 | Approval MUST apply to the plan revision on the card. The host MUST hand that revision to the executor and start the approved work at most once. |
| A3 | When the host accepts approval, Studio MUST select Agent for new submissions in that session, without waiting for execution to finish. A newer manual choice MUST take precedence. |
| A4 | A queued message MUST keep its submitted mode through queueing, retry, and reload. Running that message uses its captured mode for its own turn. |
| A5 | Rejection, cancellation before approval, planner failure, and a no-change conclusion MUST leave execution stopped and MUST NOT trigger the switch. |
| A6 | The switch MUST come from a host-confirmed decision for the session and turn. An old approval MUST NOT execute again or reset a newer mode choice. |
| A7 | The existing approval receipt, recovery, and tool permission rules MUST apply. Ordinary Plan and consulted revisions MUST keep their approval policies. |
| A8 | The handoff MUST tell the executor it MAY ask a focused question with the consultation tool when that tool is available. |

## Enforcement

1. Enable automatic plan approval and submit an Agent + Planner task. The card MUST appear before execution. Accept it and verify that the approved revision starts once.
2. Check that the next new submission uses Agent routing. Repeat after entering Agent + Planner from ordinary Plan.
3. Reject or cancel the plan, fail the planner, or conclude that no changes are needed. Each case MUST leave execution stopped and preserve the selected mode.
4. Queue a follow-up before approval. Its captured mode MUST survive acceptance, retry, and reload.
5. A duplicate or delayed acceptance MUST preserve a newer manual choice. An acceptance in another session MUST leave this session's mode alone.
6. Reopen a session with an approval receipt. It MUST show one receipt and MUST NOT replay an old approval as a new switch.
7. Verify ordinary Plan, consulted revisions, and tool confirmations. Verify that consultation guidance appears only for a session that offers the tool.
