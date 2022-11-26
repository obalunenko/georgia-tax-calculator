package converter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/obalunenko/georgia-tax-calculator/internal/moneyutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

// ErrCurrencyNotSet returned when currency code for conversion not set.
var ErrCurrencyNotSet = errors.New("currency not set")

// Converter is a converter of money from one currency to another.
type Converter struct {
	client nbggovge.Client
}

// NewConverter constructor.
func NewConverter(client nbggovge.Client) *Converter {
	return &Converter{client: client}
}

// Response of conversion.
type Response struct {
	Amount   float64
	Currency string
}

func (r Response) String() string {
	return fmt.Sprintf("%.2f %s", r.Amount, r.Currency)
}

// ConvertToGel shortcut for Convert to GEL.
func (c Converter) ConvertToGel(ctx context.Context, amount float64, from string, date time.Time) (Response, error) {
	return c.Convert(ctx, amount, from, currencies.GEL, date)
}

// Convert converts amount from currency to with rates according to passed date.
func (c Converter) Convert(ctx context.Context, amount float64, from, to string, date time.Time) (Response, error) {
	if from == "" {
		return Response{}, fmt.Errorf("from: %w", ErrCurrencyNotSet)
	}

	if to == "" {
		return Response{}, fmt.Errorf("to: %w", ErrCurrencyNotSet)
	}

	rates, err := c.client.Rates(ctx, option.WithDate(date), option.WithCurrency(from), option.WithCurrency(to))
	if err != nil {
		return Response{}, err
	}

	fromCurrency, err := c.getCurrencyRates(from, rates)
	if err != nil {
		return Response{}, err
	}

	toCurrency, err := c.getCurrencyRates(to, rates)
	if err != nil {
		return Response{}, err
	}

	fromingel := convert(amount, fromCurrency.Rate)

	tosum := convert(fromingel, 1/toCurrency.Rate)

	return Response{
		Amount:   round(tosum, 2),
		Currency: to,
	}, nil
}

func (c Converter) getCurrencyRates(code string, rates nbggovge.RatesResponse) (nbggovge.Currency, error) {
	var (
		currency nbggovge.Currency
		err      error
	)

	if code == currencies.GEL {
		currency = nbggovge.Currency{
			Code:          currencies.GEL,
			Quantity:      1,
			RateFormated:  "1",
			DiffFormated:  "0",
			Rate:          1,
			Name:          "GEL",
			Diff:          0,
			Date:          rates[0].Date,
			ValidFromDate: rates[0].Date,
		}
	} else {
		currency, err = rates.CurrencyByCode(code)
		if err != nil {
			return nbggovge.Currency{}, err
		}
	}

	return currency, nil
}

func convert(amount, rate float64) float64 {
	return moneyutils.Multiply(amount, rate)
}

func round(amount float64, places int32) float64 {
	return moneyutils.Round(amount, places)
}
