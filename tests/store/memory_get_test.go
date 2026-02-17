package store_test

import (
	"errors"
	"testing"

	"github.com/synctera/tech-challenge/internal/store"
)

// Test: TestGet_existingID
// What: Get retrieves a transaction by its ID after it has been stored
// Input: store with one transaction (id="txn-1"), lookup by "txn-1"
// Output: the stored transaction, nil error
func TestGet_existingID(t *testing.T) {
	s := store.NewMemoryStore()
	txn := makeTxn("txn-1", 500, "USD", jan(1))
	_ = s.Create(txn)

	got, err := s.Get("txn-1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !got.Equal(txn) {
		t.Errorf("retrieved transaction does not match stored: got %+v", got)
	}
}

// Test: TestGet_missingID
// What: Get returns ErrNotFound when the requested ID does not exist in the store
// Input: empty store, lookup by "nonexistent"
// Output: ErrNotFound
func TestGet_missingID(t *testing.T) {
	s := store.NewMemoryStore()

	_, err := s.Get("nonexistent")
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// Test: TestGet_returnsCorrectFieldValues
// What: Get returns a transaction with all fields intact
// Input: store with transaction (id="txn-42", amount=1234, currency="EUR", effectiveAt=jan(15))
// Output: transaction with identical field values
func TestGet_returnsCorrectFieldValues(t *testing.T) {
	s := store.NewMemoryStore()
	txn := makeTxn("txn-42", 1234, "EUR", jan(15))
	_ = s.Create(txn)

	got, _ := s.Get("txn-42")

	if got.ID != "txn-42" {
		t.Errorf("ID: expected %q, got %q", "txn-42", got.ID)
	}
	if got.Amount != 1234 {
		t.Errorf("Amount: expected 1234, got %d", got.Amount)
	}
	if got.Currency != "EUR" {
		t.Errorf("Currency: expected EUR, got %q", got.Currency)
	}
	if !got.EffectiveAt.Equal(jan(15)) {
		t.Errorf("EffectiveAt: expected %v, got %v", jan(15), got.EffectiveAt)
	}
}

// Test: TestGet_afterMultipleCreates
// What: Get correctly identifies a specific transaction among several stored ones
// Input: store with 3 transactions (USD, EUR, GBP), lookup by "txn-2"
// Output: the EUR transaction, nil error
func TestGet_afterMultipleCreates(t *testing.T) {
	s := store.NewMemoryStore()
	_ = s.Create(makeTxn("txn-1", 100, "USD", jan(1)))
	_ = s.Create(makeTxn("txn-2", 200, "EUR", jan(2)))
	_ = s.Create(makeTxn("txn-3", 300, "GBP", jan(3)))

	got, err := s.Get("txn-2")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Currency != "EUR" {
		t.Errorf("expected EUR, got %q", got.Currency)
	}
}
