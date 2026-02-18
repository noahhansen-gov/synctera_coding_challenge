package model

import "time"

// Transaction represents a financial transaction.
type Transaction struct {
	ID          string            `json:"id"`
	Amount      int64             `json:"amount"`
	Currency    string            `json:"currency"`
	EffectiveAt time.Time         `json:"effective_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Clone returns a deep copy of the transaction.
// Metadata is a map (reference type), so it must be explicitly copied to
// prevent callers from mutating the store's internal state.
func (t Transaction) Clone() Transaction {
	c := t
	if t.Metadata != nil {
		c.Metadata = make(map[string]string, len(t.Metadata))
		for k, v := range t.Metadata {
			c.Metadata[k] = v
		}
	}
	return c
}

// Equal returns true if two transactions have identical field values.
// Used for idempotency checks.
func (t Transaction) Equal(other Transaction) bool {
	if t.ID != other.ID ||
		t.Amount != other.Amount ||
		t.Currency != other.Currency ||
		!t.EffectiveAt.Equal(other.EffectiveAt) {
		return false
	}

	if len(t.Metadata) != len(other.Metadata) {
		return false
	}
	for k, v := range t.Metadata {
		otherV, ok := other.Metadata[k]
		if !ok || otherV != v {
			return false
		}
	}
	return true
}
