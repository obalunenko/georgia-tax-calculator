package converter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

type Converter struct {
	client nbggovge.Client
}

func NewConverter(client nbggovge.Client) *Converter {
	return &Converter{client: client}
}

type Response struct {
	Amount   float64
	Currency string
}

func (r Response) String() string {
	return fmt.Sprintf("%.2f %s", r.Amount, r.Currency)
}

func (c Converter) ConvertToGel(ctx context.Context, amount float64, from string, date time.Time) (Response, error) {
	rates, err := c.client.Rates(ctx, option.WithDate(date), option.WithCurrency(from))
	if err != nil {
		return Response{}, err
	}

	currency, err := rates.CurrencyByCode(from)
	if err != nil {
		return Response{}, err
	}

	sum := convert(amount, currency.Rate)

	return Response{
		Amount:   sum,
		Currency: currencies.GEL,
	}, nil
}

func (c Converter) Convert(ctx context.Context, amount float64, from, to string, date time.Time) (Response, error) {
	return Response{}, errors.New("not implemented")
}

func convert(amount, rate float64) float64 {
	ad := decimal.NewFromFloat(amount)
	rd := decimal.NewFromFloat(rate)

	res := ad.Mul(rd)

	rounded := res.Round(2)

	return rounded.InexactFloat64()
}
