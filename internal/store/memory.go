package store

import (
	"github.com/synctera/tech-challenge/internal/model"
)

type MemoryStore struct {
	// TODO: implement
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) Create(txn model.Transaction) error {
	// TODO: implement
	return nil
}

func (s *MemoryStore) Get(id string) (model.Transaction, error) {
	// TODO: implement
	return model.Transaction{}, ErrNotFound
}

func (s *MemoryStore) List(limit, offset int) ([]model.Transaction, error) {
	// TODO: implement
	return nil, nil
}
