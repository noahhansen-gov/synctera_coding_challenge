package api_test

import (
	"testing"
	"time"

	"github.com/synctera/tech-challenge/internal/api"
	"github.com/synctera/tech-challenge/internal/model"
)

func makeFilterTxn(id, currency string, amount int64, year, month, day int) model.Transaction {
	return model.Transaction{
		ID:          id,
		Amount:      amount,
		Currency:    currency,
		EffectiveAt: time.Date(year, time.Month(month), day, 12, 0, 0, 0, time.UTC),
	}
}

var filterTestData = []model.Transaction{
	makeFilterTxn("usd-jan-low", "USD", 500, 2024, 1, 10),
	makeFilterTxn("usd-feb-high", "USD", 50000, 2024, 2, 15),
	makeFilterTxn("eur-jan-mid", "EUR", 5000, 2024, 1, 20),
	makeFilterTxn("gbp-mar-low", "GBP", 300, 2024, 3, 5),
}

// Test: TestApplyFilters_noFilters
// What: ApplyFilters with all nil/empty args returns all transactions unchanged
// Input: filterTestData (4 transactions), no currency/date/amount filters
// Output: all 4 transactions
func TestApplyFilters_noFilters(t *testing.T) {
	result := api.ApplyFilters(filterTestData, "", nil, nil, nil, nil)
	if len(result) != len(filterTestData) {
		t.Errorf("expected %d results with no filters, got %d", len(filterTestData), len(result))
	}
}

// Test: TestApplyFilters_emptyInput
// What: ApplyFilters on an empty slice returns an empty slice regardless of filters
// Input: empty []model.Transaction, currency="USD"
// Output: empty slice
func TestApplyFilters_emptyInput(t *testing.T) {
	result := api.ApplyFilters([]model.Transaction{}, "USD", nil, nil, nil, nil)
	if len(result) != 0 {
		t.Errorf("expected empty result for empty input, got %d", len(result))
	}
}

// Test: TestApplyFilters_byCurrency
// What: ApplyFilters with a currency filter returns only matching transactions
// Input: filterTestData, currency="USD"
// Output: 2 USD transactions (usd-jan-low, usd-feb-high)
func TestApplyFilters_byCurrency(t *testing.T) {
	result := api.ApplyFilters(filterTestData, "USD", nil, nil, nil, nil)
	if len(result) != 2 {
		t.Errorf("expected 2 USD transactions, got %d", len(result))
	}
	for _, txn := range result {
		if txn.Currency != "USD" {
			t.Errorf("expected USD, got %q", txn.Currency)
		}
	}
}

// Test: TestApplyFilters_byCurrencyCaseInsensitive
// What: ApplyFilters matches currency case-insensitively via strings.EqualFold
// Input: filterTestData, currency="usd" (lowercase)
// Output: 2 transactions (same as "USD")
func TestApplyFilters_byCurrencyCaseInsensitive(t *testing.T) {
	result := api.ApplyFilters(filterTestData, "usd", nil, nil, nil, nil)
	if len(result) != 2 {
		t.Errorf("expected 2 results for lowercase 'usd', got %d", len(result))
	}
}

// Test: TestApplyFilters_byStartDate
// What: ApplyFilters with a start date excludes transactions before that date
// Input: filterTestData, startDate=2024-02-01
// Output: 2 transactions (Feb and Mar; Jan filtered out)
func TestApplyFilters_byStartDate(t *testing.T) {
	startDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	result := api.ApplyFilters(filterTestData, "", &startDate, nil, nil, nil)

	if len(result) != 2 {
		t.Errorf("expected 2 results after start_date=2024-02-01, got %d", len(result))
	}
}

// Test: TestApplyFilters_byEndDate
// What: ApplyFilters with an end date excludes transactions after that date (inclusive of the end day)
// Input: filterTestData, endDate=2024-01-31
// Output: 2 transactions (Jan 10 + Jan 20)
func TestApplyFilters_byEndDate(t *testing.T) {
	endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	result := api.ApplyFilters(filterTestData, "", nil, &endDate, nil, nil)

	if len(result) != 2 {
		t.Errorf("expected 2 Jan results, got %d", len(result))
	}
}

// Test: TestApplyFilters_endDateIsInclusive
// What: ApplyFilters treats end date as inclusive — a transaction occurring on the end date is included
// Input: txns=[noon Jan 10, Jan 12], endDate=2024-01-10 (midnight)
// Output: 1 transaction ("included" — noon Jan 10 is within Jan 10 day)
func TestApplyFilters_endDateIsInclusive(t *testing.T) {
	endDate := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	txns := []model.Transaction{
		makeFilterTxn("included", "USD", 100, 2024, 1, 10),
		makeFilterTxn("excluded", "USD", 100, 2024, 1, 12),
	}

	result := api.ApplyFilters(txns, "", nil, &endDate, nil, nil)
	if len(result) != 1 {
		t.Errorf("expected 1 result (inclusive end date), got %d", len(result))
	}
	if len(result) > 0 && result[0].ID != "included" {
		t.Errorf("expected 'included' transaction, got %q", result[0].ID)
	}
}

// Test: TestApplyFilters_byDateRange
// What: ApplyFilters with both start and end returns only transactions within the window
// Input: filterTestData, start=2024-01-15, end=2024-02-28
// Output: 2 transactions (Jan 20 EUR + Feb 15 USD)
func TestApplyFilters_byDateRange(t *testing.T) {
	start := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)
	result := api.ApplyFilters(filterTestData, "", &start, &end, nil, nil)

	if len(result) != 2 {
		t.Errorf("expected 2 results in date range, got %d", len(result))
	}
}

// Test: TestApplyFilters_byMinAmount
// What: ApplyFilters with a min amount excludes transactions below the threshold
// Input: filterTestData, minAmount=1000
// Output: 2 transactions (eur-jan-mid=5000, usd-feb-high=50000)
func TestApplyFilters_byMinAmount(t *testing.T) {
	min := int64(1000)
	result := api.ApplyFilters(filterTestData, "", nil, nil, &min, nil)

	if len(result) != 2 {
		t.Errorf("expected 2 results with min_amount=1000, got %d", len(result))
	}
}

// Test: TestApplyFilters_byMaxAmount
// What: ApplyFilters with a max amount excludes transactions above the threshold
// Input: filterTestData, maxAmount=1000
// Output: 2 transactions (usd-jan-low=500, gbp-mar-low=300)
func TestApplyFilters_byMaxAmount(t *testing.T) {
	max := int64(1000)
	result := api.ApplyFilters(filterTestData, "", nil, nil, nil, &max)

	if len(result) != 2 {
		t.Errorf("expected 2 results with max_amount=1000, got %d", len(result))
	}
}

// Test: TestApplyFilters_byExactAmountRange
// What: ApplyFilters with min == max acts as an exact amount match
// Input: filterTestData, minAmount=500, maxAmount=500
// Output: 1 transaction (usd-jan-low with amount=500)
func TestApplyFilters_byExactAmountRange(t *testing.T) {
	min := int64(500)
	max := int64(500)
	result := api.ApplyFilters(filterTestData, "", nil, nil, &min, &max)

	if len(result) != 1 || result[0].ID != "usd-jan-low" {
		t.Errorf("expected only 'usd-jan-low' for exact amount 500, got %d results", len(result))
	}
}

// Test: TestApplyFilters_combined
// What: ApplyFilters ANDs all active filters — only transactions matching every filter are returned
// Input: filterTestData, currency="USD", start=2024-01-01, end=2024-01-31, minAmount=100, maxAmount=600
// Output: 1 transaction (usd-jan-low — USD, Jan, amount=500 within [100,600])
func TestApplyFilters_combined(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	min := int64(100)
	max := int64(600)
	result := api.ApplyFilters(filterTestData, "USD", &start, &end, &min, &max)

	if len(result) != 1 {
		t.Errorf("expected 1 result with combined filters, got %d", len(result))
	}
	if len(result) > 0 && result[0].ID != "usd-jan-low" {
		t.Errorf("expected 'usd-jan-low', got %q", result[0].ID)
	}
}

// Test: TestApplyFilters_noMatches
// What: ApplyFilters returns an empty slice when no transactions match the filter
// Input: filterTestData, currency="JPY" (not present in data)
// Output: empty slice
func TestApplyFilters_noMatches(t *testing.T) {
	result := api.ApplyFilters(filterTestData, "JPY", nil, nil, nil, nil)
	if len(result) != 0 {
		t.Errorf("expected 0 results for JPY filter, got %d", len(result))
	}
}
