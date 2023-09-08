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

	// The mathematical formula for the calculation is:
	// rate = (fromCurrency.Rate / fromCurrency.Quantity) / (toCurrency.Rate / toCurrency.Quantity).

	// Dividing fromCurrency's rate by its quantity.
	divFrom := moneyutils.Div(fromCurrency.Rate, float64(fromCurrency.Quantity))

	// Dividing toCurrency's rate by its quantity.
	divTo := moneyutils.Div(toCurrency.Rate, float64(toCurrency.Quantity))

	// Calculating the rate by dividing divFrom by divTo according to the formula.
	rate := moneyutils.Div(divFrom, divTo)

	convertedAmount := moneyutils.Multiply(m.Amount, rate)

	const (
		amountPlaces int32 = 2
		ratePlaces   int32 = 4
	)

	return Response{
		Money: models.Money{
			Amount:   moneyutils.Round(convertedAmount, amountPlaces),
			Currency: to,
		},
		Rate: moneyutils.Round(rate, ratePlaces),
	}, nil
}

func (c converter) getCurrencyRates(code string, rates nbggovge.Rates) (nbggovge.Currency, error) {
	currency, err := rates.CurrencyByCode(code)
	if err != nil {
		return nbggovge.Currency{}, err
	}

	return currency, nil
}
