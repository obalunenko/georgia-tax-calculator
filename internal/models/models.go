package models

import (
	"fmt"
)

type ResultOutput struct {
	Message string
	Money
}

func NewResultOutput(msg string, m Money) ResultOutput {
	return ResultOutput{
		Message: msg,
		Money:   m,
	}
}

func (r ResultOutput) String() string {
	return fmt.Sprintf("%s: %s\n", r.Message, r.Money.String())
}

type Money struct {
	Amount   float64
	Currency string
}

func NewMoney(amount float64, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

func (r Money) String() string {
	return fmt.Sprintf("%.2f %s", r.Amount, r.Currency)
}
