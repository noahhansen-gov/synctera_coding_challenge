package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/synctera/tech-challenge/internal/api"
	"github.com/synctera/tech-challenge/internal/store"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	h := api.NewHandler(store.NewMemoryStore())
	mux := http.NewServeMux()
	mux.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateTransaction(w, r)
		case http.MethodGet:
			h.ListTransactions(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/transactions/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetTransaction(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func postTxn(t *testing.T, srv *httptest.Server, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(srv.URL+"/transactions", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("POST /transactions failed: %v", err)
	}
	return resp
}

func getTxns(t *testing.T, srv *httptest.Server, query string) *http.Response {
	t.Helper()
	url := srv.URL + "/transactions"
	if query != "" {
		url += "?" + query
	}
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET /transactions failed: %v", err)
	}
	return resp
}

func getTxnByID(t *testing.T, srv *httptest.Server, id string) *http.Response {
	t.Helper()
	resp, err := http.Get(srv.URL + "/transactions/" + id)
	if err != nil {
		t.Fatalf("GET /transactions/%s failed: %v", id, err)
	}
	return resp
}

func seedTxn(t *testing.T, srv *httptest.Server, body string) {
	t.Helper()
	resp := postTxn(t, srv, body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("seed failed with status %d for body: %s", resp.StatusCode, body)
	}
}
