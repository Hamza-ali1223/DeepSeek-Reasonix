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
	held := make(map[string]bool, len(s.Messages[target].DecisionReceipts))
	for _, r := range s.Messages[target].DecisionReceipts {
		if r != nil {
			held[r.ID] = true
		}
	}
	existing := s.receiptIDsLocked()
	added := false
	for _, r := range receipts {
		if r == nil || existing[r.ID] || held[r.ID] {
			continue
		}
		held[r.ID] = true
		s.Messages[target].DecisionReceipts = append(s.Messages[target].DecisionReceipts, r)
		added = true
	}
	if added {
		s.rewriteVersion++
		s.version++
	}
	return true
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
