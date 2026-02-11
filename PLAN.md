# Development Plan

## Overview
This plan outlines the implementation approach for the Synctera Backend Tech Challenge - building an HTTP service for financial transaction ingestion and querying.

**Estimated Time:** 2-4 hours
**Focus:** Correctness, testability, and clear reasoning over feature count

### Current State of the Codebase

**Already implemented:**
- ✅ Project structure with `cmd/` and `internal/` layout
- ✅ Go module setup (`go.mod` with go 1.22)
- ✅ `model.Transaction` struct with all required fields (id, amount, currency, effective_at) plus optional metadata
- ✅ `Transaction.Equal()` method for idempotency checks (compares all fields including metadata)
- ✅ `store.Store` interface defining Create, Get, List methods
- ✅ Store errors: `ErrNotFound`, `ErrConflict`, `ErrDuplicate` (as typed constants)
- ✅ HTTP server setup in `main.go` with routing for POST/GET `/transactions` and `/health` endpoint
- ✅ `api.Handler` struct with method signatures (not implemented)
- ✅ `store.MemoryStore` struct skeleton (empty implementation)

**What needs to be built:**
- ❌ MemoryStore implementation (data structures + all methods)
- ❌ API handler implementations (CreateTransaction, ListTransactions)
- ❌ Input validation and error handling
- ❌ Tests (no test files exist yet)
- ❌ DESIGN.md documentation

**Key insight:** The starter code already handles routing and has well-defined interfaces. The implementation just needs to fill in the TODOs.

---

## Phase 1: Core Storage Implementation
**Goal:** Build the in-memory transaction store with proper concurrency control

### 1. Implement MemoryStore data structures

**What to build:**
```go
type MemoryStore struct {
    mu           sync.RWMutex
    transactions map[string]model.Transaction  // Fast O(1) lookups by ID
    ordered      []model.Transaction            // Maintains sorted order for queries
}
```

**Why these choices:**

- **Map for ID-based storage:**
  - Provides O(1) lookup time for the Get operation
  - Enables fast duplicate detection during Create (critical for idempotency)
  - Natural fit for the "transaction ID is globally unique" requirement
  - Allows instant conflict detection when same ID arrives with different payload

- **Slice for ordered storage:**
  - List operation needs to return transactions in deterministic order
  - Sorting on every query is expensive - pre-maintaining order is more efficient
  - Enables efficient pagination with simple slice operations
  - Use `sort.Search()` to find insertion point and `sort.Slice()` for custom comparison
  - Binary search for insertion: O(log n) find + O(n) insert (acceptable for in-memory store)
  - Alternative would be sorting on every GET request, but that's wasteful for read-heavy workloads

- **Mutex for thread-safety:**
  - HTTP servers handle requests concurrently via goroutines
  - Without synchronization, concurrent Create operations could cause race conditions
  - RWMutex allows multiple concurrent readers (List/Get) but exclusive writers (Create)
  - This is critical for financial data - race conditions could lead to data corruption
  - Go's race detector would catch missing synchronization in tests

**Why not other approaches:**
- ❌ Single slice only: Would require O(n) linear search for duplicate detection
- ❌ No mutex: Would fail under concurrent load (data races, lost writes)
- ❌ Channels: Over-engineered for this use case; mutex is simpler and sufficient

---

### 2. Implement Create method

**What to build:**
```go
func (s *MemoryStore) Create(txn model.Transaction) error {
    // 1. Acquire write lock
    // 2. Check if ID exists
    // 3. If exists, compare payloads (idempotency vs conflict)
    // 4. If new, insert into map and maintain sorted order in slice
    // 5. Return appropriate error
}
```

**Why these implementation details:**

- **Idempotency handling (same ID + same payload = success):**
  - Financial systems must handle retries gracefully
  - Network failures are common; clients retry requests
  - If same transaction arrives twice (identical payload), it's a retry - not an error
  - This prevents "duplicate transaction" errors when network hiccups occur
  - Use the `Transaction.Equal()` method already provided in the model
  - Return `ErrDuplicate` to signal this is an idempotent retry (handler will treat as success)
  - Alternative: return nil, but then handler can't distinguish 201 vs 200 responses

- **Conflict detection (same ID + different payload = error):**
  - Transaction IDs are supposed to be globally unique
  - If same ID arrives with different data, something is wrong
  - Could be client bug, ID collision, or malicious activity
  - Must reject to maintain data integrity
  - Return `ErrConflict` to signal this specific error case
  - This protects against accidental overwrites of immutable transactions

- - **Dual storage (map + ordered slice):**
  - Map insertion: O(1) for future lookups
  - Slice insertion: Find correct position via binary search, insert to maintain order
  - Both must stay in sync - transaction should appear in both or neither
  - This duplication trades memory for query performance

- **Note on the Metadata field:**
  - Transaction already has `Metadata map[string]string` field
  - The Equal() method already handles metadata comparison correctly
  - No special handling needed beyond what's in the model

**Why this matters in financial systems:**
- Transactions are immutable once accepted (stated in requirements)
- Retries are expected behavior in distributed systems
- Rejecting valid retries would break client integrations
- Accepting conflicting data would violate data integrity

---

### 3. Implement Get method

**What to build:**
```go
func (s *MemoryStore) Get(id string) (model.Transaction, error) {
    // 1. Acquire read lock
    // 2. Look up in map
    // 3. Return transaction or ErrNotFound
}
```

**Why this approach:**
- Simple O(1) lookup using the map
- RWMutex allows concurrent Get operations (multiple readers don't block each other)
- Return `ErrNotFound` for missing IDs (enables proper HTTP 404 responses)
- This operation is likely frequent for reconciliation, debugging, and audit trails

---

### 4. Implement List method

**What to build:**
```go
func (s *MemoryStore) List(limit, offset int) ([]model.Transaction, error) {
    // 1. Acquire read lock
    // 2. Validate pagination parameters
    // 3. Slice the ordered array [offset:offset+limit]
    // 4. Return subset
}
```

**Why deterministic ordering matters:**

- **Primary sort: `effective_at` (ascending):**
  - Business logic: transactions ordered by when they take effect
  - Most queries want chronological order
  - Matches how accountants and auditors think about transactions
  - "Show me all transactions from last week" is a natural query

- **Secondary sort: `id` (lexicographic):**
  - Two transactions can have identical timestamps
  - Without secondary sort, order would be undefined (random/unstable)
  - Unstable ordering breaks pagination (item could appear on multiple pages)
  - Breaks client caching and reconciliation
  - String comparison is deterministic and stable
  - Alternative: use insertion order, but that's harder to reproduce after restart

- **Why consistent ordering is critical:**
  - Clients paginating through results need stability
  - If order changes between page 1 and page 2, records could be skipped or duplicated
  - Financial reconciliation requires repeatable queries
  - Auditors need to see the same results when re-running queries

**Pagination with limit/offset:**
- Simple and stateless (no cursor to manage)
- Works with any ordering scheme
- Easy to implement: just array slicing
- Downside: inefficient for large offsets (must skip N items)
- For this challenge, simplicity wins over optimization

---

## Phase 2: API Handlers
**Goal:** Implement HTTP endpoints with proper validation and error handling

### 5. Implement CreateTransaction handler (POST /transactions)

**What to build:**
```go
func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
    // 1. Parse JSON body into Transaction struct
    // 2. Validate required fields
    // 3. Call store.Create()
    // 4. Map store errors to HTTP status codes
    // 5. Return appropriate response
}
```

**Why each step matters:**

#### JSON Parsing and Validation

**Parse JSON body:**
```go
var txn model.Transaction
if err := json.NewDecoder(r.Body).Decode(&txn); err != nil {
    http.Error(w, "invalid JSON", http.StatusBadRequest)
    return
}
```

- **Why:** Clients send data as JSON (REST API standard)
- **Why Decoder:** Streaming parser, memory efficient for large payloads
- **Why 400 on parse failure:** Malformed JSON is client error, not server error
- Prevents garbage data from reaching business logic

**Validate required fields:**
```go
if txn.ID == "" || txn.Currency == "" || txn.EffectiveAt.IsZero() {
    http.Error(w, "missing required fields", http.StatusBadRequest)
    return
}
```

- **Why validate ID:** Empty ID would break map lookups and uniqueness checks
- **Why validate currency:** Business requirement; transactions need currency context
- **Why validate timestamp:** Zero time (0001-01-01) indicates missing/invalid timestamp
- **Why validate amount:** Might allow zero for refunds/reversals (business decision)
- **Fail fast principle:** Reject bad data before it reaches storage layer
- Better error messages help client developers debug integration issues

#### HTTP Status Code Mapping

**Why specific status codes matter:**

**201 Created - New transaction accepted:**
```go
w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(txn)
```
- Standard HTTP semantics for resource creation
- Tells client: "Transaction recorded successfully"
- Can return the transaction back for confirmation (includes server-side defaults if any)
- RESTful convention: POST that creates returns 201

**200 OK - Idempotent retry (same payload):**
```go
if errors.Is(err, store.ErrDuplicate) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(txn)
    return
}
```
- **Why 200 not 201:** Resource already exists, not newly created
- **Why not an error:** Idempotency means retries should succeed
- **Why return transaction:** Client gets confirmation their data is stored
- Critical for distributed systems: network failures cause retries
- Makes API safe to call multiple times with same data
- Without this, clients would see errors on legitimate retries

**409 Conflict - Same ID, different payload:**
```go
if errors.Is(err, store.ErrConflict) {
    http.Error(w, "transaction ID already exists with different data", http.StatusConflict)
    return
}
```
- **Why 409:** Semantic meaning is "resource state conflict"
- **Why not 400:** This isn't a validation error; the ID itself is valid
- **Why not 500:** Not a server error; this is a business rule violation
- Tells client: "This ID is taken; use a different one"
- In financial systems, prevents accidental overwrites of transaction records
- Could indicate client bug (poor ID generation) or malicious activity

**400 Bad Request - Validation errors:**
```go
if txn.ID == "" {
    http.Error(w, "id is required", http.StatusBadRequest)
    return
}
```
- **Why 400:** Client sent invalid data
- **Why specific messages:** Help client developers fix their integration
- Could be missing fields, invalid formats, or business rule violations
- Clear errors reduce support burden and integration time

#### Why Not Other Approaches?

- ❌ **Always return 200:** Loses semantic meaning; client can't distinguish success types
- ❌ **Return 500 for conflicts:** Implies server error; client would retry forever
- ❌ **No validation:** Garbage in, garbage out; debugging nightmares later
- ❌ **Accept duplicate ID silently:** Violates immutability; data integrity risk

---

### 6. Implement ListTransactions handler (GET /transactions)

**What to build:**
```go
func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
    // 1. Parse query parameters (limit, offset)
    // 2. Set default values
    // 3. Validate parameters
    // 4. Call store.List()
    // 5. Return JSON array
}
```

**Why each design choice matters:**

#### Query Parameter Parsing

**Parse limit and offset:**
```go
limit := parseIntOrDefault(r.URL.Query().Get("limit"), 100)
offset := parseIntOrDefault(r.URL.Query().Get("offset"), 0)
```

- **Why query parameters:** RESTful convention for GET requests
- **Why not request body:** GET requests shouldn't have bodies (HTTP semantics)
- **Why parse to int:** Type safety; prevents injection attacks
- **Why default limit:** Unbounded queries could return millions of records
- **Why default offset of 0:** Start at beginning if not specified

#### Default Values - Why They Matter

**Default limit = 100:**
- **Why not unlimited:** Protects server from memory exhaustion
- **Why not 10:** Too small for practical use; too many round trips
- **Why 100:** Balances response size vs round trips
- **Real production:** Would make this configurable with max cap (e.g., 1000)
- Prevents accidental DoS from `GET /transactions` without params

**Default offset = 0:**
- Natural starting point (first page)
- Makes `GET /transactions` return first page by default
- Allows simple pagination: `?offset=0`, `?offset=100`, `?offset=200`

#### Input Validation

**Validate pagination parameters:**
```go
if limit < 0 || limit > 1000 {
    http.Error(w, "limit must be between 0 and 1000", http.StatusBadRequest)
    return
}
if offset < 0 {
    http.Error(w, "offset must be non-negative", http.StatusBadRequest)
    return
}
```

- **Why validate limit:** Prevent client from requesting 10M records
- **Why max limit:** Cap protects server resources
- **Why validate offset:** Negative offset is nonsensical
- **Why return 400:** Client error, not server error
- Financial systems need predictable performance

#### Response Format

**Return JSON array:**
```go
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(transactions)
```

- **Why JSON array:** Standard for collection responses
- **Why set Content-Type:** Tells client how to parse response
- **Why Encoder:** Streams JSON output, memory efficient

**Optional: Add pagination metadata:**
```go
response := struct {
    Data   []model.Transaction `json:"data"`
    Total  int                 `json:"total"`
    Limit  int                 `json:"limit"`
    Offset int                 `json:"offset"`
}{
    Data:   transactions,
    Total:  len(store.transactions), // total records (requires exposing or adding Count())
    Limit:  limit,
    Offset: offset,
}
```

- **Why total count:** Client can calculate total pages
- **Why include limit/offset:** Confirms what was used (debugging)
- **Why wrap in object:** Allows adding metadata without breaking existing clients
- **Tradeoff:** Requires adding Count() method to Store interface or exposing field
- **For MVP:** Can skip metadata and just return `[]model.Transaction` directly (simpler)

#### Why These Query Patterns?

**Simple offset/limit pagination:**
- **Pros:** Stateless, no server-side cursor storage, jump to any page
- **Cons:** Inefficient for large offsets, inconsistent if data changes mid-pagination
- **Why acceptable here:** In-memory store makes offset cheap; good enough for MVP

**Alternative (cursor-based pagination):**
- **Pros:** Efficient for large datasets, consistent even if data changes
- **Cons:** More complex, can't jump to page N, requires cursor management
- **Why not chosen:** Over-engineered for this challenge; offset is simpler

#### Error Handling Philosophy

**Return clear, actionable errors:**
```go
http.Error(w, "limit must be between 1 and 1000", http.StatusBadRequest)
// Not: "invalid request"
```

- **Why specific messages:** Client developers can fix issues without support
- **Why include constraints:** Documents API contract in error messages
- **Why not stack traces:** Security risk; internal details leak
- Good errors reduce integration time and support costs

---

**Summary - Why These Patterns Matter:**

1. **Validation:** Fail fast; prevent bad data from propagating
2. **Status codes:** Semantic HTTP enables proper client behavior (retry logic, error handling)
3. **Idempotency:** Makes API resilient to network failures and retries
4. **Defaults:** Protect server while keeping API ergonomic
5. **Clear errors:** Reduce integration friction and support burden

These aren't just "best practices" - they directly impact system reliability, client experience, and operational costs in production.

---

## Phase 3: Testing
**Goal:** Prove correctness through comprehensive tests

### 7. Unit tests for MemoryStore

**File:** `internal/store/memory_test.go`

- Test basic CRUD operations
- Test idempotency (same payload) - should return ErrDuplicate
- Test conflicts (different payload) - should return ErrConflict
- Test concurrent access (goroutines) - use `go test -race`
- Test ordering and pagination
- Test boundary conditions (empty store, single item, etc.)

### 8. Integration tests for API handlers

**File:** `internal/api/handlers_test.go`

- Test POST endpoint success cases (201 response)
- Test POST idempotency (200 response for duplicate)
- Test POST conflicts (409 response for different payload)
- Test POST validation errors (400 response)
- Test GET endpoint with pagination
- Test GET with filtering (if implemented)
- Test ordering determinism
- Use `httptest.NewRecorder()` for testing HTTP handlers

### 9. Edge case tests
- Transactions with same timestamp
- Out-of-order arrival
- Large payloads
- Boundary conditions (offset, limit)

---

## Phase 4: Enhancements (Optional)
**Goal:** Add polish and advanced features if time permits

### 10. Query filtering
- Filter by date range (start_date, end_date)
- Filter by currency
- Filter by amount range

### 11. Additional validation
- Currency code validation (ISO 4217)
- Amount validation (non-zero check based on requirements)
- Timestamp validation (not too far in future)

### 12. Error handling improvements
- Structured error responses (JSON with error details)
- Request ID tracking for debugging

---

## Phase 5: Documentation
**Goal:** Complete the design documentation

### 13. Fill out DESIGN.md with:
- **Assumptions** - What was assumed during implementation
- **Tradeoffs** - What was simplified or skipped and why
- **Scaling** - What breaks first under load
- **Evolution** - How this would work with a real database
- **Observability** - What metrics/logs to add
- **What you'd do next** - If ran out of time

---

## Testing & Validation

### 14. Run and verify
- `go test ./...` - ensure all tests pass
- `go run ./cmd/server` - manual testing with curl
- Test the happy path end-to-end
- Test error conditions

---

## Key Design Decisions to Consider

### Idempotency Strategy
- Same ID + same payload = success (200)
- Same ID + different payload = conflict (409)

### Ordering
- Primary sort by `effective_at` (ascending)
- Secondary sort by `id` (lexicographic) for determinism
- Ensures predictable results even with identical timestamps

### Concurrency
- Use mutex for thread-safety in MemoryStore
- Protect both read and write operations

### Pagination
- Limit/offset approach (simple, stateless)
- Document tradeoffs vs cursor-based pagination

### Validation
- Fail fast on required fields
- Be lenient on optional fields
- Return clear error messages

---

## Success Criteria

- ✅ All tests pass
- ✅ Idempotency works correctly
- ✅ Deterministic ordering with pagination
- ✅ Proper error handling and HTTP status codes
- ✅ DESIGN.md filled out with thoughtful answers
- ✅ Code is readable and maintainable

---

## Time Management

If running short on time, prioritize in this order:
1. Core functionality (Create + List with basic pagination)
2. Idempotency handling
3. Basic tests proving it works
4. DESIGN.md documentation
5. Advanced features (filtering, extensive validation)
