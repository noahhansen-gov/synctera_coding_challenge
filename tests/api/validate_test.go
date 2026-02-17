package api_test

import (
	"testing"
	"time"

	"github.com/synctera/tech-challenge/internal/api"
	"github.com/synctera/tech-challenge/internal/model"
)

// --- ValidateTransaction ---

// Test: TestValidateTransaction_valid
// What: ValidateTransaction accepts a fully populated transaction
// Input: Transaction with all required fields set (id, currency, positive amount, effective_at)
// Output: nil error
func TestValidateTransaction_valid(t *testing.T) {
	txn := model.Transaction{
		ID:          "txn-1",
		Amount:      100,
		Currency:    "USD",
		EffectiveAt: time.Now(),
	}
	if err := api.ValidateTransaction(txn); err != nil {
		t.Errorf("expected nil error for valid transaction, got %v", err)
	}
}

// Test: TestValidateTransaction_missingID
// What: ValidateTransaction rejects a transaction with no ID
// Input: Transaction with empty ID field, all other fields valid
// Output: non-nil error
func TestValidateTransaction_missingID(t *testing.T) {
	txn := model.Transaction{Amount: 100, Currency: "USD", EffectiveAt: time.Now()}
	if err := api.ValidateTransaction(txn); err == nil {
		t.Error("expected error for missing ID, got nil")
	}
}

// Test: TestValidateTransaction_missingCurrency
// What: ValidateTransaction rejects a transaction with no currency
// Input: Transaction with empty Currency field, all other fields valid
// Output: non-nil error
func TestValidateTransaction_missingCurrency(t *testing.T) {
	txn := model.Transaction{ID: "txn-1", Amount: 100, EffectiveAt: time.Now()}
	if err := api.ValidateTransaction(txn); err == nil {
		t.Error("expected error for missing currency, got nil")
	}
}

// Test: TestValidateTransaction_missingEffectiveAt
// What: ValidateTransaction rejects a transaction with a zero-value EffectiveAt
// Input: Transaction with EffectiveAt unset (zero time.Time), all other fields valid
// Output: non-nil error
func TestValidateTransaction_missingEffectiveAt(t *testing.T) {
	txn := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD"}
	if err := api.ValidateTransaction(txn); err == nil {
		t.Error("expected error for missing effective_at, got nil")
	}
}

// Test: TestValidateTransaction_negativeAmount
// What: ValidateTransaction rejects a transaction with a negative amount
// Input: Transaction with Amount = -1, all other fields valid
// Output: non-nil error
func TestValidateTransaction_negativeAmount(t *testing.T) {
	txn := model.Transaction{ID: "txn-1", Amount: -1, Currency: "USD", EffectiveAt: time.Now()}
	if err := api.ValidateTransaction(txn); err == nil {
		t.Error("expected error for negative amount, got nil")
	}
}

// Test: TestValidateTransaction_zeroAmountAllowed
// What: ValidateTransaction permits amount = 0 (zero-value transactions are valid)
// Input: Transaction with Amount = 0, all other fields valid
// Output: nil error
func TestValidateTransaction_zeroAmountAllowed(t *testing.T) {
	txn := model.Transaction{ID: "txn-1", Amount: 0, Currency: "USD", EffectiveAt: time.Now()}
	if err := api.ValidateTransaction(txn); err != nil {
		t.Errorf("expected nil error for zero amount, got %v", err)
	}
}

// --- ValidatePagination ---

// Test: TestValidatePagination_validDefaults
// What: ValidatePagination accepts the default limit=100 and offset=0
// Input: limit=100, offset=0
// Output: nil error
func TestValidatePagination_validDefaults(t *testing.T) {
	if err := api.ValidatePagination(100, 0); err != nil {
		t.Errorf("expected nil for default pagination, got %v", err)
	}
}

// Test: TestValidatePagination_limitOne
// What: ValidatePagination accepts the minimum valid limit of 1
// Input: limit=1, offset=0
// Output: nil error
func TestValidatePagination_limitOne(t *testing.T) {
	if err := api.ValidatePagination(1, 0); err != nil {
		t.Errorf("expected nil for limit=1, got %v", err)
	}
}

// Test: TestValidatePagination_limitMax
// What: ValidatePagination accepts the maximum valid limit of 1000
// Input: limit=1000, offset=0
// Output: nil error
func TestValidatePagination_limitMax(t *testing.T) {
	if err := api.ValidatePagination(1000, 0); err != nil {
		t.Errorf("expected nil for limit=1000, got %v", err)
	}
}

// Test: TestValidatePagination_zeroLimit
// What: ValidatePagination rejects limit=0 (must be at least 1)
// Input: limit=0, offset=0
// Output: non-nil error
func TestValidatePagination_zeroLimit(t *testing.T) {
	if err := api.ValidatePagination(0, 0); err == nil {
		t.Error("expected error for limit=0, got nil")
	}
}

// Test: TestValidatePagination_negativeLimit
// What: ValidatePagination rejects a negative limit
// Input: limit=-1, offset=0
// Output: non-nil error
func TestValidatePagination_negativeLimit(t *testing.T) {
	if err := api.ValidatePagination(-1, 0); err == nil {
		t.Error("expected error for limit=-1, got nil")
	}
}

// Test: TestValidatePagination_limitTooHigh
// What: ValidatePagination rejects limit > 1000 to cap response size
// Input: limit=1001, offset=0
// Output: non-nil error
func TestValidatePagination_limitTooHigh(t *testing.T) {
	if err := api.ValidatePagination(1001, 0); err == nil {
		t.Error("expected error for limit=1001, got nil")
	}
}

// Test: TestValidatePagination_negativeOffset
// What: ValidatePagination rejects a negative offset
// Input: limit=10, offset=-1
// Output: non-nil error
func TestValidatePagination_negativeOffset(t *testing.T) {
	if err := api.ValidatePagination(10, -1); err == nil {
		t.Error("expected error for offset=-1, got nil")
	}
}
