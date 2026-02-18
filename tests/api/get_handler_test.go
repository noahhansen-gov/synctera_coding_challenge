package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/synctera/tech-challenge/internal/model"
)

// Test: TestGetTransaction_success
// What: GET /transactions/{id} returns 200 and the stored transaction when the ID exists
// Input: one seeded transaction with id="txn-1", lookup by "txn-1"
// Output: HTTP 200, response body contains the transaction
func TestGetTransaction_success(t *testing.T) {
	srv := newTestServer(t)
	seedTxn(t, srv, `{"id":"txn-1","amount":1000,"currency":"USD","effective_at":"2024-01-15T12:00:00Z"}`)

	resp := getTxnByID(t, srv, "txn-1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var got model.Transaction
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.ID != "txn-1" {
		t.Errorf("expected ID 'txn-1', got %q", got.ID)
	}
}

// Test: TestGetTransaction_notFound
// What: GET /transactions/{id} returns 404 when the ID does not exist in the store
// Input: empty store, lookup by "nonexistent"
// Output: HTTP 404
func TestGetTransaction_notFound(t *testing.T) {
	srv := newTestServer(t)

	resp := getTxnByID(t, srv, "nonexistent")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// Test: TestGetTransaction_responseBodyFields
// What: GET /transactions/{id} returns all fields of the stored transaction intact
// Input: transaction with id="txn-42", amount=4200, currency="EUR", effective_at="2024-06-01T00:00:00Z"
// Output: HTTP 200, decoded body has matching ID, Amount, and Currency
func TestGetTransaction_responseBodyFields(t *testing.T) {
	srv := newTestServer(t)
	seedTxn(t, srv, `{"id":"txn-42","amount":4200,"currency":"EUR","effective_at":"2024-06-01T00:00:00Z"}`)

	resp := getTxnByID(t, srv, "txn-42")
	defer resp.Body.Close()

	var got model.Transaction
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got.ID != "txn-42" {
		t.Errorf("ID: expected 'txn-42', got %q", got.ID)
	}
	if got.Amount != 4200 {
		t.Errorf("Amount: expected 4200, got %d", got.Amount)
	}
	if got.Currency != "EUR" {
		t.Errorf("Currency: expected 'EUR', got %q", got.Currency)
	}
}

// Test: TestGetTransaction_correctTransactionAmongMany
// What: GET /transactions/{id} retrieves the correct transaction when multiple are stored
// Input: three seeded transactions (txn-1, txn-2, txn-3), lookup by "txn-2"
// Output: HTTP 200, response body contains txn-2 with amount=200
func TestGetTransaction_correctTransactionAmongMany(t *testing.T) {
	srv := newTestServer(t)
	seedTxn(t, srv, `{"id":"txn-1","amount":100,"currency":"USD","effective_at":"2024-01-01T00:00:00Z"}`)
	seedTxn(t, srv, `{"id":"txn-2","amount":200,"currency":"EUR","effective_at":"2024-01-02T00:00:00Z"}`)
	seedTxn(t, srv, `{"id":"txn-3","amount":300,"currency":"GBP","effective_at":"2024-01-03T00:00:00Z"}`)

	resp := getTxnByID(t, srv, "txn-2")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var got model.Transaction
	json.NewDecoder(resp.Body).Decode(&got)
	if got.ID != "txn-2" || got.Amount != 200 || got.Currency != "EUR" {
		t.Errorf("unexpected transaction returned: %+v", got)
	}
}

// Test: TestGetTransaction_withMetadata
// What: GET /transactions/{id} returns the metadata field when it was set on creation
// Input: transaction seeded with metadata={"source":"mobile"}, lookup by id
// Output: HTTP 200, response body contains Metadata["source"]="mobile"
func TestGetTransaction_withMetadata(t *testing.T) {
	srv := newTestServer(t)
	seedTxn(t, srv, `{"id":"txn-meta","amount":500,"currency":"USD","effective_at":"2024-01-01T00:00:00Z","metadata":{"source":"mobile"}}`)

	resp := getTxnByID(t, srv, "txn-meta")
	defer resp.Body.Close()

	var got model.Transaction
	json.NewDecoder(resp.Body).Decode(&got)
	if got.Metadata["source"] != "mobile" {
		t.Errorf("expected metadata source=mobile, got %v", got.Metadata)
	}
}

// Test: TestGetTransaction_contentTypeJSON
// What: GET /transactions/{id} sets Content-Type: application/json on a successful response
// Input: one seeded transaction, lookup by id
// Output: HTTP 200, Content-Type header is "application/json"
func TestGetTransaction_contentTypeJSON(t *testing.T) {
	srv := newTestServer(t)
	seedTxn(t, srv, `{"id":"txn-1","amount":100,"currency":"USD","effective_at":"2024-01-01T00:00:00Z"}`)

	resp := getTxnByID(t, srv, "txn-1")
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
