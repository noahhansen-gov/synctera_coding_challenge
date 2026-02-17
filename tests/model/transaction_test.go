package model_test

import (
	"testing"
	"time"

	"github.com/synctera/tech-challenge/internal/model"
)

var t0 = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

// Test: TestEqual_identical
// What: Transaction.Equal returns true when compared with itself
// Input: same Transaction value passed as both receiver and argument
// Output: true
func TestEqual_identical(t *testing.T) {
	txn := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0}
	if !txn.Equal(txn) {
		t.Fatal("identical transaction should equal itself")
	}
}

// Test: TestEqual_differentID
// What: Transaction.Equal returns false when IDs differ
// Input: two transactions identical except ID ("txn-1" vs "txn-2")
// Output: false
func TestEqual_differentID(t *testing.T) {
	a := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0}
	b := model.Transaction{ID: "txn-2", Amount: 100, Currency: "USD", EffectiveAt: t0}
	if a.Equal(b) {
		t.Fatal("transactions with different IDs should not be equal")
	}
}

// Test: TestEqual_differentAmount
// What: Transaction.Equal returns false when amounts differ
// Input: two transactions identical except Amount (100 vs 200)
// Output: false
func TestEqual_differentAmount(t *testing.T) {
	a := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0}
	b := model.Transaction{ID: "txn-1", Amount: 200, Currency: "USD", EffectiveAt: t0}
	if a.Equal(b) {
		t.Fatal("transactions with different amounts should not be equal")
	}
}

// Test: TestEqual_differentCurrency
// What: Transaction.Equal returns false when currencies differ
// Input: two transactions identical except Currency ("USD" vs "EUR")
// Output: false
func TestEqual_differentCurrency(t *testing.T) {
	a := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0}
	b := model.Transaction{ID: "txn-1", Amount: 100, Currency: "EUR", EffectiveAt: t0}
	if a.Equal(b) {
		t.Fatal("transactions with different currencies should not be equal")
	}
}

// Test: TestEqual_differentEffectiveAt
// What: Transaction.Equal returns false when timestamps differ
// Input: two transactions identical except EffectiveAt (t0 vs t0+1h)
// Output: false
func TestEqual_differentEffectiveAt(t *testing.T) {
	a := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0}
	b := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0.Add(time.Hour)}
	if a.Equal(b) {
		t.Fatal("transactions with different effective_at should not be equal")
	}
}

// Test: TestEqual_identicalMetadata
// What: Transaction.Equal returns true when both metadata maps have the same key-value pairs
// Input: two transactions with Metadata={"key":"val"} each
// Output: true
func TestEqual_identicalMetadata(t *testing.T) {
	meta := map[string]string{"key": "val"}
	a := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0, Metadata: meta}
	b := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0, Metadata: map[string]string{"key": "val"}}
	if !a.Equal(b) {
		t.Fatal("transactions with identical metadata should be equal")
	}
}

// Test: TestEqual_differentMetadataValue
// What: Transaction.Equal returns false when a metadata key has a different value
// Input: two transactions with Metadata={"key":"val-a"} and {"key":"val-b"}
// Output: false
func TestEqual_differentMetadataValue(t *testing.T) {
	a := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0, Metadata: map[string]string{"key": "val-a"}}
	b := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0, Metadata: map[string]string{"key": "val-b"}}
	if a.Equal(b) {
		t.Fatal("transactions with different metadata values should not be equal")
	}
}

// Test: TestEqual_extraMetadataKey
// What: Transaction.Equal returns false when one transaction has more metadata keys than the other
// Input: a has {"key":"val","extra":"x"}, b has {"key":"val"}
// Output: false
func TestEqual_extraMetadataKey(t *testing.T) {
	a := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0, Metadata: map[string]string{"key": "val", "extra": "x"}}
	b := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0, Metadata: map[string]string{"key": "val"}}
	if a.Equal(b) {
		t.Fatal("transactions with different metadata key counts should not be equal")
	}
}

// Test: TestEqual_nilMetadataEqualsEmptyMetadata
// What: Transaction.Equal treats nil Metadata and an empty map as equivalent
// Input: a.Metadata=nil, b.Metadata=map[string]string{}
// Output: true
func TestEqual_nilMetadataEqualsEmptyMetadata(t *testing.T) {
	a := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0, Metadata: nil}
	b := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0, Metadata: map[string]string{}}
	if !a.Equal(b) {
		t.Fatal("nil metadata should equal empty metadata map")
	}
}

// Test: TestEqual_bothNilMetadata
// What: Transaction.Equal returns true when both transactions have nil Metadata
// Input: two transactions with no Metadata field set
// Output: true
func TestEqual_bothNilMetadata(t *testing.T) {
	a := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0}
	b := model.Transaction{ID: "txn-1", Amount: 100, Currency: "USD", EffectiveAt: t0}
	if !a.Equal(b) {
		t.Fatal("transactions with both nil metadata should be equal")
	}
}
