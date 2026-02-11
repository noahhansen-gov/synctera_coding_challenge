# Synctera Backend Tech Challenge

## Overview

Build an HTTP service that ingests and queries financial transactions.

**Time expectation:** 2–4 hours. If you run out of time, document what you'd do next.

**On using AI:** Feel free to use AI tools to help with your implementation. In the follow-up interview, we'll discuss your solution and the decisions you made, so be prepared to explain your approach.

**Expectation:** This is not just a spec for you to implement, we care more about your thinking and approach in the design space, feel free to mention whatever you think is important in a real production system.

## Domain Context

You're building a simplified transaction ingestion system. Some things to keep in mind:

- Transactions are immutable once accepted
- Each transaction has a globally unique ID, but clients may retry requests
- Transactions may arrive out of order relative to their timestamps
- Query results should be predictable

This is not a full ledger or balance engine. No authentication, real database, or UI required.

## The Problem

### 1. Transaction Ingestion

Implement `POST /transactions` to accept and store transactions.

Consider how to handle, for example:
- Requests where the transaction ID has already been seen
- Requests where the same ID is submitted with a different payload
- ...

### 2. Transaction Querying

Implement `GET /transactions` to retrieve stored transactions.

Consider how to handle, for example:
- Consistent, deterministic ordering
- Transactions that share the same timestamp
- Pagination
- Filtering
- ...

## Transaction Schema

At minimum, a transaction should have:
- `id` (string) - client-provided identifier
- `amount` (integer) - minor units (e.g., cents)
- `currency` (string) - e.g., "USD"
- `effective_at` (timestamp)

You may add other fields if useful.

## What We're Looking For

- **Correctness** - Does it work? Are edge cases handled?
- **Tests** - How do you prove it works?
- **Reasoning** - What tradeoffs did you make and why?

We care more about your thinking than feature count.

## Getting Started

```bash
go run ./cmd/server
go test ./...
```

The starter code provides basic project structure. You may modify anything.

## Written Section

Fill out `DESIGN.md` with brief answers to:

1. **Assumptions** - What did you assume?
2. **Tradeoffs** - What did you skip or simplify?
3. **Scaling** - What breaks first under load?
4. **Evolution** - How would this change with a real database?
5. **Observability** - What would you instrument first?

Bullet points are fine. If you ran out of time, note what you'd do next.

## Submission

1. Complete your implementation
2. Ensure your tests pass
3. Fill out `DESIGN.md`
