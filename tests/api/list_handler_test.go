package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/synctera/tech-challenge/internal/model"
)

// Test: TestListTransactions_emptyStore
// What: GET /transactions on a fresh store returns an empty JSON array
// Input: no stored transactions, no query params
// Output: HTTP 200, body decodes to empty []Transaction
func TestListTransactions_emptyStore(t *testing.T) {
	srv := newTestServer(t)

	resp := getTxns(t, srv, "")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var result []model.Transaction
	json.NewDecoder(resp.Body).Decode(&result)
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d items", len(result))
	}
}

// Test: TestListTransactions_returnsAllByDefault
// What: GET /transactions with no filters returns all stored transactions
// Input: 2 transactions seeded, no query params
// Output: HTTP 200, 2 transactions in the response body
func TestListTransactions_returnsAllByDefault(t *testing.T) {
	srv := newTestServer(t)
	seedTxn(t, srv, `{"id":"txn-1","amount":100,"currency":"USD","effective_at":"2024-01-01T00:00:00Z"}`)
	seedTxn(t, srv, `{"id":"txn-2","amount":200,"currency":"USD","effective_at":"2024-01-02T00:00:00Z"}`)

	resp := getTxns(t, srv, "")
	defer resp.Body.Close()

	var result []model.Transaction
	json.NewDecoder(resp.Body).Decode(&result)
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
}

// Test: TestListTransactions_orderedChronologically
// What: GET /transactions returns results sorted by effective_at ascending regardless of insertion order
// Input: transactions seeded in order [Mar, Jan, Feb]
// Output: response contains [txn-1(Jan), txn-2(Feb), txn-3(Mar)]
func TestListTransactions_orderedChronologically(t *testing.T) {
	srv := newTestServer(t)
	seedTxn(t, srv, `{"id":"txn-3","amount":300,"currency":"USD","effective_at":"2024-03-01T00:00:00Z"}`)
	seedTxn(t, srv, `{"id":"txn-1","amount":100,"currency":"USD","effective_at":"2024-01-01T00:00:00Z"}`)
	seedTxn(t, srv, `{"id":"txn-2","amount":200,"currency":"USD","effective_at":"2024-02-01T00:00:00Z"}`)

	resp := getTxns(t, srv, "")
	defer resp.Body.Close()

	var result []model.Transaction
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result))
	}
	expected := []string{"txn-1", "txn-2", "txn-3"}
	for i, txn := range result {
		if txn.ID != expected[i] {
			t.Errorf("index %d: expected %q, got %q", i, expected[i], txn.ID)
		}
	}
}

// Test: TestListTransactions_paginationLimit
// What: GET /transactions?limit=2 returns at most 2 results
// Input: 3 transactions seeded, query param limit=2
// Output: 2 transactions in the response body
func TestListTransactions_paginationLimit(t *testing.T) {
	srv := newTestServer(t)
	for _, body := range []string{
		`{"id":"a","amount":100,"currency":"USD","effective_at":"2024-01-01T00:00:00Z"}`,
		`{"id":"b","amount":200,"currency":"USD","effective_at":"2024-01-02T00:00:00Z"}`,
		`{"id":"c","amount":300,"currency":"USD","effective_at":"2024-01-03T00:00:00Z"}`,
	} {
		seedTxn(t, srv, body)
	}

	resp := getTxns(t, srv, "limit=2")
	defer resp.Body.Close()

	var result []model.Transaction
	json.NewDecoder(resp.Body).Decode(&result)
	if len(result) != 2 {
		t.Errorf("expected 2 items with limit=2, got %d", len(result))
	}
}

// Test: TestListTransactions_paginationOffset
// What: GET /transactions?offset=1 skips the first result and returns the rest
// Input: 3 transactions [a, b, c] seeded, query params limit=10&offset=1
// Output: 2 transactions starting at "b"
func TestListTransactions_paginationOffset(t *testing.T) {
	srv := newTestServer(t)
	for _, body := range []string{
		`{"id":"a","amount":100,"currency":"USD","effective_at":"2024-01-01T00:00:00Z"}`,
		`{"id":"b","amount":200,"currency":"USD","effective_at":"2024-01-02T00:00:00Z"}`,
		`{"id":"c","amount":300,"currency":"USD","effective_at":"2024-01-03T00:00:00Z"}`,
	} {
		seedTxn(t, srv, body)
	}

	resp := getTxns(t, srv, "limit=10&offset=1")
	defer resp.Body.Close()

	var result []model.Transaction
	json.NewDecoder(resp.Body).Decode(&result)
	if len(result) != 2 {
		t.Errorf("expected 2 items with offset=1, got %d", len(result))
	}
	if result[0].ID != "b" {
		t.Errorf("expected first item 'b', got %q", result[0].ID)
	}
}

// Test: TestListTransactions_filterByCurrency
// What: GET /transactions?currency=USD returns only transactions matching that currency
// Input: 3 transactions (2 USD, 1 EUR), query param currency=USD
// Output: 2 transactions, all with Currency="USD"
func TestListTransactions_filterByCurrency(t *testing.T) {
	srv := newTestServer(t)
	seedTxn(t, srv, `{"id":"usd-1","amount":100,"currency":"USD","effective_at":"2024-01-01T00:00:00Z"}`)
	seedTxn(t, srv, `{"id":"eur-1","amount":200,"currency":"EUR","effective_at":"2024-01-02T00:00:00Z"}`)
	seedTxn(t, srv, `{"id":"usd-2","amount":300,"currency":"USD","effective_at":"2024-01-03T00:00:00Z"}`)

	resp := getTxns(t, srv, "currency=USD")
	defer resp.Body.Close()

	var result []model.Transaction
	json.NewDecoder(resp.Body).Decode(&result)
	if len(result) != 2 {
		t.Errorf("expected 2 USD transactions, got %d", len(result))
	}
	for _, txn := range result {
		if txn.Currency != "USD" {
			t.Errorf("expected USD, got %q", txn.Currency)
		}
	}
}

// Test: TestListTransactions_filterByDateRange
// What: GET /transactions?start_date=...&end_date=... returns only transactions within that window
// Input: 3 transactions (Jan, Feb, Mar), query params start_date=2024-01-10&end_date=2024-02-20
// Output: 2 transactions (Jan and Feb)
func TestListTransactions_filterByDateRange(t *testing.T) {
	srv := newTestServer(t)
	seedTxn(t, srv, `{"id":"jan","amount":100,"currency":"USD","effective_at":"2024-01-15T12:00:00Z"}`)
	seedTxn(t, srv, `{"id":"feb","amount":200,"currency":"USD","effective_at":"2024-02-15T12:00:00Z"}`)
	seedTxn(t, srv, `{"id":"mar","amount":300,"currency":"USD","effective_at":"2024-03-15T12:00:00Z"}`)

	resp := getTxns(t, srv, "start_date=2024-01-10&end_date=2024-02-20")
	defer resp.Body.Close()

	var result []model.Transaction
	json.NewDecoder(resp.Body).Decode(&result)
	if len(result) != 2 {
		t.Errorf("expected 2 results in date range, got %d", len(result))
	}
}

// Test: TestListTransactions_filterByAmountRange
// What: GET /transactions?min_amount=...&max_amount=... returns only transactions within that range
// Input: 3 transactions (amounts 100, 500, 9000), query params min_amount=200&max_amount=1000
// Output: 1 transaction (amount=500, id="mid")
func TestListTransactions_filterByAmountRange(t *testing.T) {
	srv := newTestServer(t)
	seedTxn(t, srv, `{"id":"low","amount":100,"currency":"USD","effective_at":"2024-01-01T00:00:00Z"}`)
	seedTxn(t, srv, `{"id":"mid","amount":500,"currency":"USD","effective_at":"2024-01-02T00:00:00Z"}`)
	seedTxn(t, srv, `{"id":"high","amount":9000,"currency":"USD","effective_at":"2024-01-03T00:00:00Z"}`)

	resp := getTxns(t, srv, "min_amount=200&max_amount=1000")
	defer resp.Body.Close()

	var result []model.Transaction
	json.NewDecoder(resp.Body).Decode(&result)
	if len(result) != 1 || result[0].ID != "mid" {
		t.Errorf("expected only 'mid', got %d results", len(result))
	}
}

// Test: TestListTransactions_invalidLimit
// What: GET /transactions?limit=0 returns 400 Bad Request (limit must be at least 1)
// Input: query param limit=0
// Output: HTTP 400
func TestListTransactions_invalidLimit(t *testing.T) {
	srv := newTestServer(t)

	resp := getTxns(t, srv, "limit=0")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for limit=0, got %d", resp.StatusCode)
	}
}

// Test: TestListTransactions_limitTooHigh
// What: GET /transactions?limit=9999 returns 400 Bad Request (limit must be at most 1000)
// Input: query param limit=9999
// Output: HTTP 400
func TestListTransactions_limitTooHigh(t *testing.T) {
	srv := newTestServer(t)

	resp := getTxns(t, srv, "limit=9999")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for limit=9999, got %d", resp.StatusCode)
	}
}

// Test: TestListTransactions_negativeOffset
// What: GET /transactions?offset=-1 returns 400 Bad Request (offset must be >= 0)
// Input: query param offset=-1
// Output: HTTP 400
func TestListTransactions_negativeOffset(t *testing.T) {
	srv := newTestServer(t)

	resp := getTxns(t, srv, "offset=-1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for offset=-1, got %d", resp.StatusCode)
	}
}

// Test: TestListTransactions_invalidStartDate
// What: GET /transactions?start_date=not-a-date returns 400 Bad Request
// Input: query param start_date=not-a-date
// Output: HTTP 400
func TestListTransactions_invalidStartDate(t *testing.T) {
	srv := newTestServer(t)

	resp := getTxns(t, srv, "start_date=not-a-date")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Test: TestListTransactions_invalidEndDate
// What: GET /transactions?end_date=not-a-date returns 400 Bad Request
// Input: query param end_date=not-a-date
// Output: HTTP 400
func TestListTransactions_invalidEndDate(t *testing.T) {
	srv := newTestServer(t)

	resp := getTxns(t, srv, "end_date=not-a-date")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Test: TestListTransactions_startAfterEndDate
// What: GET /transactions with start_date > end_date returns 400 Bad Request (invalid range)
// Input: query params start_date=2024-12-31&end_date=2024-01-01
// Output: HTTP 400
func TestListTransactions_startAfterEndDate(t *testing.T) {
	srv := newTestServer(t)

	resp := getTxns(t, srv, "start_date=2024-12-31&end_date=2024-01-01")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when start > end, got %d", resp.StatusCode)
	}
}

// Test: TestListTransactions_invalidMinAmount
// What: GET /transactions?min_amount=abc returns 400 Bad Request (must be numeric)
// Input: query param min_amount=abc
// Output: HTTP 400
func TestListTransactions_invalidMinAmount(t *testing.T) {
	srv := newTestServer(t)

	resp := getTxns(t, srv, "min_amount=abc")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Test: TestListTransactions_minAmountGreaterThanMax
// What: GET /transactions with min_amount > max_amount returns 400 Bad Request (invalid range)
// Input: query params min_amount=500&max_amount=100
// Output: HTTP 400
func TestListTransactions_minAmountGreaterThanMax(t *testing.T) {
	srv := newTestServer(t)

	resp := getTxns(t, srv, "min_amount=500&max_amount=100")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when min > max, got %d", resp.StatusCode)
	}
}

// Test: TestListTransactions_sameTimestampOrderedByID
// What: GET /transactions with same-timestamp transactions returns them sorted alphabetically by ID
// Input: 3 transactions with identical effective_at, seeded in order: zzz, aaa, mmm
// Output: [aaa, mmm, zzz]
func TestListTransactions_sameTimestampOrderedByID(t *testing.T) {
	srv := newTestServer(t)
	ts := "2024-05-01T12:00:00Z"
	seedTxn(t, srv, `{"id":"zzz","amount":100,"currency":"USD","effective_at":"`+ts+`"}`)
	seedTxn(t, srv, `{"id":"aaa","amount":200,"currency":"USD","effective_at":"`+ts+`"}`)
	seedTxn(t, srv, `{"id":"mmm","amount":300,"currency":"USD","effective_at":"`+ts+`"}`)

	resp := getTxns(t, srv, "")
	defer resp.Body.Close()

	var result []model.Transaction
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}
	expected := []string{"aaa", "mmm", "zzz"}
	for i, txn := range result {
		if txn.ID != expected[i] {
			t.Errorf("index %d: expected %q, got %q", i, expected[i], txn.ID)
		}
	}
}
