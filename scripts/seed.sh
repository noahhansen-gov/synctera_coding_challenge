#!/bin/bash
# Seed script: posts 20 transactions for manual testing
# Covers: multiple currencies, date ranges, and amount ranges

BASE_URL="http://localhost:8080"

post() {
  local body="$1"
  local label="$2"
  response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/transactions" \
    -H "Content-Type: application/json" \
    -d "$body")
  echo "[$response] $label"
}

# USD transactions - low amounts (Jan 2024)
post '{"id":"txn-001","amount":500,"currency":"USD","effective_at":"2024-01-05T10:00:00Z"}' "USD $5.00 - Jan 5"
post '{"id":"txn-002","amount":1200,"currency":"USD","effective_at":"2024-01-10T14:30:00Z"}' "USD $12.00 - Jan 10"
post '{"id":"txn-003","amount":750,"currency":"USD","effective_at":"2024-01-15T09:00:00Z"}' "USD $7.50 - Jan 15"
post '{"id":"txn-004","amount":300,"currency":"USD","effective_at":"2024-01-20T16:00:00Z"}' "USD $3.00 - Jan 20"
post '{"id":"txn-005","amount":999,"currency":"USD","effective_at":"2024-01-25T11:00:00Z"}' "USD $9.99 - Jan 25"

# USD transactions - high amounts (Feb 2024)
post '{"id":"txn-006","amount":50000,"currency":"USD","effective_at":"2024-02-01T08:00:00Z"}' "USD $500.00 - Feb 1"
post '{"id":"txn-007","amount":75000,"currency":"USD","effective_at":"2024-02-14T12:00:00Z"}' "USD $750.00 - Feb 14"
post '{"id":"txn-008","amount":100000,"currency":"USD","effective_at":"2024-02-28T17:00:00Z"}' "USD $1000.00 - Feb 28"

# EUR transactions (Mar 2024)
post '{"id":"txn-009","amount":2500,"currency":"EUR","effective_at":"2024-03-01T10:00:00Z"}' "EUR $25.00 - Mar 1"
post '{"id":"txn-010","amount":8000,"currency":"EUR","effective_at":"2024-03-10T13:00:00Z"}' "EUR $80.00 - Mar 10"
post '{"id":"txn-011","amount":15000,"currency":"EUR","effective_at":"2024-03-20T09:30:00Z"}' "EUR $150.00 - Mar 20"
post '{"id":"txn-012","amount":45000,"currency":"EUR","effective_at":"2024-03-31T23:59:00Z"}' "EUR $450.00 - Mar 31"

# GBP transactions (Apr 2024)
post '{"id":"txn-013","amount":1000,"currency":"GBP","effective_at":"2024-04-05T10:00:00Z"}' "GBP $10.00 - Apr 5"
post '{"id":"txn-014","amount":3500,"currency":"GBP","effective_at":"2024-04-15T14:00:00Z"}' "GBP $35.00 - Apr 15"
post '{"id":"txn-015","amount":22000,"currency":"GBP","effective_at":"2024-04-25T16:00:00Z"}' "GBP $220.00 - Apr 25"

# Same timestamp (tests tie-breaking by ID)
post '{"id":"txn-016","amount":5000,"currency":"USD","effective_at":"2024-05-01T12:00:00Z"}' "USD same-ts A"
post '{"id":"txn-017","amount":6000,"currency":"EUR","effective_at":"2024-05-01T12:00:00Z"}' "EUR same-ts B"
post '{"id":"txn-018","amount":7000,"currency":"GBP","effective_at":"2024-05-01T12:00:00Z"}' "GBP same-ts C"

# Transaction with metadata
post '{"id":"txn-019","amount":9999,"currency":"USD","effective_at":"2024-06-01T10:00:00Z","metadata":{"source":"mobile","user_id":"u-42"}}' "USD with metadata"

# Zero amount (edge case)
post '{"id":"txn-020","amount":0,"currency":"USD","effective_at":"2024-06-15T10:00:00Z"}' "USD zero amount"

echo ""
echo "Done. Try these queries:"
echo "  curl \"$BASE_URL/transactions\""
echo "  curl \"$BASE_URL/transactions?currency=EUR\""
echo "  curl \"$BASE_URL/transactions?start_date=2024-02-01&end_date=2024-03-31\""
echo "  curl \"$BASE_URL/transactions?min_amount=10000&max_amount=50000\""
echo "  curl \"$BASE_URL/transactions?limit=5&offset=0\""
echo "  curl \"$BASE_URL/transactions?limit=5&offset=5\""
