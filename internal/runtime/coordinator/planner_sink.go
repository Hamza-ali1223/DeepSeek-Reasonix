package coordinator

import (
	"strings"

	"reasonix/internal/base/nilutil"
	"reasonix/internal/contract/event"
)

// plannerEventSink relabels the planner's stream and hides its turn
// boundaries. A type, not a FuncSink: audits are delivered by type assertion,
// and a func sink implements none of them.
type plannerEventSink struct {
	event.AuditForwarder
	inner    event.Sink
	modelRef string
}

func (s plannerEventSink) Emit(e event.Event) {
	switch e.Kind {
	case event.TurnStarted, event.TurnDone:
		return
	default:
		if e.Source == "" {
			e.Source = event.UsageSourcePlanner
		}
		if e.ModelRef == "" {
			e.ModelRef = s.modelRef
		}
		s.inner.Emit(e)
	}
}

func plannerSink(sink event.Sink, modelRef string) event.Sink {
	if nilutil.IsNil(sink) {
		sink = event.Discard
	}
	return plannerEventSink{AuditForwarder: event.AuditForwarder{Inner: sink}, inner: sink, modelRef: strings.TrimSpace(modelRef)}
}
