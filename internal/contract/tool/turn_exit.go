// turn_exit.go — a tool whose call ends the turn instead of extending it.
package tool

// TurnExit is implemented by a tool whose call ends the turn with the result
// the turn was gathered for, rather than doing more work. A turn the host has
// told to stop using tools may still call one: closing research while refusing
// the exit strands exactly the evidence the research was spent collecting.
type TurnExit interface {
	ExitsTurn() bool
}

// ExitsTurn reports whether this call finishes the turn instead of extending it.
func ExitsTurn(t Tool) bool {
	e, ok := t.(TurnExit)
	return ok && e.ExitsTurn()
}
