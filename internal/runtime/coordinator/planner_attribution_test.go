package coordinator

import (
	"testing"

	"reasonix/internal/contract/event"
	"reasonix/internal/contract/eventwire"
)

func TestPlannerStreamNamesItsModelAcrossTheWire(t *testing.T) {
	for _, kind := range []event.Kind{event.Text, event.Message, event.Reasoning} {
		var frame eventwire.Event
		sink := plannerSink(event.FuncSink(func(e event.Event) { frame = eventwire.ToWire(e) }), "gateway/sol")
		sink.Emit(event.Event{Kind: kind})
		if frame.ModelRef != "gateway/sol" || frame.Source != event.UsageSourcePlanner {
			t.Errorf("%v: model = %q, source = %q", kind, frame.ModelRef, frame.Source)
		}
	}
}
