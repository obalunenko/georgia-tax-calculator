package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/moneyutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

var nonCash = taxes.IncomeTypeNonCash.String()

func gelIncome(month, day, amount, incomeType string) Income {
	return Income{
		Year: "2026", Month: month, Day: day,
		Currency: currencies.GEL,
		Amount:   amount,
		Type:     incomeType,
	}
}

func TestCheckIncomeTypes(t *testing.T) {
	tests := []struct {
		name    string
		taxType taxes.TaxType
		types   []string
		want    []taxes.IncomeType
		wantErr string
	}{
		{
			name:    "small business types parsed",
			taxType: taxes.TaxTypeSmallBusiness,
			types:   []string{"Cash", "non-cash"},
			want:    []taxes.IncomeType{taxes.IncomeTypeCash, taxes.IncomeTypeNonCash},
		},
		{
			name:    "missing type",
			taxType: taxes.TaxTypeSmallBusinessOverLimit,
			types:   []string{""},
			wantErr: `income 1: "": invalid income type`,
		},
		{
			name:    "unknown type at position 2",
			taxType: taxes.TaxTypeSmallBusiness,
			types:   []string{"Cash", "Card"},
			wantErr: `income 2: "Card": invalid income type`,
		},
		{
			name:    "employment ignores types",
			taxType: taxes.TaxTypeEmployment,
			types:   []string{"Card", ""},
			want:    []taxes.IncomeType{0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			incomes := make([]Income, 0, len(tt.types))
			for _, it := range tt.types {
				incomes = append(incomes, gelIncome(time.October.String(), "05", "100", it))
			}

			got, err := checkIncomeTypes(tt.taxType, incomes)
			if tt.wantErr != "" {
				require.ErrorIs(t, err, taxes.ErrInvalidIncomeType)
				assert.EqualError(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculateRejectsInvalidIncomeTypeBeforeConversion(t *testing.T) {
	s := service{c: failingConverter{t: t}}

	_, err := s.Calculate(t.Context(), CalculateRequest{
		Income:     []Income{gelIncome(time.October.String(), "05", "100", "")},
		TaxType:    taxes.TaxTypeSmallBusiness.String(),
		YearIncome: "0",
	})

	require.ErrorIs(t, err, taxes.ErrInvalidIncomeType)
	assert.EqualError(t, err, `failed to check incomes: income 1: "": invalid income type`)
}

func TestCalculateSetsIncomeTypes(t *testing.T) {
	s := service{c: mockConverter{}}

	got, err := s.Calculate(t.Context(), CalculateRequest{
		Income: []Income{
			gelIncome(time.October.String(), "05", "100", "Cash"),
			gelIncome(time.October.String(), "20", "50", "Non-cash"),
		},
		TaxType:    taxes.TaxTypeSmallBusiness.String(),
		YearIncome: "0",
	})
	require.NoError(t, err)
	require.Len(t, got.Incomes, 2)

	assert.Equal(t, taxes.IncomeTypeCash, got.Incomes[0].Type)
	assert.Equal(t, taxes.IncomeTypeNonCash, got.Incomes[1].Type)
	assert.Contains(t, got.String(), "\t\tType: Non-cash (20)\n")
}

func TestCalculateEmploymentIgnoresIncomeTypes(t *testing.T) {
	s := service{c: mockConverter{}}

	got, err := s.Calculate(t.Context(), CalculateRequest{
		Income:     []Income{gelIncome(time.October.String(), "05", "100", "Card")},
		TaxType:    taxes.TaxTypeEmployment.String(),
		YearIncome: "0",
	})
	require.NoError(t, err)
	require.Len(t, got.Incomes, 1)

	assert.False(t, got.Incomes[0].Type.Valid())
	assert.NotContains(t, got.String(), "Type:")
}

func formattedFields(d *Declaration) map[int]string {
	got := make(map[int]string, len(d.Fields))

	for _, f := range d.Fields {
		got[f.Number] = f.Formatted()
	}

	return got
}

func TestCalculateDeclaration(t *testing.T) {
	tests := []struct {
		name       string
		taxType    taxes.TaxType
		yearIncome string
		incomes    []Income
		want       map[int]string
		wantTax    float64
	}{
		{
			name:       "screenshot example (US1.1)",
			taxType:    taxes.TaxTypeSmallBusiness,
			yearIncome: "31792",
			incomes: []Income{
				gelIncome(time.October.String(), "05", "3950", "Cash"),
				gelIncome(time.October.String(), "20", "200", "Non-cash"),
			},
			want: map[int]string{
				15: "35942.00", 16: "1.00", 17: "4150.00",
				18: "3950.00", 19: "0.00", 20: "200.00", 21: "0.00", 22: "0.00",
			},
			wantTax: 41.5,
		},
		{
			name:       "several currencies in one type (US1.2)",
			taxType:    taxes.TaxTypeSmallBusiness,
			yearIncome: "0",
			incomes: []Income{
				{Year: "2026", Month: time.October.String(), Day: "05", Currency: currencies.USD, Amount: "1000", Type: nonCash},
				{Year: "2026", Month: time.October.String(), Day: "06", Currency: currencies.EUR, Amount: "500.55", Type: nonCash},
			},
			want: map[int]string{
				15: "1500.55", 16: "1.00", 17: "1500.55",
				18: "0.00", 19: "0.00", 20: "1500.55", 21: "0.00", 22: "0.00",
			},
			wantTax: 15.01,
		},
		{
			name:       "over limit rate (US1.3)",
			taxType:    taxes.TaxTypeSmallBusinessOverLimit,
			yearIncome: "600000",
			incomes:    []Income{gelIncome(time.October.String(), "05", "100", "Unpaid supply")},
			want: map[int]string{
				15: "600100.00", 16: "3.00", 17: "100.00",
				18: "0.00", 19: "0.00", 20: "0.00", 21: "0.00", 22: "100.00",
			},
			wantTax: 3,
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
			require.NotNil(t, got.Declaration)

			numbers := make([]int, 0, len(got.Declaration.Fields))
			for _, f := range got.Declaration.Fields {
				numbers = append(numbers, f.Number)
			}

			assert.Equal(t, []int{15, 16, 17, 18, 19, 20, 21, 22}, numbers, "all fields, in form order")
			assert.Equal(t, tt.want, formattedFields(got.Declaration))
			assert.InDelta(t, tt.wantTax, got.Tax.Amount, 0)

			fields := got.Declaration.Fields

			var split float64
			for _, f := range fields[3:] {
				split = moneyutils.Add(split, f.Value)
			}

			assert.InDelta(t, fields[2].Value, split, 0, "field 17 is the sum of fields 18-22")
			assert.InDelta(t, got.TotalIncomeConverted.Amount, fields[2].Value, 0, "field 17 is the total converted income")
			assert.InDelta(t, got.YearIncome.Amount, fields[0].Value, 0, "field 15 is the reported year income")
		})
	}
}

func TestCalculateEmploymentHasNoDeclaration(t *testing.T) {
	s := service{c: mockConverter{}}

	got, err := s.Calculate(t.Context(), CalculateRequest{
		Income:     []Income{gelIncome(time.October.String(), "05", "100", "")},
		TaxType:    taxes.TaxTypeEmployment.String(),
		YearIncome: "0",
	})
	require.NoError(t, err)

	assert.Nil(t, got.Declaration)
	assert.NotContains(t, got.String(), DeclarationHeader)
}

func TestCalculateResponseStringWithDeclaration(t *testing.T) {
	s := service{c: mockConverter{}}

	got, err := s.Calculate(t.Context(), CalculateRequest{
		Income: []Income{
			gelIncome(time.October.String(), "05", "3950", "Cash"),
			gelIncome(time.October.String(), "20", "200", "Non-cash"),
		},
		TaxType:    taxes.TaxTypeSmallBusiness.String(),
		YearIncome: "31792",
	})
	require.NoError(t, err)

	// The CLI example in contracts/calculation-output.md.
	want := "Tax Rate: Small Business 1 %\n" +
		"Year Income: 35942 GEL\n" +
		"Incomes:\n" +
		"\t- 1:\n" +
		"\t\tDate: 2026-10-05\n" +
		"\t\tAmount: 3950 GEL\n" +
		"\t\tConverted: 3950 GEL\n" +
		"\t\tRate: 1\n" +
		"\t\tType: Cash (18)\n" +
		"\t- 2:\n" +
		"\t\tDate: 2026-10-20\n" +
		"\t\tAmount: 200 GEL\n" +
		"\t\tConverted: 200 GEL\n" +
		"\t\tRate: 1\n" +
		"\t\tType: Non-cash (20)\n" +
		"Total Income Converted: 4150 GEL\n" +
		"Taxes: 41.5 GEL\n" +
		"\n" +
		"Declaration fields (Small Business monthly declaration):\n" +
		"(15) Income since the start of the year: 35942.00\n" +
		"(16) Tax rate, %: 1.00\n" +
		"(17) Taxable income for the month: 4150.00\n" +
		"(18) Cash: 3950.00\n" +
		"(19) POS terminal: 0.00\n" +
		"(20) Non-cash: 200.00\n" +
		"(21) Other: 0.00\n" +
		"(22) Unpaid supply: 0.00"

	assert.Equal(t, want, got.String())
}

func TestCalculateRoundsYearIncomeToTetri(t *testing.T) {
	tests := []struct {
		name        string
		yearIncome  string
		income      string
		wantYear    float64
		wantField   string
		wantWarning bool
	}{
		{name: "field 15 equals the Year Income line", yearIncome: "31792.005", income: "4150", wantYear: 35942.01, wantField: "35942.01"},
		{name: "limit compared after rounding", yearIncome: "499999.004", income: "1", wantYear: 500000, wantField: "500000.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service{c: mockConverter{}}

			got, err := s.Calculate(t.Context(), CalculateRequest{
				Income:     []Income{gelIncome(time.October.String(), "05", tt.income, "Cash")},
				TaxType:    taxes.TaxTypeSmallBusiness.String(),
				YearIncome: tt.yearIncome,
			})
			require.NoError(t, err)

			assert.InDelta(t, tt.wantYear, got.YearIncome.Amount, 0)
			assert.Equal(t, tt.wantField, got.Declaration.Fields[0].Formatted())
			assert.Equal(t, tt.wantWarning, got.LimitWarning != nil)
		})
	}
}

func TestDeclarationFieldString(t *testing.T) {
	f := DeclarationField{Number: 17, Label: "Taxable income for the month", Value: 4150}

	assert.Equal(t, "(17) Taxable income for the month: ", f.Prefix())
	assert.Equal(t, "4150.00", f.Formatted())
	assert.Equal(t, f.Prefix()+f.Formatted(), f.String())
}
