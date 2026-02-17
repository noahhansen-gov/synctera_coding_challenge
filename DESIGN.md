# Design Notes

## Assumptions

- Amount is in minor units (i.e. cents), so it is stored as int64. No floating point or rounding errors.
- effective_at is the business timestamp, not the ingestion time. Not tracking when a transaction arrived, only when it occurred.
- Idempotency is client-driven via the transaction ID. A retry with the same ID and identical payload succeeds silently (HTTP 200). A retry with the same ID but different data is a conflict (HTTP 409). This matches how real payment systems could handle retries.
- Transactions are immutable once accepted. There is no PATCH or DELETE endpoint.
- Currency filtering is case-insensitive (usd and USD match the same transactions).
- Date filters use day-level granularity in YYYY-MM-DD format. The end date is inclusive.
- No authentication or authorization is required.
- Read-heavy API due to transactions being written once but queried repeatedly for reporting, reconciliation, and audit.
- metadata is optional and free-form. It is included in the idempotency check talked about above.

## Tradeoffs

- In-memory store over a real database. Keeps the implementation simple and self-contained. The Store interface (Create, Get, List) abstracts this away so the storage backend can be swapped without touching handler code.
- Sorted slice maintained on insert, not on read. The ordered slice is kept in sorted order at write time using a binary search to find the insertion point. This makes reads O(1) slice operations, at the cost of O(n) insert time due to element shifting.
- Dual data structure. The store holds both Transaction map for O(1) ID lookups (used by Get and the idempotency check in Create) and a sorted Transaction array for ordered queries (used by List). The memory overhead is worth the performance clarity.
- Filters applied in-memory after fetching. ListTransactions fetches up to 10,000 records and filters in Go code. In production, filter predicates would be pushed down to the database as SQL WHERE clauses with indexes. The current approach is correct but does not scale.
- limit/offset pagination is simple to implement and reason about, but has a known flaw: if new transactions are inserted between page requests, results can shift. Cursor-based pagination (using the last-seen effective_at + id as a bookmark) would be stable across pages.
- No request body size limit. The JSON decoder will read whatever the client sends. In production this should be capped to prevent memory exhaustion from malicious or oversized payloads.
- No structured logging or request IDs. Errors surface as plain-text HTTP responses. In production every request would carry a trace ID and errors would be logged as structured JSON.

## Scaling

- Memory. All transactions live in RAM. With no eviction or persistence, the store will eventually OOM. This is the first thing that breaks under sustained load.
- O(n) insert due to slice shifting. Inserting into the middle of the ordered slice requires copying all subsequent elements. At millions of transactions this degrades write throughput noticeably. A skip list or B-tree would give O(log n) inserts while preserving sorted order.
- O(n) full-scan filtering. Every GET /transactions with filters scans up to 10,000 records in memory. As data grows this gets slower and the 10,000 cap becomes either a correctness problem (missing results) or must be raised (costing more memory per request).
- No horizontal scaling. State is in-process, so you cannot run multiple instances behind a load balancer. Any real deployment would need the store backed by a shared external system (database, cache).

## Evolution

Moving to a real database:

- Swap the implementation, not the interface. A PostgresStore implementing Create, Get, and List drops in without changing any handler code.
- Filters become SQL WHERE clauses. The in-memory scan goes away entirely.
- Ordering is handled by the database.
- Idempotency uses a UNIQUE constraint on the id so there is duplicate detection at the database level.
- Pagination shifts to keyset. Instead of offset, use a WHERE clause for stable, index-friendly pagination that does not degrade at large offsets.
- Persistence and durability come for free. Transactions survive restarts. Write-ahead logging protects against data loss.

## Observability

In a production version I would:

- Track Error rate by status code. A spike in 409s (conflicts) suggests a client retry bug. A spike in 400s suggests a schema change broke a caller. A spike in 500s means something is broken internally.
- Create a Transaction count gauge. An increasing counter of stored transactions tells you how fast the store is growing and helps anticipate when you will hit memory limits.
- Develop Structured logging with a request ID. Every request should log a unique ID, method, path, status code, and duration.

## What I'd Do Next

- Replace limit/offset with cursor-based pagination for correctness under concurrent writes.
- Add a request body size cap to guard against oversized payloads.
- Add a PostgresStore implementation behind the Store interface.
- Add structured logging (e.g., log/slog) with request IDs.
- Expose a /health endpoint for readiness/liveness probes.
