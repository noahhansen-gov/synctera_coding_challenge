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
		if other.Metadata[k] != v {
			return false
		}
	}
	return true
}
