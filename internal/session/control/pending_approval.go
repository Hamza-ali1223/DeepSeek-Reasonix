package control

import (
	"encoding/json"

	"reasonix/internal/contract/event"
)

// pendingApproval is one outstanding decision card.
type pendingApproval struct {
	id           string
	tool         string
	subject      string
	reason       string
	rawInput     json.RawMessage
	fresh        bool
	requireHuman bool
	autoDrain    bool
	kind         string // tool | plan | recovery; empty = tool
	recovery     *event.RecoveryApproval
	turn         turnStamp
	reply        chan approvalReply
}

// turnStamp is the in-flight turn a decision card was raised under. It is
// captured at registration because the resolver runs without the turn's context.
type turnStamp struct {
	id    string
	start int
}
