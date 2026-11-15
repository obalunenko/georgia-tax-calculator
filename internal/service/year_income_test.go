package service

import (
	"context"
	"errors"
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
	"The tax above is calculated at 1%.\n" +
	"For this month's declaration, recalculate with Small Business (over limit): field (16) is 3.00."

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
		return Income{Year: year, Month: time.June.String(), Day: testDay, Currency: currencies.GEL, Amount: "1"}
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

	_, err := s.Calculate(t.Context(), CalculateRequest{
		Income: []Income{
			{Year: "2024", Month: time.June.String(), Day: testDay, Currency: currencies.EUR, Amount: "100"},
			{Year: "2025", Month: time.June.String(), Day: testDay, Currency: currencies.EUR, Amount: "100"},
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
				Year: "2025", Month: time.June.String(), Day: testDay,
				Currency: currencies.GEL,
				Amount:   a,
				Type:     nonCash,
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

			got, err := s.Calculate(t.Context(), CalculateRequest{
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

func TestCheckSingleMonth(t *testing.T) {
	inc := func(month string) Income {
		return Income{Year: "2026", Month: month, Day: "05"}
	}

	tests := []struct {
		name      string
		incomes   []Income
		wantErr   string
		wantMonth bool
	}{
		{name: "no incomes"},
		{name: "one month", incomes: []Income{inc(time.October.String()), inc(time.October.String())}},
		{
			name:      "two months",
			incomes:   []Income{inc(time.September.String()), inc(time.October.String())},
			wantErr:   "September and October: incomes span more than one calendar month",
			wantMonth: true,
		},
		{name: "invalid month", incomes: []Income{inc("Smarch")}, wantErr: "income 1: Smarch: incorrect month"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkSingleMonth(tt.incomes)
			if tt.wantErr == "" {
				assert.NoError(t, err)

				return
			}

			require.Error(t, err)
			assert.Equal(t, tt.wantMonth, errors.Is(err, ErrIncomesSpanMonths))
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestCalculateMonthRule(t *testing.T) {
	incomes := []Income{
		gelIncome(time.September.String(), "05", "100", nonCash),
		gelIncome(time.October.String(), "05", "100", nonCash),
	}

	t.Run("small business rejects two months before conversion", func(t *testing.T) {
		s := service{c: failingConverter{t: t}}

		_, err := s.Calculate(t.Context(), CalculateRequest{
			Income:     incomes,
			TaxType:    taxes.TaxTypeSmallBusinessOverLimit.String(),
			YearIncome: "0",
		})

		require.ErrorIs(t, err, ErrIncomesSpanMonths)
		assert.EqualError(t, err, "failed to check incomes: September and October: incomes span more than one calendar month")
	})

	t.Run("employment allows two months of one year", func(t *testing.T) {
		s := service{c: mockConverter{}}

		got, err := s.Calculate(t.Context(), CalculateRequest{
			Income:     incomes,
			TaxType:    taxes.TaxTypeEmployment.String(),
			YearIncome: "0",
		})

		require.NoError(t, err)
		assert.Equal(t, models.NewMoney(40, currencies.GEL), got.Tax)

		assert.Nil(t, got.Declaration)

		for _, inc := range got.Incomes {
			assert.False(t, inc.Type.Valid())
		}

		out := got.String()
		assert.NotContains(t, out, "Type:")
		assert.NotContains(t, out, DeclarationHeader)
	})
}
