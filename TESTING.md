# Running the Tests

## Prerequisites

Go 1.22+ is required.

```bash
go version
```

## Run All Tests

```bash
go test ./...
```

## Run with Race Detector

Recommended to verify concurrency safety in the store layer:

```bash
go test ./... -race
```

## Run with Verbose Output

```bash
go test ./... -v
```

## Run a Specific Package

```bash
go test ./internal/model/...
go test ./internal/store/...
go test ./internal/api/...
```

## Run a Single Test

```bash
go test ./internal/api/... -run TestCreateTransaction_idempotentRetry
go test ./internal/store/... -run TestCreate_concurrent
```

## Test Structure

Tests are organized by package and split into focused files:

```
internal/
  model/
    transaction_test.go         # Transaction.Equal() logic

  store/
    testhelpers_test.go         # Shared helpers (makeTxn, jan)
    memory_create_test.go       # Create(): new, duplicate, conflict, concurrent writes
    memory_get_test.go          # Get(): found, not found, field values
    memory_list_test.go         # List(): ordering, pagination, copy safety

  api/
    testhelpers_test.go         # Shared helpers (newTestHandler, seedTxn, etc.)
    validate_test.go            # validateTransaction, validatePagination
    parse_test.go               # parseIntOrDefault, parseDateOrNil, date/amount filters
    filters_test.go             # applyFilters: currency, date range, amount range
    pagination_test.go          # applyPagination: offset, limit, page boundaries
    create_handler_test.go      # POST /transactions end-to-end
    list_handler_test.go        # GET /transactions end-to-end
```

## Manual Testing

Start the server:

```bash
go run ./cmd/server
```

Seed 20 test transactions:

```bash
bash scripts/seed.sh
```

Example queries (URL must be quoted in zsh):

```bash
# All transactions
curl "http://localhost:8080/transactions"

# Filter by currency
curl "http://localhost:8080/transactions?currency=EUR"

# Date range
curl "http://localhost:8080/transactions?start_date=2024-02-01&end_date=2024-02-28"

# Amount range
curl "http://localhost:8080/transactions?min_amount=1000&max_amount=50000"

# Pagination
curl "http://localhost:8080/transactions?limit=5&offset=0"
curl "http://localhost:8080/transactions?limit=5&offset=5"
```

> **Note:** The server uses in-memory storage. Restart it before re-running the seed script to avoid 409 conflicts on duplicate IDs.
