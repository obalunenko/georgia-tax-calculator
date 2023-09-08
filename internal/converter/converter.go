// Package converter provides functionality for converting money from currency to currency.
package converter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/obalunenko/georgia-tax-calculator/internal/models"
	"github.com/obalunenko/georgia-tax-calculator/pkg/moneyutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

// ErrCurrencyNotSet returned when currency code for conversion not set.
var ErrCurrencyNotSet = errors.New("currency not set")

// Converter is a converter of money from one currency to another.
type Converter interface {
	Convert(ctx context.Context, m models.Money, toCurrency string, date time.Time) (Response, error)
}

type converter struct {
	client nbggovge.Client
}

// NewConverter constructor.
func NewConverter(client nbggovge.Client) Converter {
	return &converter{client: client}
}

// Response of conversion.
type Response struct {
	models.Money
	Rate float64
}

// Convert converts amount from currency to with rates according to passed date.
func (c converter) Convert(ctx context.Context, m models.Money, to string, date time.Time) (Response, error) {
	if m.Currency == "" {
		return Response{}, fmt.Errorf("from: %w", ErrCurrencyNotSet)
	}

	if to == "" {
		return Response{}, fmt.Errorf("to: %w", ErrCurrencyNotSet)
	}

	rates, err := c.client.Rates(ctx, option.WithDate(date), option.WithCurrency(m.Currency), option.WithCurrency(to))
	if err != nil {
		return Response{}, err
	}

	fromCurrency, err := c.getCurrencyRates(m.Currency, rates)
	if err != nil {
		return Response{}, err
	}

	toCurrency, err := c.getCurrencyRates(to, rates)
	if err != nil {
		return Response{}, err
	}

	fromingel := convert(m.Amount, fromCurrency.Rate, float64(fromCurrency.Quantity))

	tosum := convert(fromingel, 1/toCurrency.Rate, 1/float64(toCurrency.Quantity))

	rate := moneyutils.Div(tosum, m.Amount)

	const (
		amountPlaces int32 = 2
		ratePlaces   int32 = 4
	)

	return Response{
		Money: models.Money{
			Amount:   round(tosum, amountPlaces),
			Currency: to,
		},
		Rate: round(rate, ratePlaces),
	}, nil
}

func (c converter) getCurrencyRates(code string, rates nbggovge.Rates) (nbggovge.Currency, error) {
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
			Date:          rates.Date,
			ValidFromDate: rates.Date,
		}
	} else {
		currency, err = rates.CurrencyByCode(code)
		if err != nil {
			return nbggovge.Currency{}, err
		}
	}

	return currency, nil
}

func convert(amount, rate, quantity float64) float64 {
	r := moneyutils.Div(rate, quantity)

	return moneyutils.Multiply(amount, r)
}

func round(amount float64, places int32) float64 {
	return moneyutils.Round(amount, places)
}
