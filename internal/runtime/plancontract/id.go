package plancontract

import "crypto/rand"

// NewID creates a host-owned identity for a plan that has no predecessor.
func NewID() string { return "plan-" + rand.Text() }
