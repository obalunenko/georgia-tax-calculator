package converter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/obalunenko/georgia-tax-calculator/internal/models"
	"github.com/obalunenko/georgia-tax-calculator/internal/moneyutils"
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

	fromingel := convert(m.Amount, fromCurrency.Rate)

	tosum := convert(fromingel, 1/toCurrency.Rate)

	return Response{
		Money: models.Money{
			Amount:   round(tosum, 2),
			Currency: to,
		},
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

func convert(amount, rate float64) float64 {
	return moneyutils.Multiply(amount, rate)
}

func round(amount float64, places int32) float64 {
	return moneyutils.Round(amount, places)
}
