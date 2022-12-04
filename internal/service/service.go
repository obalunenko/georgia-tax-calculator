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

// InputParams for tax Service calculations.
type InputParams struct {
	Year     string `survey:"year"`
	Month    string `survey:"month"`
	Day      string `survey:"day"`
	Currency string `survey:"currency"`
	Amount   string `survey:"amount"`
	Taxtype  string `survey:"tax_type"`
}

// Service for calculations of taxes.
type Service struct {
	c converter.Converter
}

// New is a Service constructor.
func New() Service {
	client := nbggovge.New()

	c := converter.NewConverter(client)

	return Service{
		c: c,
	}
}

// Calculate calculates taxes amount.
func (s Service) Calculate(ctx context.Context, p InputParams) (string, error) {
	stop := spinner.Start()
	defer stop()

	year, err := dateutils.ParseYear(p.Year)
	if err != nil {
		return "", err
	}

	month, err := dateutils.ParseMonth(p.Month)
	if err != nil {
		return "", err
	}

	day, err := dateutils.ParseDay(p.Day)
	if err != nil {
		return "", err
	}

	date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)

	mv, err := moneyutils.Parse(p.Amount)
	if err != nil {
		return "", err
	}

	income := models.NewMoney(mv, p.Currency)
	incomeOut := models.NewResultOutput("Income", income)

	converted, err := s.convert(ctx, convertParams{
		date:  date,
		m:     income,
		tocur: currencies.GEL,
	})
	if err != nil {
		return "", fmt.Errorf("failed to convert: %w", err)
	}

	convertedOut := models.NewResultOutput("Converted", converted.Money)

	tt, err := taxes.ParseTaxType(p.Taxtype)
	if err != nil {
		return "", fmt.Errorf("failed to parse tax type: %w", err)
	}

	tax, err := taxes.Calc(converted.Money, tt)
	if err != nil {
		return "", fmt.Errorf("failed to Calculate taxes: %w", err)
	}

	taxesOut := models.NewResultOutput("Taxes", tax.Money)

	const (
		layout = "2006-01-02"
	)

	var resp string
	resp += fmt.Sprintf("Date: %s\n", date.Format(layout))
	resp += fmt.Sprintf("Tax Rate: %s\n", tax.Rate.String())
	resp += fmt.Sprintf("%s\n", incomeOut.String())
	resp += fmt.Sprintf("%s\n", convertedOut.String())
	resp += fmt.Sprintf("%s\n", taxesOut.String())

	return resp, nil
}

type convertParams struct {
	date  time.Time
	m     models.Money
	tocur string
}

func (s Service) convert(ctx context.Context, p convertParams) (converter.Response, error) {
	resp, err := s.c.Convert(ctx, p.m, p.tocur, p.date)
	if err != nil {
		return converter.Response{}, err
	}

	return resp, nil
}
