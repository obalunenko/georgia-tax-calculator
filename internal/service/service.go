// Package service holds business logic.
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/obalunenko/georgia-tax-calculator/internal/converter"
	"github.com/obalunenko/georgia-tax-calculator/internal/models"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/dateutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/moneyutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

const (
	layout = "2006-01-02"
)

// DeclarationsURL is the Revenue Service portal where tax declarations are filed.
const DeclarationsURL = "https://decl.rs.ge/decls.aspx"

// CalculateRequest model.
type CalculateRequest struct {
	Income     []Income
	TaxType    string `survey:"tax_type"`
	YearIncome string `survey:"year_income"`
}

// Income model.
type Income struct {
	DateRequest
	Currency string `survey:"currency"`
	Amount   string `survey:"amount"`
}

// DateRequest model.
type DateRequest struct {
	Year  string `survey:"year"`
	Month string `survey:"month"`
	Day   string `survey:"day"`
}

func (d DateRequest) String() string {
	return fmt.Sprintf("%s-%s-%s", d.Year, d.Month, d.Day)
}

// CalculateResponse model.
type CalculateResponse struct {
	TaxRate              taxes.TaxRate
	YearIncome           models.Money
	Incomes              []ConvertResponse
	TotalIncomeConverted models.Money
	Tax                  models.Money
	// LimitWarning is set when the Small Business turnover limit is exceeded.
	LimitWarning *LimitWarning
}

func (c CalculateResponse) String() string {
	var resp strings.Builder

	resp.WriteString(fmt.Sprintf("Tax Rate: %s\n", c.TaxRate.String()))

	resp.WriteString(fmt.Sprintf("Year Income: %s\n", c.YearIncome.String()))

	if len(c.Incomes) != 0 {
		resp.WriteString("Incomes:\n")

		for i := range c.Incomes {
			inc := c.Incomes[i]

			raw := inc.String()

			if strings.TrimSpace(raw) == "" {
				continue
			}

			resp.WriteString(fmt.Sprintf("\t- %d:\n", i+1))

			lines := strings.SplitSeq(raw, "\n")

			for line := range lines {
				if strings.TrimSpace(line) != "" {
					resp.WriteString(fmt.Sprintf("\t\t%s\n", line))
				}
			}
		}
	}

	resp.WriteString(fmt.Sprintf("Total Income Converted: %s\n", c.TotalIncomeConverted.String()))

	resp.WriteString(fmt.Sprintf("Taxes: %s", c.Tax.String()))

	if c.LimitWarning != nil {
		resp.WriteString("\n⚠️ ")
		resp.WriteString(c.LimitWarning.String())
	}

	return resp.String()
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
	Rate      models.Money
}

func (c ConvertResponse) String() string {
	var resp string

	resp += fmt.Sprintf("Date: %s\n", c.Date.Format(layout))
	resp += fmt.Sprintf("Amount: %s\n", c.Amount.String())
	resp += fmt.Sprintf("Converted: %s\n", c.Converted.String())
	resp += fmt.Sprintf("Rate: %s", c.Rate.String())

	return resp
}

// Service for calculations of taxes and currency conversions.
type Service interface {
	Converter
	TaxCalculator
}

// Converter converts currencies.
type Converter interface {
	Convert(ctx context.Context, p ConvertRequest) (*ConvertResponse, error)
}

// TaxCalculator calculates taxes.
type TaxCalculator interface {
	Calculate(ctx context.Context, p CalculateRequest) (*CalculateResponse, error)
}

type service struct {
	c converter.Converter
}

// New is a Service constructor.
func New() Service {
	client := nbggovge.NewCached()

	c := converter.NewConverter(client)

	return service{
		c: c,
	}
}

func (s service) Convert(ctx context.Context, p ConvertRequest) (*ConvertResponse, error) {
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

	date, err := newDate(year, month, day)
	if err != nil {
		return nil, err
	}

	mv, err := ParseAmount(p.Amount)
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

	rate := models.NewMoney(converted.Rate, "")

	return &ConvertResponse{
		Date:      date,
		Amount:    amount,
		Converted: converted.Money,
		Rate:      rate,
	}, nil
}

// Calculate calculates taxes amount.
func (s service) Calculate(ctx context.Context, req CalculateRequest) (*CalculateResponse, error) {
	tt, err := taxes.ParseTaxType(req.TaxType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tax type: %w", err)
	}

	tr, err := tt.Rate()
	if err != nil {
		return nil, fmt.Errorf("failed to get tax rate: %w", err)
	}

	yi, err := ParseYearIncome(req.YearIncome)
	if err != nil {
		return nil, fmt.Errorf("failed to parse year income: %w", err)
	}

	// Checked before any conversion, so an invalid request makes no NBG calls.
	if err = checkSingleYear(req.Income); err != nil {
		return nil, fmt.Errorf("failed to check incomes: %w", err)
	}

	var inc float64

	incomes := make([]ConvertResponse, 0, len(req.Income))

	for _, p := range req.Income {
		r := ConvertRequest{
			DateRequest: DateRequest{
				Year:  p.Year,
				Month: p.Month,
				Day:   p.Day,
			},
			CurrencyFrom: p.Currency,
			CurrencyTo:   currencies.GEL,
			Amount:       p.Amount,
		}

		convertResp, err := s.Convert(ctx, r)
		if err != nil {
			return nil, fmt.Errorf("failed to convert income: %w", err)
		}

		incomes = append(incomes, ConvertResponse{
			Date:      convertResp.Date,
			Amount:    convertResp.Amount,
			Converted: convertResp.Converted,
			Rate:      convertResp.Rate,
		})

		yi = moneyutils.Add(yi, convertResp.Converted.Amount)
		inc = moneyutils.Add(inc, convertResp.Converted.Amount)
	}

	// The tax is calculated once on the total, as the declaration does.
	tax, err := taxes.Calc(models.NewMoney(inc, currencies.GEL), tt)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate taxes: %w", err)
	}

	return &CalculateResponse{
		TaxRate:              tr,
		YearIncome:           models.NewMoney(yi, currencies.GEL),
		Incomes:              incomes,
		TotalIncomeConverted: models.NewMoney(inc, currencies.GEL),
		Tax:                  tax.Money,
		LimitWarning:         newLimitWarning(tt, yi),
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
