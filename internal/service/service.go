// Package service holds business logic.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/obalunenko/georgia-tax-calculator/internal/converter"
	"github.com/obalunenko/georgia-tax-calculator/internal/models"
	"github.com/obalunenko/georgia-tax-calculator/internal/moneyutils"
	"github.com/obalunenko/georgia-tax-calculator/internal/spinner"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/dateutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

const (
	layout = "2006-01-02"
)

// CalculateRequest model.
type CalculateRequest struct {
	DateRequest
	Currency string `survey:"currency"`
	Amount   string `survey:"amount"`
	Taxtype  string `survey:"tax_type"`
}

// DateRequest model.
type DateRequest struct {
	Year  string `survey:"year"`
	Month string `survey:"month"`
	Day   string `survey:"day"`
}

// CalculateResponse model.
type CalculateResponse struct {
	Date            time.Time
	TaxRate         taxes.TaxRate
	Income          models.Money
	IncomeConverted models.Money
	Tax             models.Money
}

func (c CalculateResponse) String() string {
	var resp string

	resp += fmt.Sprintf("Date: %s\n", c.Date.Format(layout))
	resp += fmt.Sprintf("Tax Rate: %s\n", c.TaxRate.String())
	resp += fmt.Sprintf("Income: %s\n", c.Income.String())
	resp += fmt.Sprintf("Converted: %s\n", c.IncomeConverted.String())
	resp += fmt.Sprintf("Taxes: %s\n", c.Tax.String())

	return resp
}

// ConvertRequest model.
type ConvertRequest struct {
	DateRequest
	CurrencyFrom string `survey:"currency_from"`
	CurrencyTo   string `survey:"currency_to"`
	Amount       string `survey:"amount"`
}

// ConvertResponse model.
type ConvertResponse struct {
	Date      time.Time
	Amount    models.Money
	Converted models.Money
}

func (c ConvertResponse) String() string {
	var resp string

	resp += fmt.Sprintf("Date: %s\n", c.Date.Format(layout))
	resp += fmt.Sprintf("Amount: %s\n", c.Amount.String())
	resp += fmt.Sprintf("Converted: %s\n", c.Converted.String())

	return resp
}

// Service for calculations of taxes and currency conversions.
type Service interface {
	Convert(ctx context.Context, p ConvertRequest) (*ConvertResponse, error)
	Calculate(ctx context.Context, p CalculateRequest) (*CalculateResponse, error)
}

type service struct {
	c converter.Converter
}

// New is a Service constructor.
func New() Service {
	client := nbggovge.New()

	c := converter.NewConverter(client)

	return service{
		c: c,
	}
}

func (s service) Convert(ctx context.Context, p ConvertRequest) (*ConvertResponse, error) {
	stop := spinner.Start()
	defer stop()

	year, err := dateutils.ParseYear(p.Year)
	if err != nil {
		return nil, err
	}

	month, err := dateutils.ParseMonth(p.Month)
	if err != nil {
		return nil, err
	}

	day, err := dateutils.ParseDay(p.Day)
	if err != nil {
		return nil, err
	}

	date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)

	mv, err := moneyutils.Parse(p.Amount)
	if err != nil {
		return nil, err
	}

	amount := models.NewMoney(mv, p.CurrencyFrom)

	converted, err := s.convert(ctx, convertParams{
		date:  date,
		m:     amount,
		tocur: p.CurrencyTo,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert: %w", err)
	}

	return &ConvertResponse{
		Date:      date,
		Amount:    amount,
		Converted: converted.Money,
	}, nil
}

// Calculate calculates taxes amount.
func (s service) Calculate(ctx context.Context, p CalculateRequest) (*CalculateResponse, error) {
	convertResp, err := s.Convert(ctx, ConvertRequest{
		DateRequest: DateRequest{
			Year:  p.Year,
			Month: p.Month,
			Day:   p.Day,
		},
		CurrencyFrom: p.Currency,
		CurrencyTo:   currencies.GEL,
		Amount:       p.Amount,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert income: %w", err)
	}

	tt, err := taxes.ParseTaxType(p.Taxtype)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tax type: %w", err)
	}

	tax, err := taxes.Calc(convertResp.Converted, tt)
	if err != nil {
		return nil, fmt.Errorf("failed to Calculate taxes: %w", err)
	}

	return &CalculateResponse{
		Date:            convertResp.Date,
		TaxRate:         tax.Rate,
		Income:          convertResp.Amount,
		IncomeConverted: convertResp.Converted,
		Tax:             tax.Money,
	}, nil
}

type convertParams struct {
	date  time.Time
	m     models.Money
	tocur string
}

func (s service) convert(ctx context.Context, p convertParams) (converter.Response, error) {
	resp, err := s.c.Convert(ctx, p.m, p.tocur, p.date)
	if err != nil {
		return converter.Response{}, err
	}

	return resp, nil
}
