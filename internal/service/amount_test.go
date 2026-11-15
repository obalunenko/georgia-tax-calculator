package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

func TestParseAmount(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    float64
		wantErr string
	}{
		{name: "integer", raw: "1500", want: 1500},
		{name: "decimal", raw: "1500.50", want: 1500.5},
		{name: "spaces are trimmed", raw: " 12 ", want: 12},
		{name: "empty", raw: "  ", wantErr: "invalid amount: value is required"},
		{name: "trailing text", raw: "12abc", wantErr: `invalid amount: "12abc" is not a number`},
		{name: "thousands separator", raw: "1,500", wantErr: `invalid amount: "1,500" is not a number`},
		{name: "NaN", raw: "NaN", wantErr: `invalid amount: "NaN" is not a number`},
		{name: "Inf", raw: "Inf", wantErr: `invalid amount: "Inf" is not a number`},
		{name: "zero", raw: "0", wantErr: "invalid amount: must be greater than 0"},
		{name: "zero with decimals", raw: "0.00", wantErr: "invalid amount: must be greater than 0"},
		{name: "negative", raw: "-5", wantErr: "invalid amount: must be greater than 0"},
		{name: "beyond float64", raw: "1e309", wantErr: `invalid amount: "1e309" is too large`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAmount(tt.raw)
			if tt.wantErr != "" {
				require.ErrorIs(t, err, ErrInvalidAmount)
				assert.EqualError(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestServiceRejectsInvalidAmounts(t *testing.T) {
	s := service{c: mockConverter{}}
	date := DateRequest{Year: "2022", Month: time.December.String(), Day: testDay}

	t.Run("convert zero amount", func(t *testing.T) {
		_, err := s.Convert(t.Context(), ConvertRequest{
			DateRequest:  date,
			CurrencyFrom: currencies.EUR,
			CurrencyTo:   currencies.GEL,
			Amount:       "0",
		})
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})

	t.Run("calculate negative year income", func(t *testing.T) {
		_, err := s.Calculate(t.Context(), CalculateRequest{
			Income:     []Income{{DateRequest: date, Currency: currencies.EUR, Amount: "100", Type: nonCash}},
			TaxType:    taxes.TaxTypeSmallBusiness.String(),
			YearIncome: "-1",
		})
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})

	t.Run("calculate zero year income", func(t *testing.T) {
		_, err := s.Calculate(t.Context(), CalculateRequest{
			Income:     []Income{{DateRequest: date, Currency: currencies.EUR, Amount: "100", Type: nonCash}},
			TaxType:    taxes.TaxTypeSmallBusiness.String(),
			YearIncome: "0",
		})
		assert.NoError(t, err)
	})

	t.Run("calculate negative income amount", func(t *testing.T) {
		_, err := s.Calculate(t.Context(), CalculateRequest{
			Income:     []Income{{DateRequest: date, Currency: currencies.EUR, Amount: "-100", Type: nonCash}},
			TaxType:    taxes.TaxTypeSmallBusiness.String(),
			YearIncome: "0",
		})
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})
}

func TestParseYearIncome(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    float64
		wantErr string
	}{
		{name: "zero", raw: "0", want: 0},
		{name: "zero with decimals", raw: "0.00", want: 0},
		{name: "positive", raw: "67.99", want: 67.99},
		{name: "negative", raw: "-1", wantErr: "invalid amount: must be 0 or greater"},
		{name: "empty", raw: "", wantErr: "invalid amount: value is required"},
		{name: "trailing text", raw: "12abc", wantErr: `invalid amount: "12abc" is not a number`},
		{name: "beyond float64", raw: "1e400", wantErr: `invalid amount: "1e400" is too large`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseYearIncome(tt.raw)
			if tt.wantErr != "" {
				require.ErrorIs(t, err, ErrInvalidAmount)
				assert.EqualError(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculateRejectsYearIncomeBeyondFloat64(t *testing.T) {
	s := service{c: failingConverter{t: t}}

	require.NotPanics(t, func() {
		_, err := s.Calculate(t.Context(), CalculateRequest{
			Income:     []Income{gelIncome(time.October.String(), "05", "100", nonCash)},
			TaxType:    taxes.TaxTypeSmallBusiness.String(),
			YearIncome: "1e400",
		})
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})
}
