package mock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

const (
	day1 = "2024-02-10T00:00:00.000Z"
	day2 = "2024-02-11T00:00:00.000Z"
)

func currency(code string, rate float64, date string) nbggovge.Currency {
	return nbggovge.Currency{
		Code:          code,
		Quantity:      1,
		RateFormated:  "",
		DiffFormated:  "",
		Rate:          rate,
		Name:          code,
		Diff:          0,
		Date:          date,
		ValidFromDate: date,
	}
}

func gel(date string) nbggovge.Currency {
	return nbggovge.Currency{
		Code:          currencies.GEL,
		Quantity:      1,
		RateFormated:  "1.0000",
		DiffFormated:  "0.0000",
		Rate:          1,
		Name:          "Georgian Lari",
		Diff:          0,
		Date:          date,
		ValidFromDate: date,
	}
}

func testData() []nbggovge.Rates {
	return []nbggovge.Rates{
		{
			Date: day1,
			Currencies: []nbggovge.Currency{
				currency(currencies.USD, 2.65, day1),
				currency(currencies.EUR, 2.85, day1),
				currency(currencies.GBP, 3.35, day1),
			},
		},
		{
			Date: day2,
			Currencies: []nbggovge.Currency{
				currency(currencies.USD, 2.66, day2),
				currency(currencies.EUR, 2.86, day2),
			},
		},
	}
}

func TestNewClient_Rates(t *testing.T) {
	date1 := time.Date(2024, time.February, 10, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2024, time.February, 11, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		opts []option.RatesOption
		want nbggovge.Rates
	}{
		{
			name: "requested currencies for the requested date",
			opts: []option.RatesOption{
				option.WithDate(date2),
				option.WithCurrency(currencies.USD),
			},
			want: nbggovge.Rates{
				Date:       day2,
				Currencies: []nbggovge.Currency{currency(currencies.USD, 2.66, day2)},
			},
		},
		{
			name: "several currencies are sorted and GEL is added when requested",
			opts: []option.RatesOption{
				option.WithDate(date1),
				option.WithCurrency(currencies.USD),
				option.WithCurrency(currencies.GEL),
				option.WithCurrency(currencies.EUR),
			},
			want: nbggovge.Rates{
				Date: day1,
				Currencies: []nbggovge.Currency{
					currency(currencies.EUR, 2.85, day1),
					gel(day1),
					currency(currencies.USD, 2.65, day1),
				},
			},
		},
		{
			name: "no currencies requested - all currencies and GEL",
			opts: []option.RatesOption{
				option.WithDate(date1),
			},
			want: nbggovge.Rates{
				Date: day1,
				Currencies: []nbggovge.Currency{
					currency(currencies.EUR, 2.85, day1),
					currency(currencies.GBP, 3.35, day1),
					gel(day1),
					currency(currencies.USD, 2.65, day1),
				},
			},
		},
		{
			name: "currency missing for the date - empty set",
			opts: []option.RatesOption{
				option.WithDate(date2),
				option.WithCurrency(currencies.GBP),
			},
			want: nbggovge.Rates{
				Date:       "",
				Currencies: []nbggovge.Currency{},
			},
		},
		{
			name: "no rates for the date - empty set",
			opts: []option.RatesOption{
				option.WithDate(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)),
				option.WithCurrency(currencies.USD),
			},
			want: nbggovge.Rates{
				Date:       "",
				Currencies: []nbggovge.Currency{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := testData()

			got, err := NewClient(data).Rates(t.Context(), tt.opts...)
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, testData(), data, "data passed to NewClient must not change")
		})
	}
}
