package moneyutils

import (
	"github.com/shopspring/decimal"
)

// Multiply returns result of multiplication of two float64.
func Multiply(a, b float64) float64 {
	d := multiply(decimal.NewFromFloat(a), decimal.NewFromFloat(b))

	return d.InexactFloat64()
}

func multiply(a, b decimal.Decimal) decimal.Decimal {
	res := a.Mul(b)

	return res
}

// Round rounds the decimal to places decimal places.
// If places < 0, it will round the integer part to the nearest 10^(-places).
func Round(a float64, places int32) float64 {
	rounded := round(decimal.NewFromFloat(a), places)

	return rounded.InexactFloat64()
}

func Parse(raw string) (float64, error) {
	d, err := decimal.NewFromString(raw)
	if err != nil {
		return 0, err
	}

	return d.InexactFloat64(), nil
}

func ToString(v float64) string {
	d := decimal.NewFromFloat(v)

	return d.String()
}

func round(amount decimal.Decimal, places int32) decimal.Decimal {
	rounded := amount.Round(places)

	return rounded
}
