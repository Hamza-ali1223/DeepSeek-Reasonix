package control

import (
	"context"
	"errors"
)

// WorkMode is the explicit per-submission work selection carried by the host.
type WorkMode string

const (
	WorkModeLegacy WorkMode = ""
	// WorkModeAgentPlanner opts this turn into the dedicated planner workflow.
	WorkModeAgentPlanner WorkMode = "agent_planner"
)

var (
	// ErrUnsupportedWorkMode identifies a work mode the host does not accept.
	ErrUnsupportedWorkMode = errors.New("unsupported work mode")
	// ErrPlannerUnavailable identifies a requested but unavailable dedicated planner.
	ErrPlannerUnavailable = errors.New("dedicated planner is unavailable. Configure a planner and retry, or select Agent and resubmit to run without it")
)

func (m WorkMode) Valid() bool {
	return m == WorkModeLegacy || m == WorkModeAgentPlanner
}

type workModeContextKey struct{}

func workModeFromContext(ctx context.Context) WorkMode {
	if ctx == nil {
		return WorkModeLegacy
	}
	mode, _ := ctx.Value(workModeContextKey{}).(WorkMode)
	return mode
}
