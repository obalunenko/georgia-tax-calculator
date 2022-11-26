package moneyutils

import (
	"github.com/shopspring/decimal"
)

func Multiply(a, b float64) float64 {
	d := multiply(decimal.NewFromFloat(a), decimal.NewFromFloat(b))

	return d.InexactFloat64()
}

func multiply(a, b decimal.Decimal) decimal.Decimal {
	res := a.Mul(b)

	return res
}

func Round(a float64, places int32) float64 {
	rounded := round(decimal.NewFromFloat(a), places)

	return rounded.InexactFloat64()
}

func round(amount decimal.Decimal, places int32) decimal.Decimal {
	rounded := amount.Round(places)

	return rounded
}
