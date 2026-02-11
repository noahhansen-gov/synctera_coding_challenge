package api

import (
	"net/http"

	"github.com/synctera/tech-challenge/internal/store"
)

type Handler struct {
	store store.Store
}

func NewHandler(s store.Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
