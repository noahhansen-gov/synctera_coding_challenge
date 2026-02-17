package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/synctera/tech-challenge/internal/model"
)

// Test: TestCreateTransaction_success
// What: POST /transactions with a valid payload stores the transaction and returns it
// Input: JSON body with id, amount, currency, effective_at all set
// Output: HTTP 201, response body contains the created transaction
func TestCreateTransaction_success(t *testing.T) {
	srv := newTestServer(t)
	body := `{"id":"txn-1","amount":1000,"currency":"USD","effective_at":"2024-01-15T12:00:00Z"}`

	resp := postTxn(t, srv, body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	var got model.Transaction
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.ID != "txn-1" {
		t.Errorf("expected ID 'txn-1', got %q", got.ID)
	}
}

// Test: TestCreateTransaction_idempotentRetry
// What: POSTing the same payload twice returns 201 first, then 200 (idempotent retry, not an error)
// Input: same JSON body posted twice
// Output: first call HTTP 201, second call HTTP 200
func TestCreateTransaction_idempotentRetry(t *testing.T) {
	srv := newTestServer(t)
	body := `{"id":"txn-1","amount":1000,"currency":"USD","effective_at":"2024-01-15T12:00:00Z"}`

	resp1 := postTxn(t, srv, body)
	resp1.Body.Close()
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("first request: expected 201, got %d", resp1.StatusCode)
	}

	resp2 := postTxn(t, srv, body)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("idempotent retry: expected 200, got %d", resp2.StatusCode)
	}
}

// Test: TestCreateTransaction_conflict
// What: POST with the same ID but different payload returns 409 Conflict
// Input: original (amount=1000), then conflicting (same id, amount=9999)
// Output: second call returns HTTP 409
func TestCreateTransaction_conflict(t *testing.T) {
	srv := newTestServer(t)
	original := `{"id":"txn-1","amount":1000,"currency":"USD","effective_at":"2024-01-15T12:00:00Z"}`
	conflicting := `{"id":"txn-1","amount":9999,"currency":"USD","effective_at":"2024-01-15T12:00:00Z"}`

	resp1 := postTxn(t, srv, original)
	resp1.Body.Close()

	resp2 := postTxn(t, srv, conflicting)
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusConflict {
		t.Errorf("expected 409, got %d", resp2.StatusCode)
	}
}

// Test: TestCreateTransaction_invalidJSON
// What: POST with a malformed JSON body returns 400 Bad Request
// Input: body="{not valid json"
// Output: HTTP 400
func TestCreateTransaction_invalidJSON(t *testing.T) {
	srv := newTestServer(t)

	resp := postTxn(t, srv, `{not valid json`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Test: TestCreateTransaction_missingID
// What: POST without an "id" field returns 400 Bad Request
// Input: JSON body with amount, currency, effective_at but no id
// Output: HTTP 400
func TestCreateTransaction_missingID(t *testing.T) {
	srv := newTestServer(t)
	body := `{"amount":1000,"currency":"USD","effective_at":"2024-01-15T12:00:00Z"}`

	resp := postTxn(t, srv, body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Test: TestCreateTransaction_missingCurrency
// What: POST without a "currency" field returns 400 Bad Request
// Input: JSON body with id, amount, effective_at but no currency
// Output: HTTP 400
func TestCreateTransaction_missingCurrency(t *testing.T) {
	srv := newTestServer(t)
	body := `{"id":"txn-1","amount":1000,"effective_at":"2024-01-15T12:00:00Z"}`

	resp := postTxn(t, srv, body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Test: TestCreateTransaction_missingEffectiveAt
// What: POST without an "effective_at" field returns 400 Bad Request
// Input: JSON body with id, amount, currency but no effective_at
// Output: HTTP 400
func TestCreateTransaction_missingEffectiveAt(t *testing.T) {
	srv := newTestServer(t)
	body := `{"id":"txn-1","amount":1000,"currency":"USD"}`

	resp := postTxn(t, srv, body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Test: TestCreateTransaction_negativeAmount
// What: POST with a negative amount returns 400 Bad Request (amounts must be >= 0)
// Input: JSON body with amount=-100
// Output: HTTP 400
func TestCreateTransaction_negativeAmount(t *testing.T) {
	srv := newTestServer(t)
	body := `{"id":"txn-1","amount":-100,"currency":"USD","effective_at":"2024-01-15T12:00:00Z"}`

	resp := postTxn(t, srv, body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Test: TestCreateTransaction_zeroAmountAllowed
// What: POST with amount=0 is valid and returns 201 (zero-value transactions are permitted)
// Input: JSON body with amount=0, all other fields valid
// Output: HTTP 201
func TestCreateTransaction_zeroAmountAllowed(t *testing.T) {
	srv := newTestServer(t)
	body := `{"id":"txn-1","amount":0,"currency":"USD","effective_at":"2024-01-15T12:00:00Z"}`

	resp := postTxn(t, srv, body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201 for zero amount, got %d", resp.StatusCode)
	}
}

// Test: TestCreateTransaction_responseBodyContainsTransaction
// What: POST response body contains the full transaction that was just created
// Input: JSON body with id="txn-abc", amount=4200, currency="EUR"
// Output: HTTP 201, response body decodes to a Transaction with matching fields
func TestCreateTransaction_responseBodyContainsTransaction(t *testing.T) {
	srv := newTestServer(t)
	body := `{"id":"txn-abc","amount":4200,"currency":"EUR","effective_at":"2024-06-01T00:00:00Z"}`

	resp := postTxn(t, srv, body)
	defer resp.Body.Close()

	var got model.Transaction
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if got.ID != "txn-abc" || got.Amount != 4200 || got.Currency != "EUR" {
		t.Errorf("response body mismatch: got %+v", got)
	}
}

// Test: TestCreateTransaction_withMetadata
// What: POST with optional metadata field stores and returns it in the response
// Input: JSON body with metadata={"source":"mobile"}
// Output: HTTP 201, response body contains Metadata["source"]="mobile"
func TestCreateTransaction_withMetadata(t *testing.T) {
	srv := newTestServer(t)
	body := `{"id":"txn-1","amount":100,"currency":"USD","effective_at":"2024-01-01T00:00:00Z","metadata":{"source":"mobile"}}`

	resp := postTxn(t, srv, body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	var got model.Transaction
	json.NewDecoder(resp.Body).Decode(&got)
	if got.Metadata["source"] != "mobile" {
		t.Errorf("expected metadata source=mobile, got %v", got.Metadata)
	}
}
