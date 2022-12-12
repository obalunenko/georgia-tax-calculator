// Package moneyutils provide functionality for work with money.
package moneyutils

import (
	"github.com/shopspring/decimal"
)

// Multiply returns result of multiplication of two float64.
func Multiply(a, b float64) float64 {
	d := multiply(decimal.NewFromFloat(a), decimal.NewFromFloat(b))

	return d.InexactFloat64()
}

// Div returns result of div of two float64.
func Div(a, b float64) float64 {
	d := div(decimal.NewFromFloat(a), decimal.NewFromFloat(b))

	return d.InexactFloat64()
}

// Add returns sum of two floats.
func Add(a, b float64) float64 {
	s := add(decimal.NewFromFloat(a), decimal.NewFromFloat(b))

	return s.InexactFloat64()
}

func add(a, b decimal.Decimal) decimal.Decimal {
	return a.Add(b)
}
func div(a, b decimal.Decimal) decimal.Decimal {
	return a.Div(b)
}

func multiply(a, b decimal.Decimal) decimal.Decimal {
	return a.Mul(b)
}

// Round rounds the decimal to places decimal places.
// If places < 0, it will round the integer part to the nearest 10^(-places).
func Round(a float64, places int32) float64 {
	rounded := round(decimal.NewFromFloat(a), places)

	return rounded.InexactFloat64()
}

// Parse float from string.
func Parse(raw string) (float64, error) {
	d, err := decimal.NewFromString(raw)
	if err != nil {
		return 0, err
	}

	return d.InexactFloat64(), nil
}

// ToString converts float to string.
func ToString(v float64) string {
	d := decimal.NewFromFloat(v)

	return d.String()
}

func round(amount decimal.Decimal, places int32) decimal.Decimal {
	rounded := amount.Round(places)

	return rounded
}
