package sessionstore

import "reasonix/internal/contract/provider"

// Turn-owned receipt attachment. A receipt recorded before its turn's user
// message lands has no home in the transcript: appending it would take the
// message index the host announced for that user message.

// AttachTurnReceipts attaches receipts to the newest non-local assistant message
// at or after floor — the first message that can own them in that turn. It never
// looks below floor, so an approval recorded early cannot land on the previous
// turn's reply. A receipt already anywhere in the transcript is skipped, so a
// retry yields one copy. Reports whether every receipt is now owned.
func (s *Session) AttachTurnReceipts(floor int, receipts []*provider.DecisionReceipt) bool {
	if s == nil || len(receipts) == 0 {
		return true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	target := -1
	for i := max(floor, 0); i < len(s.Messages); i++ {
		if s.Messages[i].Role == provider.RoleAssistant && !s.Messages[i].LocalOnly {
			target = i
		}
	}
	if target < 0 {
		// Never append here: the announced slot may still be empty.
		return false
	}
	s.attachReceiptsLocked(target, receipts)
	return true
}

// AttachRecoveredTurnReceipts uses the recovered user message as the carrier
// because recovery can discard the unfinished assistant reply.
func (s *Session) AttachRecoveredTurnReceipts(userIndex int, receipts []*provider.DecisionReceipt) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if userIndex < 0 || userIndex >= len(s.Messages) {
		return false
	}
	user := s.Messages[userIndex]
	if user.Role != provider.RoleUser || user.LocalOnly {
		return false
	}
	s.attachReceiptsLocked(userIndex, receipts)
	return true
}

func (s *Session) attachReceiptsLocked(target int, receipts []*provider.DecisionReceipt) {
	existing := s.receiptIDsLocked()
	var added []*provider.DecisionReceipt
	for _, r := range receipts {
		if r == nil || existing[r.ID] {
			continue
		}
		existing[r.ID] = true
		added = append(added, r)
	}
	if len(added) > 0 {
		current := s.Messages[target].DecisionReceipts
		s.Messages[target].DecisionReceipts = append(current[:len(current):len(current)], added...)
		s.rewriteVersion++
		s.version++
	}
}

// receiptIDsLocked collects every receipt id in the transcript, so a retried
// attach cannot duplicate one a previous save made durable.
func (s *Session) receiptIDsLocked() map[string]bool {
	ids := map[string]bool{}
	for i := range s.Messages {
		for _, r := range s.Messages[i].DecisionReceipts {
			if r != nil {
				ids[r.ID] = true
			}
		}
		if r := s.Messages[i].DecisionReceipt; r != nil {
			ids[r.ID] = true
		}
	}
	return ids
}
