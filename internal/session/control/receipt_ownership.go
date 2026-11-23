package control

import (
	"context"
	"log/slog"

	"reasonix/internal/contract/provider"
	"reasonix/internal/state/sessionstore"
)

// Decision receipts belong to a turn, not to the transcript's current tail: a
// plan card can be answered before the turn's user message lands, so the receipt
// waits on the turn's own marker until a message in that turn can carry it.

type receiptTurnKey struct{}

func (c *Controller) currentTurnStamp(ctx context.Context) turnStamp {
	stamp, _ := ctx.Value(receiptTurnKey{}).(turnStamp)
	return stamp
}

// holdTurnReceipt keeps a receipt out of the transcript until its turn can carry
// it. A durable marker holds it beside the transcript; without one it is held in
// memory under the same turn, because an index was announced either way.
func (c *Controller) holdTurnReceipt(ctx context.Context, turn turnStamp, receipt *provider.DecisionReceipt) {
	if c == nil || c.executor == nil || receipt == nil {
		return
	}
	if turn.id == "" {
		c.executor.Session().AttachTurnReceipts(turn.start, []*provider.DecisionReceipt{receipt})
		return
	}
	if path := c.SessionPath(); path != "" {
		held, err := sessionstore.HoldTurnReceipt(path, turn.id, receipt)
		if err == nil && held {
			return
		}
		if err != nil {
			slog.Warn("controller: persist pending receipt", "err", err)
		}
	}
	c.keepPendingReceipt(turn.id, receipt)
}

func (c *Controller) keepPendingReceipt(id string, receipts ...*provider.DecisionReceipt) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.pendingReceipts == nil {
		c.pendingReceipts = map[string][]*provider.DecisionReceipt{}
	}
	c.pendingReceipts[id] = append(c.pendingReceipts[id], receipts...)
}

// takePendingReceipts merges a turn's pending receipts and forgets the in-memory
// copy. The marker is read from the sidecar because the caller's value may
// predate the hold; reading that stale copy would see nothing and lose the only one.
func (c *Controller) takePendingReceipts(turn sessionstore.InFlightTurnMeta) []*provider.DecisionReceipt {
	out := append([]*provider.DecisionReceipt(nil), turn.PendingReceipts...)
	if turn.ID == "" {
		return out
	}
	if path := c.SessionPath(); path != "" {
		out = append(out, sessionstore.PendingTurnReceipts(path, turn.ID)...)
	}
	c.mu.Lock()
	out = append(out, c.pendingReceipts[turn.ID]...)
	delete(c.pendingReceipts, turn.ID)
	c.mu.Unlock()
	return out
}

// dropTurnReceipts forgets receipts for a turn abandoned before it landed.
func (c *Controller) dropTurnReceipts(id string) {
	if c == nil || id == "" {
		return
	}
	c.mu.Lock()
	delete(c.pendingReceipts, id)
	c.mu.Unlock()
}

// settleTurnReceipts uses the turn's reply or, if absent, its retained request.
// Without a carrier, the receipts remain pending.
func (c *Controller) settleTurnReceipts(turn sessionstore.InFlightTurnMeta) bool {
	if c == nil || c.executor == nil {
		return true
	}
	receipts := c.takePendingReceipts(turn)
	if len(receipts) == 0 {
		return true
	}
	sess := c.executor.Session()
	if turn.PreserveUser {
		msgs := sess.Snapshot()
		start, found := resolveInterruptedTurnStart(msgs, turn.StartMessageIndex, true, turn.StartedAt, provider.Message{})
		if found && !interruptedTurnCrossesLaterTurn(msgs, start) &&
			(sess.AttachTurnReceipts(start, receipts) || sess.AttachRecoveredTurnReceipts(start, receipts)) {
			return true
		}
	} else if sess.AttachTurnReceipts(turn.StartMessageIndex, receipts) {
		return true
	}
	c.keepPendingReceipt(turn.ID, receipts...)
	return false
}

func (c *Controller) recoverCommittedTurnReceipts(turn sessionstore.InFlightTurnMeta) bool {
	if len(turn.PendingReceipts) == 0 {
		return true
	}
	sess := c.executor.Session()
	before := sess.RewriteVersion()
	if !c.settleTurnReceipts(turn) {
		return false
	}
	if sess.RewriteVersion() == before {
		return true
	}
	durable, err := c.guardedSnapshotWithDurability(false, true, false)
	if err != nil {
		slog.Warn("controller: save committed turn receipts", "err", err)
	}
	return durable
}
