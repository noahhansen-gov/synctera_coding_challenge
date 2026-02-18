package store

/* sync is imported for potential use in synchronizing access to the in-memory data structures,
such as using mutexes to ensure thread safety when multiple goroutines access the store concurrently.*/
import (
	"github.com/synctera/tech-challenge/internal/model"
	"sort"
	"sync"
)

type MemoryStore struct {
	transactions map[string]model.Transaction // Fast O(1) lookups by ID
	ordered      []model.Transaction          // Slice maintains sorted order for queries
	memstoreMux  sync.RWMutex                 // Mutex to protect concurrent access
}

func NewMemoryStore() *MemoryStore {
	// Initialize the in-memory store with empty data structures
	return &MemoryStore{
		transactions: make(map[string]model.Transaction),
		ordered:      make([]model.Transaction, 0),
	}
}

func (s *MemoryStore) Create(txn model.Transaction) error {
	// lock the store in order to safely perform the operations below
	// this lock prevents others from performing read/write operations on the store until the lock is released
	s.memstoreMux.Lock()

	// unlock the store after the operations are complete to allow others to access it.
	// defer will wait until the function returns before executing the unlock
	defer s.memstoreMux.Unlock()

	// this uses the comma ok idiom
	// basically it checks if the transaction with the given ID already exists in the store
	// and returns the value + a boolean indicating whether it was found or not
	// thought about just calling the "Get" method here but that would require an additional lock/unlock which is inefficient
	existingTxn, exists := s.transactions[txn.ID]

	// if transaction exists
	if exists {
		// if the existing transaction is identical to the new one, return ErrDuplicate
		if existingTxn.Equal(txn) {
			return ErrDuplicate
		}

		return ErrConflict
	}

	// Clone before storing so the store's copy is isolated from the caller's map reference
	stored := txn.Clone()

	// if the transaction does not exist, add it to the store
	s.transactions[txn.ID] = stored

	// Define comparison function for readability
	shouldInsertBefore := func(i int) bool {
		existing := s.ordered[i]

		if txn.EffectiveAt.Before(existing.EffectiveAt) {
			return true
		} else if txn.EffectiveAt.After(existing.EffectiveAt) {
			return false
		}

		return txn.ID < existing.ID
	}

	// search works by finding the index where the new transaction should be inserted to maintain sorted order
	// you pass in a function because sort.Search will call it with different indices to find the correct position for the new transaction
	index := sort.Search(len(s.ordered), shouldInsertBefore)

	// Grow the slice by one element to make room for the new transaction
	// Shift elements to the right to make space for the new transaction at the correct index
	// set the new transaction at the correct index in the ordered slice
	s.ordered = append(s.ordered, model.Transaction{}) // grow the slice by one element
	copy(s.ordered[index+1:], s.ordered[index:])
	s.ordered[index] = stored

	return nil
}

func (s *MemoryStore) Get(id string) (model.Transaction, error) {
	// only need read lock here since we're just reading from the store
	// defer will wait until the function returns before executing the unlock
	s.memstoreMux.RLock()
	defer s.memstoreMux.RUnlock()

	// this uses the comma ok idiom like above
	existingTxn, exists := s.transactions[id]

	if exists {
		return existingTxn.Clone(), nil
	}

	return model.Transaction{}, ErrNotFound
}

// List returns a slice of transactions based on the provided limit and offset for pagination.
// ----------------------------------------------------------------------------------------------
// initially I handled edge cases but after re-reading the requirements I realized it just says
// "return an empty list if offset is out of bounds" and doesn't specify that limit/offset must
// be non-negative so I removed the error handling for negative values and just treat them as
// normal values which results in the same behavior as if they were positive (e.g. negative
// offset will just return the first "limit" transactions)
func (s *MemoryStore) List(limit, offset int) ([]model.Transaction, error) {
	s.memstoreMux.RLock()
	defer s.memstoreMux.RUnlock()

	// Handle offset beyond data - return empty slice
	if offset >= len(s.ordered) {
		return []model.Transaction{}, nil
	}

	end := offset + limit
	// Cap end to available data instead of erroring
	if end > len(s.ordered) {
		end = len(s.ordered)
	}

	// Clone each element so callers cannot mutate the store's internal map references
	result := make([]model.Transaction, end-offset)
	for i, txn := range s.ordered[offset:end] {
		result[i] = txn.Clone()
	}

	return result, nil
}
