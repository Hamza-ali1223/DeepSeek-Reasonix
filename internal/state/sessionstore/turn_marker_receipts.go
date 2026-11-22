package sessionstore

import "reasonix/internal/contract/provider"

// A decision receipt recorded before its turn's user message lands has no home
// in the transcript. These keep it on the turn's in-flight marker, beside the
// transcript, until a message in that turn can carry it.

// HoldTurnReceipt appends receipt to the pending list of the turn identified by
// id. The identity is matched, never guessed, so a stale approval writes nowhere
// and a later turn's marker is never updated.
func HoldTurnReceipt(sessionPath, id string, receipt *provider.DecisionReceipt) (bool, error) {
	if sessionPath == "" || id == "" || receipt == nil {
		return false, nil
	}
	unlock, err := LockSessionMetaPath(sessionPath)
	if err != nil {
		return false, err
	}
	defer unlock()
	m, ok, err := LoadBranchMeta(sessionPath)
	if err != nil || !ok || m.InFlightTurn == nil || m.InFlightTurn.ID != id {
		return false, err
	}
	marker := *m.InFlightTurn
	marker.PendingReceipts = append(append([]*provider.DecisionReceipt(nil), marker.PendingReceipts...), receipt)
	m.InFlightTurn = &marker
	return true, saveBranchMeta(sessionPath, m, false)
}

// ClearPendingReceiptsIfDurable drops the pending receipts of the turn named by
// id. Called only once the transcript carrying them is durable, so a failed save
// leaves them for the next attempt.
func ClearPendingReceiptsIfDurable(sessionPath, id string) (bool, error) {
	if sessionPath == "" || id == "" {
		return false, nil
	}
	unlock, err := LockSessionMetaPath(sessionPath)
	if err != nil {
		return false, err
	}
	defer unlock()
	m, ok, err := LoadBranchMeta(sessionPath)
	if err != nil || !ok || m.InFlightTurn == nil || m.InFlightTurn.ID != id {
		return false, err
	}
	if len(m.InFlightTurn.PendingReceipts) == 0 {
		return false, nil
	}
	marker := *m.InFlightTurn
	marker.PendingReceipts = nil
	m.InFlightTurn = &marker
	return true, saveBranchMeta(sessionPath, m, false)
}

// PendingTurnReceipts returns a copy of the pending receipts on the turn named
// by id, so recovery can decide whether to preserve or drop them.
func PendingTurnReceipts(sessionPath, id string) []*provider.DecisionReceipt {
	if sessionPath == "" || id == "" {
		return nil
	}
	m, ok, err := LoadBranchMeta(sessionPath)
	if err != nil || !ok || m.InFlightTurn == nil || m.InFlightTurn.ID != id {
		return nil
	}
	return append([]*provider.DecisionReceipt(nil), m.InFlightTurn.PendingReceipts...)
}
