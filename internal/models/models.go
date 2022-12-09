// Package models represents service common models.
package models

import (
	"fmt"
)

// Money model.
type Money struct {
	Amount   float64
	Currency string
}

// NewMoney constructor for Money.
func NewMoney(amount float64, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

func (r Money) String() string {
	return fmt.Sprintf("%.2f %s", r.Amount, r.Currency)
}
