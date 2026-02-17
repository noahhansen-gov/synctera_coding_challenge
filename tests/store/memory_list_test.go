package store_test

import (
	"testing"

	"github.com/synctera/tech-challenge/internal/store"
)

// Test: TestList_emptyStore
// What: List on an empty store returns an empty slice without error
// Input: MemoryStore with no data, limit=10, offset=0
// Output: empty []Transaction, nil error
func TestList_emptyStore(t *testing.T) {
	s := store.NewMemoryStore()

	list, err := s.List(10, 0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty slice, got %d items", len(list))
	}
}

// Test: TestList_returnsAllWithinLimit
// What: List returns all stored transactions when limit is larger than the total count
// Input: store with 3 transactions, limit=10, offset=0
// Output: 3 transactions, nil error
func TestList_returnsAllWithinLimit(t *testing.T) {
	s := store.NewMemoryStore()
	_ = s.Create(makeTxn("a", 100, "USD", jan(1)))
	_ = s.Create(makeTxn("b", 200, "USD", jan(2)))
	_ = s.Create(makeTxn("c", 300, "USD", jan(3)))

	list, err := s.List(10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 items, got %d", len(list))
	}
}

// Test: TestList_respectsLimit
// What: List returns at most `limit` transactions
// Input: store with 5 transactions, limit=2, offset=0
// Output: exactly 2 transactions
func TestList_respectsLimit(t *testing.T) {
	s := store.NewMemoryStore()
	for i := 1; i <= 5; i++ {
		_ = s.Create(makeTxn("txn-"+string(rune('0'+i)), 100, "USD", jan(i)))
	}

	list, _ := s.List(2, 0)
	if len(list) != 2 {
		t.Errorf("expected 2 items, got %d", len(list))
	}
}

// Test: TestList_respectsOffset
// What: List skips the first `offset` transactions and starts from the next one
// Input: store with [a, b, c] (sorted by date), limit=10, offset=1
// Output: [b, c] — "a" is skipped
func TestList_respectsOffset(t *testing.T) {
	s := store.NewMemoryStore()
	_ = s.Create(makeTxn("a", 100, "USD", jan(1)))
	_ = s.Create(makeTxn("b", 200, "USD", jan(2)))
	_ = s.Create(makeTxn("c", 300, "USD", jan(3)))

	list, _ := s.List(10, 1)
	if len(list) != 2 {
		t.Errorf("expected 2 items after offset 1, got %d", len(list))
	}
	if list[0].ID != "b" {
		t.Errorf("expected first item to be 'b', got %q", list[0].ID)
	}
}

// Test: TestList_offsetBeyondLength
// What: List returns an empty slice without error when offset exceeds the number of stored items
// Input: store with 1 transaction, limit=10, offset=100
// Output: empty slice, nil error
func TestList_offsetBeyondLength(t *testing.T) {
	s := store.NewMemoryStore()
	_ = s.Create(makeTxn("a", 100, "USD", jan(1)))

	list, err := s.List(10, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty slice for out-of-bounds offset, got %d items", len(list))
	}
}

// Test: TestList_limitExceedsRemaining
// What: List returns only the remaining items when limit > (total - offset)
// Input: store with 2 transactions [a, b], limit=10, offset=1
// Output: 1 transaction [b]
func TestList_limitExceedsRemaining(t *testing.T) {
	s := store.NewMemoryStore()
	_ = s.Create(makeTxn("a", 100, "USD", jan(1)))
	_ = s.Create(makeTxn("b", 200, "USD", jan(2)))

	list, _ := s.List(10, 1)
	if len(list) != 1 {
		t.Errorf("expected 1 item, got %d", len(list))
	}
}

// Test: TestList_orderedByEffectiveAt
// What: List returns transactions in ascending chronological order regardless of insertion order
// Input: transactions inserted as [c(Jan 3), a(Jan 1), b(Jan 2)]
// Output: [a, b, c] sorted by EffectiveAt
func TestList_orderedByEffectiveAt(t *testing.T) {
	s := store.NewMemoryStore()
	_ = s.Create(makeTxn("c", 100, "USD", jan(3)))
	_ = s.Create(makeTxn("a", 100, "USD", jan(1)))
	_ = s.Create(makeTxn("b", 100, "USD", jan(2)))

	list, _ := s.List(10, 0)
	expected := []string{"a", "b", "c"}
	for i, txn := range list {
		if txn.ID != expected[i] {
			t.Errorf("index %d: expected %q, got %q", i, expected[i], txn.ID)
		}
	}
}

// Test: TestList_sameTimestampOrderedByID
// What: List breaks ties on identical effective_at by sorting alphabetically by ID
// Input: transactions with same timestamp inserted as [zzz, aaa, mmm]
// Output: [aaa, mmm, zzz]
func TestList_sameTimestampOrderedByID(t *testing.T) {
	s := store.NewMemoryStore()
	ts := jan(1)
	_ = s.Create(makeTxn("zzz", 100, "USD", ts))
	_ = s.Create(makeTxn("aaa", 100, "USD", ts))
	_ = s.Create(makeTxn("mmm", 100, "USD", ts))

	list, _ := s.List(10, 0)
	expected := []string{"aaa", "mmm", "zzz"}
	for i, txn := range list {
		if txn.ID != expected[i] {
			t.Errorf("index %d: expected %q, got %q", i, expected[i], txn.ID)
		}
	}
}

// Test: TestList_outOfOrderInsertionResultsInCorrectOrder
// What: List returns chronological order even when transactions are inserted in reverse order
// Input: transactions inserted as [e(Jan 5), d(Jan 4), c(Jan 3), b(Jan 2), a(Jan 1)]
// Output: transactions listed with EffectiveAt.Day() == index+1
func TestList_outOfOrderInsertionResultsInCorrectOrder(t *testing.T) {
	s := store.NewMemoryStore()
	_ = s.Create(makeTxn("e", 100, "USD", jan(5)))
	_ = s.Create(makeTxn("d", 100, "USD", jan(4)))
	_ = s.Create(makeTxn("c", 100, "USD", jan(3)))
	_ = s.Create(makeTxn("b", 100, "USD", jan(2)))
	_ = s.Create(makeTxn("a", 100, "USD", jan(1)))

	list, _ := s.List(10, 0)
	for i, txn := range list {
		if txn.EffectiveAt.Day() != i+1 {
			t.Errorf("index %d: expected day %d, got %d", i, i+1, txn.EffectiveAt.Day())
		}
	}
}

// Test: TestList_paginationCoversAllItems
// What: Consecutive pages cover all stored items with no gaps when limit divides the total unevenly
// Input: store with 5 transactions, three pages of size 2 (offsets 0, 2, 4)
// Output: total 5 items across three pages
func TestList_paginationCoversAllItems(t *testing.T) {
	s := store.NewMemoryStore()
	for i := 1; i <= 5; i++ {
		_ = s.Create(makeTxn("txn-"+string(rune('0'+i)), 100, "USD", jan(i)))
	}

	page1, _ := s.List(2, 0)
	page2, _ := s.List(2, 2)
	page3, _ := s.List(2, 4)

	total := len(page1) + len(page2) + len(page3)
	if total != 5 {
		t.Errorf("expected 5 total items across pages, got %d", total)
	}
}

// Test: TestList_returnsACopy
// What: List returns a copy of the internal data — mutating the returned slice does not affect the store
// Input: store with one transaction (amount=100); mutate returned slice to amount=9999
// Output: Get still returns amount=100 (store is unaffected)
func TestList_returnsACopy(t *testing.T) {
	s := store.NewMemoryStore()
	_ = s.Create(makeTxn("a", 100, "USD", jan(1)))

	list, _ := s.List(10, 0)
	list[0].Amount = 9999

	got, _ := s.Get("a")
	if got.Amount == 9999 {
		t.Error("List should return a copy; modifying it should not affect the store")
	}
}
