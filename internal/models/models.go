package models

import (
	"fmt"
)

// ResultOutput model.
type ResultOutput struct {
	Message string
	Money
}

// NewResultOutput constructor for ResultOutput.
func NewResultOutput(msg string, m Money) ResultOutput {
	return ResultOutput{
		Message: msg,
		Money:   m,
	}
}

func (r ResultOutput) String() string {
	return fmt.Sprintf("%s: %s", r.Message, r.Money.String())
}

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
