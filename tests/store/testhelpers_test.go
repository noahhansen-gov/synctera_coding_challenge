package store_test

import (
	"time"

	"github.com/synctera/tech-challenge/internal/model"
)

func makeTxn(id string, amount int64, currency string, effectiveAt time.Time) model.Transaction {
	return model.Transaction{
		ID:          id,
		Amount:      amount,
		Currency:    currency,
		EffectiveAt: effectiveAt,
	}
}

func jan(day int) time.Time {
	return time.Date(2024, time.January, day, 0, 0, 0, 0, time.UTC)
}
