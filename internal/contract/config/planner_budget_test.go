package config

import (
	"errors"
	"testing"
)

func TestPlannerResearchRoundsReadsTheDefaultWhenNoValueIsSet(t *testing.T) {
	var c Config
	if got := c.PlannerResearchRounds(); got != DefaultPlannerResearchRounds {
		t.Fatalf("rounds = %d, want the default %d", got, DefaultPlannerResearchRounds)
	}
}

// A stored value the loader cannot trust must never reach the loop, where zero
// would mean no bound at all.
func TestPlannerResearchRoundsNeverReadsAsUnbounded(t *testing.T) {
	for _, stored := range []int{0, -1, MaxPlannerResearchRounds + 1, 1000} {
		c := Config{Agent: AgentConfig{PlannerResearchRounds: stored}}
		got := c.PlannerResearchRounds()
		if got < MinPlannerResearchRounds || got > MaxPlannerResearchRounds {
			t.Fatalf("stored %d read as %d, outside the enforced range", stored, got)
		}
	}
}

func TestSetPlannerResearchRoundsRefusesZeroAndOutOfRange(t *testing.T) {
	var c Config
	for _, want := range []int{0, -1, MaxPlannerResearchRounds + 1} {
		if err := c.SetPlannerResearchRounds(want); !errors.Is(err, ErrPlannerResearchRoundsOutOfRange) {
			t.Fatalf("SetPlannerResearchRounds(%d) = %v, want the range error", want, err)
		}
	}
	if err := c.SetPlannerResearchRounds(7); err != nil {
		t.Fatalf("SetPlannerResearchRounds(7): %v", err)
	}
	if c.Agent.PlannerResearchRounds != 7 {
		t.Fatalf("stored = %d, want 7", c.Agent.PlannerResearchRounds)
	}
}

func TestPlannerAimByRoundReservesTheLastTwo(t *testing.T) {
	for _, tc := range []struct{ rounds, want int }{
		{12, 10}, {3, 1}, {2, 0}, {1, 0},
	} {
		c := Config{Agent: AgentConfig{PlannerResearchRounds: tc.rounds}}
		if got := c.PlannerAimByRound(); got != tc.want {
			t.Fatalf("rounds %d aims at %d, want %d", tc.rounds, got, tc.want)
		}
	}
}

func TestConsultationTimeoutReadsTheDefaultAndTheRange(t *testing.T) {
	var c Config
	if got := c.ConsultationTimeout(); got != DefaultConsultationTimeoutSeconds {
		t.Fatalf("timeout = %d, want %d", got, DefaultConsultationTimeoutSeconds)
	}
	c.Agent.PlannerConsultationTimeoutSeconds = 240
	if got := c.ConsultationTimeout(); got != 240 {
		t.Fatalf("timeout = %d, want 240", got)
	}
	c.Agent.PlannerConsultationTimeoutSeconds = 0
	if got := c.ConsultationTimeout(); got != DefaultConsultationTimeoutSeconds {
		t.Fatalf("a zero timeout read as %d, want the default", got)
	}
}

func TestSetConsultationTimeoutRefusesAnOutOfRangeWait(t *testing.T) {
	var c Config
	for _, want := range []int{0, MinConsultationTimeoutSeconds - 1, MaxConsultationTimeoutSeconds + 1} {
		if err := c.SetConsultationTimeout(want); !errors.Is(err, ErrConsultationTimeoutOutOfRange) {
			t.Fatalf("SetConsultationTimeout(%d) = %v, want the range error", want, err)
		}
	}
	if err := c.SetConsultationTimeout(120); err != nil {
		t.Fatalf("SetConsultationTimeout(120): %v", err)
	}
}

func TestSetPlannerConsultationLimitAcceptsUpToTwenty(t *testing.T) {
	var c Config
	if err := c.SetPlannerConsultationLimit(20); err != nil {
		t.Fatalf("limit 20 must be accepted: %v", err)
	}
	for _, want := range []int{0, -1, MaxPlannerConsultationLimit + 1} {
		if err := c.SetPlannerConsultationLimit(want); !errors.Is(err, ErrPlannerConsultationLimitOutOfRange) {
			t.Fatalf("SetPlannerConsultationLimit(%d) = %v, want the range error", want, err)
		}
	}
}
