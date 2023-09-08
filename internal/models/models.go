// Package models represents service common models.
package models

import (
	"fmt"
	"strconv"
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
	a := strconv.FormatFloat(r.Amount, 'f', -1, 64)
	if r.Currency == "" {
		return a
	}

	return fmt.Sprintf("%s %s", a, r.Currency)
}
