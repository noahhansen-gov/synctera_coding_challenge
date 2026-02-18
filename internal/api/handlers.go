package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/synctera/tech-challenge/internal/model"
	"github.com/synctera/tech-challenge/internal/store"
)

type Handler struct {
	store store.Store
}

func NewHandler(s store.Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    if id == "" {
        http.Error(w, "missing transaction id", http.StatusBadRequest)
        return
    }

    txn, err := h.store.Get(id)
    if errors.Is(err, store.ErrNotFound) {
        http.Error(w, "transaction not found", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(txn)
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
    var txn model.Transaction

    // Parse JSON
    if err := json.NewDecoder(r.Body).Decode(&txn); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }

    // Validate required fields
    if err := ValidateTransaction(txn); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Call the store and create the transaction
    err := h.store.Create(txn)

    // Handle errors from store
    if errors.Is(err, store.ErrDuplicate) {
        // Idempotent retry - same transaction already exists
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(txn)
        return
    } else if errors.Is(err, store.ErrConflict) {
        // Same ID, different data - conflict
        http.Error(w, "transaction ID already exists with different data", http.StatusConflict)
        return
    } else if err != nil {
        // Some other error
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }

    // 5. Success - new transaction created
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(txn)
}


func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query()

    // Parse query parameters (no pre-declaration needed)
    limit, offset, currency,
	   startDateStr, endDateStr,
	   minAmountStr, maxAmountStr := parseQueryParams(query)

	// Validate pagination parameters
	if err := ValidatePagination(limit, offset); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse and validate date filters
	startDate, endDate, err := ParseAndValidateDateFilters(startDateStr, endDateStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse and validate amount filters
	minAmount, maxAmount, err := ParseAndValidateAmountFilters(minAmountStr, maxAmountStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// For now, get a large batch to filter from
	// In production, filters would be pushed down to the database
	maxRecords := 10000 // Reasonable limit for in-memory filtering
	allTransactions, err := h.store.List(maxRecords, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Apply filters to the retrieved transactions
	filtered := ApplyFilters(allTransactions, currency, startDate, endDate, minAmount, maxAmount)

	// Apply pagination to the filtered results
	results := ApplyPagination(filtered, limit, offset)

	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Return JSON array
	json.NewEncoder(w).Encode(results)
}

// EXPORTED HELPER FUNCTIONS
// These are exported (uppercase) so they can be tested from the external tests/api/ package.
// This is safe because internal/ packages cannot be imported from outside this module.

// ValidateTransaction validates the transaction fields before attempting to store it.
func ValidateTransaction(txn model.Transaction) error {
	switch {
	case txn.ID == "":
		return errors.New("id is required")
	case txn.Currency == "":
		return errors.New("currency is required")
	case txn.Amount < 0:
		return errors.New("amount must be positive")
	case txn.EffectiveAt.IsZero():
		return errors.New("effective_at is required")
	}
	return nil
}

// ValidatePagination checks that the limit and offset parameters are within acceptable ranges.
func ValidatePagination(limit, offset int) error {
    if limit < 1 || limit > 1000 {
        return errors.New("limit must be between 1 and 1000")
    }
    if offset < 0 {
        return errors.New("offset must be non-negative")
    }
    return nil
}

// ParseIntOrDefault parses an integer query parameter,
// returning the default value if the string is empty or invalid.
func ParseIntOrDefault(s string, defaultVal int) int {
    if s == "" {
        return defaultVal
    }
    val, err := strconv.Atoi(s)
    if err != nil {
        return defaultVal
    }
    return val
}

// ParseDateOrNil parses a YYYY-MM-DD date string into a *time.Time.
// Returns nil,nil for empty strings (meaning "no filter").
func ParseDateOrNil(dateStr string) (*time.Time, error) {
    if dateStr == "" {
        return nil, nil // No filter provided
    }

    // Parse using ISO 8601 date format (YYYY-MM-DD)
    t, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        return nil, err
    }
    return &t, nil
}

// ParseAndValidateDateFilters parses and validates the start_date and end_date
// query parameters and returns pointers to time.Time values.
func ParseAndValidateDateFilters(startDateStr, endDateStr string) (*time.Time, *time.Time, error) {
	// Using pointers to distinguish between "not provided" (nil) and "provided with zero value" (time.Time{})
    var startDate, endDate *time.Time
    var err error

    if startDateStr != "" {
        startDate, err = ParseDateOrNil(startDateStr)
        if err != nil {
            return nil, nil, errors.New("invalid start_date format, use YYYY-MM-DD")
        }
    }

    if endDateStr != "" {
        endDate, err = ParseDateOrNil(endDateStr)
        if err != nil {
            return nil, nil, errors.New("invalid end_date format, use YYYY-MM-DD")
        }
    }

    if startDate != nil && endDate != nil && startDate.After(*endDate) {
        return nil, nil, errors.New("start_date must be before or equal to end_date")
    }

    return startDate, endDate, nil
}

// ParseAndValidateAmountFilters parses and validates the min_amount and max_amount
// query parameters, returning pointers to int64 values.
func ParseAndValidateAmountFilters(minAmountStr, maxAmountStr string) (*int64, *int64, error) {
	// Using pointers to distinguish between "not provided" (nil) and "provided with zero value" (0)
	// int64 is used for amounts to avoid overflow issues with large values
    var minAmount, maxAmount *int64

    if minAmountStr != "" {
        val, err := strconv.ParseInt(minAmountStr, 10, 64)
        if err != nil {
            return nil, nil, errors.New("invalid min_amount")
        }
        minAmount = &val
    }

    if maxAmountStr != "" {
        val, err := strconv.ParseInt(maxAmountStr, 10, 64)
        if err != nil {
            return nil, nil, errors.New("invalid max_amount")
        }
        maxAmount = &val
    }

    if minAmount != nil && maxAmount != nil && *minAmount > *maxAmount {
        return nil, nil, errors.New("min_amount must be less than or equal to max_amount")
    }

    return minAmount, maxAmount, nil
}

// ApplyFilters filters a slice of transactions based on optional currency, date, and amount constraints.
func ApplyFilters(transactions []model.Transaction, currency string, startDate, endDate *time.Time, minAmount, maxAmount *int64) []model.Transaction {
	// Create a new slice to hold the filtered transactions.
	// We can preallocate it with the same length as the input slice for efficiency
	filtered := make([]model.Transaction, 0, len(transactions))

	for _, txn := range transactions {
		// Continue to the next transaction if any of the filters do not match
		if currency != "" && !strings.EqualFold(txn.Currency, currency) {
			continue
		}
		if startDate != nil && txn.EffectiveAt.Before(*startDate) {
			continue
		}

		// Add 24 hours to endDate to include transactions that occur on the endDate up until 23:59:59
		// Check nil BEFORE dereferencing
		if endDate != nil {
			endOfDay := endDate.Add(24 * time.Hour)
			if txn.EffectiveAt.After(endOfDay) {
				continue
			}
		}

		if minAmount != nil && txn.Amount < *minAmount {
			continue
		}
		if maxAmount != nil && txn.Amount > *maxAmount {
			continue
		}
		filtered = append(filtered, txn)
	}

	return filtered
}

// ApplyPagination slices a transaction list to the requested page window.
func ApplyPagination(transactions []model.Transaction, limit, offset int) []model.Transaction {
    start := offset
	// Handle edge case where offset is greater than the number of transactions
    if start > len(transactions) {
        start = len(transactions)
    }

	// Calculate the end index for slicing, ensuring it does not exceed the length of the transactions slice
    end := start + limit
    if end > len(transactions) {
        end = len(transactions)
    }

    return transactions[start:end]
}

// parseQueryParams extracts all list query parameters from the URL values.
// Kept private as it is an internal detail of ListTransactions.
func parseQueryParams(query url.Values) (limit, offset int, currency, startDateStr, endDateStr, minAmountStr, maxAmountStr string) {
    limit = ParseIntOrDefault(query.Get("limit"), 100)
    offset = ParseIntOrDefault(query.Get("offset"), 0)
    currency = strings.ToUpper(query.Get("currency"))
    startDateStr = query.Get("start_date")
    endDateStr = query.Get("end_date")
    minAmountStr = query.Get("min_amount")
    maxAmountStr = query.Get("max_amount")
    return
}
