package store_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/synctera/tech-challenge/internal/store"
)

// Test: TestCreate_newTransaction
// What: Create stores a new transaction and returns nil when the ID is not yet in the store
// Input: MemoryStore with no data, one valid transaction
// Output: nil error
func TestCreate_newTransaction(t *testing.T) {
	s := store.NewMemoryStore()
	txn := makeTxn("txn-1", 100, "USD", jan(1))

	err := s.Create(txn)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

// Test: TestCreate_returnsDuplicateForIdenticalPayload
// What: Create returns ErrDuplicate when the same transaction is submitted twice (idempotent retry)
// Input: same transaction created twice (identical ID and payload)
// Output: ErrDuplicate on the second call
func TestCreate_returnsDuplicateForIdenticalPayload(t *testing.T) {
	s := store.NewMemoryStore()
	txn := makeTxn("txn-1", 100, "USD", jan(1))

	_ = s.Create(txn)
	err := s.Create(txn)

	if !errors.Is(err, store.ErrDuplicate) {
		t.Fatalf("expected ErrDuplicate, got %v", err)
	}
}

// Test: TestCreate_returnsConflictForDifferentPayload
// What: Create returns ErrConflict when the same ID is submitted with a different amount
// Input: original (id="txn-1", amount=100), then conflicting (id="txn-1", amount=999)
// Output: ErrConflict on the second call
func TestCreate_returnsConflictForDifferentPayload(t *testing.T) {
	s := store.NewMemoryStore()
	original := makeTxn("txn-1", 100, "USD", jan(1))
	modified := makeTxn("txn-1", 999, "USD", jan(1))

	_ = s.Create(original)
	err := s.Create(modified)

	if !errors.Is(err, store.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

// Test: TestCreate_conflictOnDifferentCurrency
// What: Create returns ErrConflict when the same ID is resubmitted with a different currency
// Input: original (id="txn-1", currency="USD"), then conflicting (id="txn-1", currency="EUR")
// Output: ErrConflict
func TestCreate_conflictOnDifferentCurrency(t *testing.T) {
	s := store.NewMemoryStore()
	_ = s.Create(makeTxn("txn-1", 100, "USD", jan(1)))
	err := s.Create(makeTxn("txn-1", 100, "EUR", jan(1)))

	if !errors.Is(err, store.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

// Test: TestCreate_conflictOnDifferentEffectiveAt
// What: Create returns ErrConflict when the same ID is resubmitted with a different effective_at
// Input: original (id="txn-1", effectiveAt=jan(1)), then conflicting (id="txn-1", effectiveAt=jan(2))
// Output: ErrConflict
func TestCreate_conflictOnDifferentEffectiveAt(t *testing.T) {
	s := store.NewMemoryStore()
	_ = s.Create(makeTxn("txn-1", 100, "USD", jan(1)))
	err := s.Create(makeTxn("txn-1", 100, "USD", jan(2)))

	if !errors.Is(err, store.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

// Test: TestCreate_maintainsSortedOrderByEffectiveAt
// What: Create inserts transactions into the ordered slice in chronological order regardless of insertion order
// Input: transactions created out of order (Jan 3, Jan 1, Jan 2)
// Output: List returns [a(Jan 1), b(Jan 2), c(Jan 3)]
func TestCreate_maintainsSortedOrderByEffectiveAt(t *testing.T) {
	s := store.NewMemoryStore()

	_ = s.Create(makeTxn("c", 100, "USD", jan(3)))
	_ = s.Create(makeTxn("a", 100, "USD", jan(1)))
	_ = s.Create(makeTxn("b", 100, "USD", jan(2)))

	list, _ := s.List(10, 0)
	if len(list) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list))
	}

	expected := []string{"a", "b", "c"}
	for i, txn := range list {
		if txn.ID != expected[i] {
			t.Errorf("index %d: expected ID %q, got %q", i, expected[i], txn.ID)
		}
	}
}

// Test: TestCreate_sameTimestampSortedByID
// What: Create sorts transactions with identical effective_at alphabetically by ID (tie-breaker)
// Input: transactions with the same timestamp inserted in order: zzz, aaa, mmm
// Output: List returns [aaa, mmm, zzz]
func TestCreate_sameTimestampSortedByID(t *testing.T) {
	s := store.NewMemoryStore()
	ts := jan(1)

	_ = s.Create(makeTxn("zzz", 100, "USD", ts))
	_ = s.Create(makeTxn("aaa", 100, "USD", ts))
	_ = s.Create(makeTxn("mmm", 100, "USD", ts))

	list, _ := s.List(10, 0)
	expected := []string{"aaa", "mmm", "zzz"}
	for i, txn := range list {
		if txn.ID != expected[i] {
			t.Errorf("index %d: expected ID %q, got %q", i, expected[i], txn.ID)
		}
	}
}

// Test: TestCreate_concurrent
// What: Create is safe for concurrent use — all goroutines complete without data races or lost writes
// Input: 100 goroutines each creating a unique transaction simultaneously
// Output: all 100 transactions stored, no race detector errors
func TestCreate_concurrent(t *testing.T) {
	s := store.NewMemoryStore()
	n := 100
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			txn := makeTxn(
				fmt.Sprintf("txn-%03d", i),
				int64(i*100),
				"USD",
				time.Date(2024, 1, i%28+1, 0, 0, 0, 0, time.UTC),
			)
			_ = s.Create(txn)
		}(i)
	}
	wg.Wait()

	list, err := s.List(n+10, 0)
	if err != nil {
		t.Fatalf("unexpected List error: %v", err)
	}
	if len(list) != n {
		t.Errorf("expected %d transactions, got %d", n, len(list))
	}
}

// Test: TestCreate_doesNotStoreOnConflict
// What: Create does not overwrite the stored transaction when a conflicting request is rejected
// Input: original (id="txn-1", amount=100), then conflicting (id="txn-1", amount=999)
// Output: Get("txn-1") still returns amount=100 after the conflict
func TestCreate_doesNotStoreOnConflict(t *testing.T) {
	s := store.NewMemoryStore()
	original := makeTxn("txn-1", 100, "USD", jan(1))
	conflicting := makeTxn("txn-1", 999, "USD", jan(1))

	_ = s.Create(original)
	_ = s.Create(conflicting)

	got, _ := s.Get("txn-1")
	if got.Amount != original.Amount {
		t.Errorf("conflicting write should not modify stored transaction: got amount %d", got.Amount)
	}
}

// Test: TestCreate_doesNotShareMetadataReference
// What: Create clones the transaction before storing — mutating the caller's Metadata map after Create does not affect the stored transaction
// Input: transaction created with metadata={"k":"v"}; caller mutates map to {"k":"mutated"} after Create
// Output: Get still returns metadata["k"]="v"
func TestCreate_doesNotShareMetadataReference(t *testing.T) {
	s := store.NewMemoryStore()
	meta := map[string]string{"k": "v"}
	txn := makeTxn("txn-1", 100, "USD", jan(1))
	txn.Metadata = meta
	_ = s.Create(txn)

	meta["k"] = "mutated" // mutate the original map after storing

	got, _ := s.Get("txn-1")
	if got.Metadata["k"] != "v" {
		t.Error("Create should clone the transaction; mutating caller's Metadata after Create should not affect the store")
	}
}
