package agent

import (
	"context"
	"fmt"

	"reasonix/internal/contract/event"
	"reasonix/internal/contract/provider"
	"reasonix/internal/contract/tool"
)

// landCause is why a turn was told to finalize. kind selects the pause the Run
// ends with, so a host can tell "spent its budget" from "hit the round assert".
// exit marks a turn that may still take its structured exit, which changes what
// finalizing means: deliver the result, not summarize progress.
type landCause struct {
	kind   string
	axis   string
	detail string
	exit   bool
}

func (c landCause) nudge(state *turnRuntime) string {
	const close = "Do not call any more tools. Synthesize a final answer from the work already completed: what was accomplished, what remains, and any decision the user should make."
	if c.exit {
		return fmt.Sprintf("Your tool-call round limit (%s) has been reached. Do not call any more tools except submit_plan and conclude_no_changes: research is closed, finalization is not. Deliver this turn's result now — call submit_plan with the plan as structured data, or conclude_no_changes if no edit is needed. Use the evidence already collected and label remaining uncertainty.", state.runMaxStepsKey)
	}
	if c.kind == "task_budget" {
		return fmt.Sprintf("This task has reached its %s budget. %s Use the evidence already collected and label what is still uncertain; the user can continue in the next message.", c.axis, close)
	}
	tail := fmt.Sprintf("The user can increase %s or continue in the next turn if more work is needed.", state.runMaxStepsKey)
	if state.runLimitHostOwned {
		tail = "Use the evidence already collected, label remaining uncertainty, and keep the final answer actionable."
	}
	return fmt.Sprintf("Your tool-call round limit (%s) has been reached. %s %s", state.runMaxStepsKey, close, tail)
}

func (c landCause) noticeText() string {
	if c.exit {
		return "Research rounds are used up; asking the planner to submit its plan or conclude."
	}
	if c.kind == "task_budget" {
		return "This task reached its spend budget; asking for a final answer."
	}
	return toolBudgetNoticeText()
}

// hasStructuredExit reports whether this turn can end by delivering a result
// instead of by answering. A planning host arms PlanSubmission, and both exit
// tools read that same predicate, so a turn and its exits agree by
// construction rather than by a name the loop would have to match.
func hasStructuredExit(ctx context.Context) bool {
	_, ok := planSubmissionFromContext(ctx)
	return ok
}

// finalizationDelivered reports whether a finalizing turn has already recorded
// its structured result. It asks the submission rather than the batch, so a
// refused exit cannot read as a delivered one: only an accepted plan or
// conclusion is recorded there.
func finalizationDelivered(ctx context.Context) bool {
	submission, ok := planSubmissionFromContext(ctx)
	if !ok {
		return false
	}
	if _, ok := submission.NoChanges(); ok {
		return true
	}
	_, ok = submission.Plan()
	return ok
}

// armFinalizationRound is the single place a turn is told to stop using tools.
// The grace round it sets — not the wording — is what enforces that: unexecuted
// calls in the next round are paired and refused by stopUnexecutedBoundaryCalls.
func (a *Agent) armFinalizationRound(state *turnRuntime, cause landCause) {
	if state.graceRound {
		return
	}
	state.graceRound = true
	state.landCause = cause
	a.sess.conversation.Add(provider.Message{Role: provider.RoleUser, Content: a.withTurnPreferences(cause.nudge(state)), HostAuthored: true})
	a.svc.sink.Emit(event.Event{Kind: event.Notice, Level: event.LevelInfo, Code: event.NoticeCodeToolBudget,
		Text: cause.noticeText(), Detail: cause.detail})
}

// gracePause is the resumable stop a finalized turn ends with, chosen by what
// caused the landing. Every host-stopped turn leaves through here, so the
// end-of-turn summary is reported here too: a turn that ran out of rounds is
// the one whose unverified work matters most to report.
func (a *Agent) gracePause(ctx context.Context, state *turnRuntime) error {
	if state.landCause.kind == "task_budget" {
		return &taskBudgetPause{axis: state.landCause.axis, detail: state.landCause.detail}
	}
	return &maxStepsPause{steps: state.runMaxSteps, key: state.runMaxStepsKey, hostOwned: state.runLimitHostOwned}
}

// applyFinalizationExitGate closes research on a finalizing turn while leaving
// its structured exit open. Closing both is what left a planner that had spent
// its rounds holding a plan it could no longer deliver.
func (a *Agent) applyFinalizationExitGate(state *turnRuntime, plan *toolCallPlan) (toolOutcome, bool) {
	if state == nil || !state.graceRound || !state.landCause.exit {
		return toolOutcome{}, false
	}
	if plan.tool != nil && tool.ExitsTurn(plan.tool) {
		return toolOutcome{}, false
	}
	if plan.execTool != nil && tool.ExitsTurn(plan.execTool) {
		return toolOutcome{}, false
	}
	msg := tool.Refusal{
		Code:    CodeResearchClosed,
		Message: "this turn's research rounds are spent, so no research call runs. Deliver the result now: call submit_plan with the plan as structured data, or conclude_no_changes when no edit is needed.",
	}.String()
	return toolOutcome{output: msg, blocked: true, errMsg: firstLine(msg), refusalCode: CodeResearchClosed}, true
}

// taskBudgetPause ends a Run that spent its task budget. The work is saved and
// the next message continues it — with a fresh budget, because the user
// deciding to continue is the approval that a round counter cannot ask for.
type taskBudgetPause struct {
	axis   string
	detail string
}

func (e *taskBudgetPause) Error() string {
	return fmt.Sprintf("paused after reaching this task's %s budget (%s) — the work so far is saved; send another message to continue", e.axis, e.detail)
}
