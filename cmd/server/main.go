package main

import (
	"log"
	"net/http"

	"github.com/synctera/tech-challenge/internal/api"
	"github.com/synctera/tech-challenge/internal/store"
)

func main() {
	// Initialize store
	memStore := store.NewMemoryStore()

	// Initialize handlers
	handler := api.NewHandler(memStore)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.CreateTransaction(w, r)
		case http.MethodGet:
			handler.ListTransactions(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
