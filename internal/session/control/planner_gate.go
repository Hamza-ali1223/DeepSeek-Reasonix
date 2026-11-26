package control

import (
	"context"
	"reasonix/internal/state/sessionstore"
	"strings"

	"reasonix/internal/contract/agentpreset"
	"reasonix/internal/contract/config"
	"reasonix/internal/runtime/agent"
	"reasonix/internal/runtime/taskpolicy"
)

// The planner gate answers one question: did the user ask for a plan? Twelve
// keyword tables used to score the answer out of vocabulary; the executor reads
// the same words with more context and plans for itself when a task needs it.

const (
	plannerReasonExplicitPlanMode   = "explicit_plan_mode"
	plannerReasonSynthetic          = "synthetic"
	plannerReasonSlash              = "slash_command"
	plannerReasonEmpty              = "empty_turn"
	plannerReasonUserPlanAndExecute = "user_plan_and_execute"
	plannerReasonExecutorOwns       = "executor_owns_the_turn"
)

type plannerTurnMetadata struct {
	UserText         string
	Synthetic        bool
	ExplicitPlanMode bool
	WorkMode         WorkMode
	Policy           taskpolicy.TaskPolicy
	PolicySet        bool
}

type plannerTurnMetadataKey struct{}

func withPlannerTurnMetadata(ctx context.Context, meta plannerTurnMetadata) context.Context {
	return context.WithValue(ctx, plannerTurnMetadataKey{}, meta)
}

func plannerTurnMetadataFromContext(ctx context.Context) (plannerTurnMetadata, bool) {
	if ctx == nil {
		return plannerTurnMetadata{}, false
	}
	meta, ok := ctx.Value(plannerTurnMetadataKey{}).(plannerTurnMetadata)
	return meta, ok
}

func (c *Controller) withPlannerTurnMetadata(ctx context.Context, userText string, synthetic bool) context.Context {
	workMode := workModeFromContext(ctx)
	explicitCollaboration := eligibleAgentPlannerTurn(workMode, userText, synthetic)
	planMode := c.PlanMode() && !explicitCollaboration
	policy := taskpolicy.Derive(taskpolicy.Input{
		Preset:   agentpreset.Normalize(c.AgentPreset()),
		PlanMode: planMode,
	})
	ctx = taskpolicy.WithContext(ctx, policy)
	return withPlannerTurnMetadata(ctx, plannerTurnMetadata{
		UserText:         userText,
		Synthetic:        synthetic,
		ExplicitPlanMode: planMode,
		WorkMode:         workMode,
		Policy:           policy,
		PolicySet:        true,
	})
}

// eligibleAgentPlannerTurn is the common raw-input boundary for the typed
// collaboration intent. Host-generated, empty, and control turns keep the
// normal Plan workflow even when their parent queue item carried this mode.
func eligibleAgentPlannerTurn(mode WorkMode, userText string, synthetic bool) bool {
	text := strings.TrimSpace(sessionstore.StripTransientUserBlocks(userText))
	return mode == WorkModeAgentPlanner && !synthetic && text != "" &&
		!strings.HasPrefix(text, "/") && !IsSyntheticUserMessage(text)
}

// DecidePlannerRoute routes a turn to the two-model planner only when the user
// asked for a plan. It never calls a model and never scores the task text.
func DecidePlannerRoute(ctx context.Context, input string) agent.PlannerDecision {
	meta, hasMeta := plannerTurnMetadataFromContext(ctx)
	composedText := strings.TrimSpace(sessionstore.StripTransientUserBlocks(input))
	text := composedText
	if hasMeta && strings.TrimSpace(meta.UserText) != "" {
		text = strings.TrimSpace(meta.UserText)
	}

	if meta.Synthetic || IsSyntheticUserMessage(text) {
		return plannerExecutorDecision(plannerReasonSynthetic)
	}
	if text == "" {
		return plannerExecutorDecision(plannerReasonEmpty)
	}
	if strings.HasPrefix(text, "/") {
		return plannerExecutorDecision(plannerReasonSlash)
	}
	if eligibleAgentPlannerTurn(meta.WorkMode, text, meta.Synthetic) {
		return plannerPlanDecision(agent.PlannerRoutePlanForApproval, agent.PlannerDepthFull, "agent_planner")
	}
	if meta.ExplicitPlanMode || strings.HasPrefix(composedText, PlanModeMarker) {
		return plannerExecutorDecision(plannerReasonExplicitPlanMode)
	}

	if strings.HasPrefix(composedText, PlannerRouteMarker) {
		return plannerPlanDecision(agent.PlannerRoutePlanAndExecute, agent.PlannerDepthFull, plannerReasonUserPlanAndExecute)
	}
	// Nobody asked for a plan. The executor decides for itself whether this
	// turn needs steps, and states them with todo_write where the delivery
	// gates and the contract can both read them.
	return plannerExecutorDecision(plannerReasonExecutorOwns)
}

func plannerExecutorDecision(reason string) agent.PlannerDecision {
	return agent.PlannerDecision{
		Route:  agent.PlannerRouteExecutorOnly,
		Depth:  agent.PlannerDepthNone,
		Reason: reason,
	}
}

// plannerPlanDecision stamps the default cap on every planning route. Routing
// answers whether a plan is wanted and cannot see the user's setting, so it
// states the default rather than nothing: in this loop zero rounds means no
// bound at all, and a routing decision used on its own must never be unbounded.
// NewPlannerPolicy replaces it with the session's configured value.
func plannerPlanDecision(route agent.PlannerRoute, depth agent.PlannerDepth, reason string) agent.PlannerDecision {
	return agent.PlannerDecision{
		Route:             route,
		Depth:             depth,
		Reason:            reason,
		MaxResearchRounds: config.DefaultPlannerResearchRounds,
	}
}

// NewPlannerPolicy returns the policy the two-model product path uses, with the
// research cap this session opened with. Routing answers whether a plan is
// wanted; the budget bounds that answer, and zero would leave it unbounded.
func NewPlannerPolicy(rounds int) agent.PlannerPolicy {
	if rounds <= 0 {
		rounds = config.DefaultPlannerResearchRounds
	}
	return func(ctx context.Context, input string) agent.PlannerDecision {
		d := DecidePlannerRoute(ctx, input)
		if d.Route != agent.PlannerRouteExecutorOnly {
			d.MaxResearchRounds = rounds
		}
		return d
	}
}

// NewPlannerGate retains the historical bool shape for direct callers.
func NewPlannerGate() func(context.Context, string) bool {
	return func(ctx context.Context, input string) bool {
		return DecidePlannerRoute(ctx, input).Route != agent.PlannerRouteExecutorOnly
	}
}
