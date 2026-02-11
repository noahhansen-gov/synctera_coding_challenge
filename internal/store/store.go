package store

import (
	"github.com/synctera/tech-challenge/internal/model"
)

// Store defines the interface for transaction storage.
type Store interface {
	Create(txn model.Transaction) error
	Get(id string) (model.Transaction, error)
	List(limit, offset int) ([]model.Transaction, error)
}

// Common errors.
type StoreError string

func (e StoreError) Error() string { return string(e) }

const (
	ErrNotFound  StoreError = "transaction not found"
	ErrConflict  StoreError = "conflict"
	ErrDuplicate StoreError = "duplicate"
)
