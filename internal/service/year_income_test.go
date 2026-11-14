package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/internal/converter"
	"github.com/obalunenko/georgia-tax-calculator/internal/models"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

const wantWarning = "Year income 501000 GEL exceeds the Small Business limit of 500000 GEL by 1000 GEL. " +
	"Under Article 90(2) of the Tax Code, income from the start of the month the limit was exceeded " +
	"until the end of the year is taxed at 3%: calculate it with Small Business (over limit). " +
	"The tax above is calculated at 1%."

func TestNewLimitWarning(t *testing.T) {
	tests := []struct {
		name       string
		taxType    taxes.TaxType
		yearIncome float64
		wantExcess float64
		wantNil    bool
	}{
		{name: "small business above the limit", taxType: taxes.TaxTypeSmallBusiness, yearIncome: 501_000, wantExcess: 1000},
		{name: "small business by one tetri", taxType: taxes.TaxTypeSmallBusiness, yearIncome: 500_000.01, wantExcess: 0.01},
		{name: "small business at the limit", taxType: taxes.TaxTypeSmallBusiness, yearIncome: 500_000, wantNil: true},
		{name: "small business below the limit", taxType: taxes.TaxTypeSmallBusiness, yearIncome: 1000, wantNil: true},
		{name: "employment above the limit", taxType: taxes.TaxTypeEmployment, yearIncome: 600_000, wantNil: true},
		{name: "over-limit category above the limit", taxType: taxes.TaxTypeSmallBusinessOverLimit, yearIncome: 600_000, wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newLimitWarning(tt.taxType, tt.yearIncome)
			if tt.wantNil {
				assert.Nil(t, got)

				return
			}

			require.NotNil(t, got)
			assert.Equal(t, models.NewMoney(taxes.SmallBusinessTurnoverLimit, currencies.GEL), got.Limit)
			assert.Equal(t, models.NewMoney(tt.yearIncome, currencies.GEL), got.YearIncome)
			assert.Equal(t, models.NewMoney(tt.wantExcess, currencies.GEL), got.Excess)
		})
	}
}

func TestLimitWarningString(t *testing.T) {
	w := newLimitWarning(taxes.TaxTypeSmallBusiness, 501_000)
	require.NotNil(t, w)

	assert.Equal(t, wantWarning, w.String())
}

func TestCheckSingleYear(t *testing.T) {
	income := func(year string) Income {
		return Income{DateRequest: DateRequest{Year: year, Month: "June", Day: "08"}, Currency: currencies.GEL, Amount: "1"}
	}

	tests := []struct {
		name    string
		incomes []Income
		wantErr error
		anyErr  bool
	}{
		{name: "no incomes", incomes: nil},
		{name: "one year", incomes: []Income{income("2025"), income("2025")}},
		{name: "two years", incomes: []Income{income("2025"), income("2024")}, wantErr: ErrIncomesSpanYears},
		{name: "invalid year", incomes: []Income{income("2025"), income("next")}, anyErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkSingleYear(tt.incomes)

			switch {
			case tt.wantErr != nil:
				assert.ErrorIs(t, err, tt.wantErr)
			case tt.anyErr:
				assert.Error(t, err)
			default:
				assert.NoError(t, err)
			}
		})
	}
}

// failingConverter fails the test if a conversion is requested.
type failingConverter struct {
	t *testing.T
}

func (c failingConverter) Convert(context.Context, models.Money, string, time.Time) (converter.Response, error) {
	c.t.Error("conversion must not run")

	return converter.Response{}, nil
}

func TestCalculateRejectsMixedYearsBeforeConversion(t *testing.T) {
	s := service{c: failingConverter{t: t}}

	_, err := s.Calculate(context.Background(), CalculateRequest{
		Income: []Income{
			{DateRequest: DateRequest{Year: "2024", Month: "June", Day: "08"}, Currency: currencies.EUR, Amount: "100"},
			{DateRequest: DateRequest{Year: "2025", Month: "June", Day: "08"}, Currency: currencies.EUR, Amount: "100"},
		},
		TaxType:    taxes.TaxTypeSmallBusiness.String(),
		YearIncome: "0",
	})

	assert.ErrorIs(t, err, ErrIncomesSpanYears)
}

func TestCalculateYearIncomeRules(t *testing.T) {
	gel := func(amounts ...string) []Income {
		incomes := make([]Income, 0, len(amounts))
		for _, a := range amounts {
			incomes = append(incomes, Income{
				DateRequest: DateRequest{Year: "2025", Month: "June", Day: "08"},
				Currency:    currencies.GEL,
				Amount:      a,
			})
		}

		return incomes
	}

	tests := []struct {
		name        string
		taxType     taxes.TaxType
		yearIncome  string
		incomes     []Income
		wantTax     float64
		wantWarning bool
	}{
		{
			name:        "limit exceeded",
			taxType:     taxes.TaxTypeSmallBusiness,
			yearIncome:  "499000",
			incomes:     gel("2000"),
			wantTax:     20,
			wantWarning: true,
		},
		{
			name:       "limit reached, not exceeded",
			taxType:    taxes.TaxTypeSmallBusiness,
			yearIncome: "498000",
			incomes:    gel("2000"),
			wantTax:    20,
		},
		{
			name:       "tax on the total is rounded once",
			taxType:    taxes.TaxTypeSmallBusiness,
			yearIncome: "0",
			incomes:    gel("0.50", "0.50", "0.50"),
			wantTax:    0.02,
		},
		{
			name:       "employment above the limit has no warning",
			taxType:    taxes.TaxTypeEmployment,
			yearIncome: "600000",
			incomes:    gel("100"),
			wantTax:    20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service{c: mockConverter{}}

			got, err := s.Calculate(context.Background(), CalculateRequest{
				Income:     tt.incomes,
				TaxType:    tt.taxType.String(),
				YearIncome: tt.yearIncome,
			})
			require.NoError(t, err)

			assert.Equal(t, models.NewMoney(tt.wantTax, currencies.GEL), got.Tax)
			assert.Equal(t, tt.wantWarning, got.LimitWarning != nil)

			out := got.String()
			assert.Equal(t, tt.wantWarning, strings.HasSuffix(out, "\n⚠️ "+wantWarning), out)
		})
	}
}
