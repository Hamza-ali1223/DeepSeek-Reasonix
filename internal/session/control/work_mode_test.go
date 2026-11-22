package control

import (
	"context"
	"errors"
	"testing"
	"time"

	"reasonix/internal/contract/event"
)

type workModeRunner struct{ calls int }

func (r *workModeRunner) Run(context.Context, string) error {
	r.calls++
	return nil
}

func TestAgentPlannerWithoutDedicatedPlannerFailsWithIdentity(t *testing.T) {
	done := make(chan event.Event, 1)
	runner := &workModeRunner{}
	c := New(Options{Runner: runner, Sink: event.FuncSink(func(e event.Event) {
		if e.Kind == event.TurnDone {
			done <- e
		}
	})})
	defer c.Close()

	if got := c.SubmitHTTPOptions("implement the feature", SubmitOptions{WorkMode: WorkModeAgentPlanner}); got.Refused() {
		t.Fatalf("submission admission = %v", got)
	}
	select {
	case e := <-done:
		if !errors.Is(e.Err, ErrPlannerUnavailable) {
			t.Fatalf("turn error = %v, want ErrPlannerUnavailable", e.Err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("turn did not finish")
	}
	if runner.calls != 0 {
		t.Fatalf("executor calls = %d, want none", runner.calls)
	}
}
