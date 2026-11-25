// planner_budget.go — the two numbers a planning turn is paced by.
package config

import (
	"errors"
	"fmt"
)

const (
	// DefaultPlannerResearchRounds is what a planning turn gets unasked.
	DefaultPlannerResearchRounds = 12
	// MinPlannerResearchRounds is the smallest cap that still leaves a turn able
	// to research at all.
	MinPlannerResearchRounds = 1
	// MaxPlannerResearchRounds is the ceiling. No value above it is accepted,
	// because each round re-sends a growing prompt.
	MaxPlannerResearchRounds = 12

	DefaultConsultationTimeoutSeconds = 90
	MinConsultationTimeoutSeconds     = 15
	MaxConsultationTimeoutSeconds     = 600

	// MaxPlannerConsultationLimit bounds consultations in one executor turn.
	MaxPlannerConsultationLimit = 20
)

var (
	ErrPlannerResearchRoundsOutOfRange    = errors.New("planner research rounds out of range")
	ErrConsultationTimeoutOutOfRange      = errors.New("planner consultation timeout out of range")
	ErrPlannerConsultationLimitOutOfRange = errors.New("planner consultation limit out of range")
)

// PlannerResearchRounds is the effective cap. Zero and any value outside the
// accepted range read as the default: a stored value the loader cannot trust
// must never reach the loop as unlimited.
func (c *Config) PlannerResearchRounds() int {
	if c == nil {
		return DefaultPlannerResearchRounds
	}
	n := c.Agent.PlannerResearchRounds
	if n < MinPlannerResearchRounds || n > MaxPlannerResearchRounds {
		return DefaultPlannerResearchRounds
	}
	return n
}

// PlannerAimByRound is the round the planner is asked to finish by: two below
// the cap, so the last two stay available for an unresolved question. A cap
// below three has no room for a reserve, so there is no target.
func (c *Config) PlannerAimByRound() int {
	n := c.PlannerResearchRounds()
	if n < 3 {
		return 0
	}
	return n - 2
}

// SetPlannerResearchRounds validates a save. Zero is refused rather than read as
// the default here, because a person who typed it meant something and should be
// told what is accepted.
func (c *Config) SetPlannerResearchRounds(n int) error {
	if n < MinPlannerResearchRounds || n > MaxPlannerResearchRounds {
		return fmt.Errorf("%w: use %d through %d", ErrPlannerResearchRoundsOutOfRange, MinPlannerResearchRounds, MaxPlannerResearchRounds)
	}
	c.Agent.PlannerResearchRounds = n
	return nil
}

// ConsultationTimeout is the effective deadline for one consultation.
func (c *Config) ConsultationTimeout() int {
	if c == nil {
		return DefaultConsultationTimeoutSeconds
	}
	n := c.Agent.PlannerConsultationTimeoutSeconds
	if n < MinConsultationTimeoutSeconds || n > MaxConsultationTimeoutSeconds {
		return DefaultConsultationTimeoutSeconds
	}
	return n
}

// SetConsultationTimeout validates a save in seconds.
func (c *Config) SetConsultationTimeout(seconds int) error {
	if seconds < MinConsultationTimeoutSeconds || seconds > MaxConsultationTimeoutSeconds {
		return fmt.Errorf("%w: use %d through %d seconds", ErrConsultationTimeoutOutOfRange, MinConsultationTimeoutSeconds, MaxConsultationTimeoutSeconds)
	}
	c.Agent.PlannerConsultationTimeoutSeconds = seconds
	return nil
}

// SetPlannerConsultationLimit validates the per-turn consultation count.
func (c *Config) SetPlannerConsultationLimit(n int) error {
	if n < 1 || n > MaxPlannerConsultationLimit {
		return fmt.Errorf("%w: use 1 through %d", ErrPlannerConsultationLimitOutOfRange, MaxPlannerConsultationLimit)
	}
	c.Agent.PlannerConsultationLimit = n
	return nil
}
