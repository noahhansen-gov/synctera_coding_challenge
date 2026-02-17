package api_test

import (
	"testing"
	"time"

	"github.com/synctera/tech-challenge/internal/api"
	"github.com/synctera/tech-challenge/internal/model"
)

func makePaginationData(n int) []model.Transaction {
	txns := make([]model.Transaction, n)
	for i := range txns {
		txns[i] = model.Transaction{
			ID:          string(rune('a' + i)),
			Amount:      int64(i * 100),
			Currency:    "USD",
			EffectiveAt: time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
		}
	}
	return txns
}

// Test: TestApplyPagination_firstPage
// What: ApplyPagination with offset=0 returns the first `limit` elements
// Input: 5 transactions, limit=2, offset=0
// Output: 2 transactions ["a", "b"]
func TestApplyPagination_firstPage(t *testing.T) {
	data := makePaginationData(5)
	result := api.ApplyPagination(data, 2, 0)
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
	if result[0].ID != "a" || result[1].ID != "b" {
		t.Errorf("unexpected items: %v", result)
	}
}

// Test: TestApplyPagination_secondPage
// What: ApplyPagination with offset=2 skips the first page and returns the next slice
// Input: 5 transactions, limit=2, offset=2
// Output: 2 transactions ["c", "d"]
func TestApplyPagination_secondPage(t *testing.T) {
	data := makePaginationData(5)
	result := api.ApplyPagination(data, 2, 2)
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
	if result[0].ID != "c" || result[1].ID != "d" {
		t.Errorf("unexpected items: %v", result)
	}
}

// Test: TestApplyPagination_lastPagePartial
// What: ApplyPagination returns fewer items than limit when remaining data is smaller than limit
// Input: 5 transactions, limit=3, offset=4 (1 item remaining)
// Output: 1 transaction
func TestApplyPagination_lastPagePartial(t *testing.T) {
	data := makePaginationData(5)
	result := api.ApplyPagination(data, 3, 4)
	if len(result) != 1 {
		t.Errorf("expected 1 item on partial last page, got %d", len(result))
	}
}

// Test: TestApplyPagination_offsetBeyondLength
// What: ApplyPagination returns an empty slice when offset exceeds data length
// Input: 3 transactions, limit=10, offset=100
// Output: empty slice
func TestApplyPagination_offsetBeyondLength(t *testing.T) {
	data := makePaginationData(3)
	result := api.ApplyPagination(data, 10, 100)
	if len(result) != 0 {
		t.Errorf("expected empty result for out-of-bounds offset, got %d", len(result))
	}
}

// Test: TestApplyPagination_emptyInput
// What: ApplyPagination on an empty slice always returns an empty slice
// Input: empty []model.Transaction, limit=10, offset=0
// Output: empty slice
func TestApplyPagination_emptyInput(t *testing.T) {
	result := api.ApplyPagination([]model.Transaction{}, 10, 0)
	if len(result) != 0 {
		t.Errorf("expected empty result for empty input, got %d", len(result))
	}
}

// Test: TestApplyPagination_limitLargerThanData
// What: ApplyPagination returns all items when limit exceeds total data size
// Input: 3 transactions, limit=1000, offset=0
// Output: all 3 transactions
func TestApplyPagination_limitLargerThanData(t *testing.T) {
	data := makePaginationData(3)
	result := api.ApplyPagination(data, 1000, 0)
	if len(result) != 3 {
		t.Errorf("expected all 3 items when limit exceeds data, got %d", len(result))
	}
}

// Test: TestApplyPagination_limitExactlyFits
// What: ApplyPagination returns all items when limit == data length
// Input: 4 transactions, limit=4, offset=0
// Output: all 4 transactions
func TestApplyPagination_limitExactlyFits(t *testing.T) {
	data := makePaginationData(4)
	result := api.ApplyPagination(data, 4, 0)
	if len(result) != 4 {
		t.Errorf("expected 4 items when limit exactly matches data size, got %d", len(result))
	}
}

// Test: TestApplyPagination_doesNotOverlap
// What: consecutive pages cover all items exactly once with no overlap and no gaps
// Input: 6 transactions split into two pages (limit=3, offset=0 and offset=3)
// Output: 6 unique IDs across both pages, none repeated
func TestApplyPagination_doesNotOverlap(t *testing.T) {
	data := makePaginationData(6)
	page1 := api.ApplyPagination(data, 3, 0)
	page2 := api.ApplyPagination(data, 3, 3)

	seen := make(map[string]bool)
	for _, txn := range append(page1, page2...) {
		if seen[txn.ID] {
			t.Errorf("ID %q appeared on multiple pages", txn.ID)
		}
		seen[txn.ID] = true
	}
	if len(seen) != 6 {
		t.Errorf("expected 6 unique items across pages, got %d", len(seen))
	}
}
