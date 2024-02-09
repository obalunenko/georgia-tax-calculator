package nbggovge

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

func TestRatesResponse_CurrencyByCode(t *testing.T) {
	type args struct {
		code string
	}

	tests := []struct {
		name    string
		r       Rates
		args    args
		want    Currency
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "usd - valid",
			r:    ratesFromFile(t, filepath.Join("testdata", "2024-02-10-all.json")),
			args: args{
				code: currencies.USD,
			},
			want: Currency{
				Code:          currencies.USD,
				Quantity:      1,
				RateFormated:  "2.6543",
				DiffFormated:  "0.0030",
				Rate:          2.6543,
				Name:          "US Dollar",
				Diff:          -0.003,
				Date:          "2024-02-09T17:45:14.052Z",
				ValidFromDate: "2024-02-10T00:00:00.000Z",
			},
			wantErr: assert.NoError,
		},
		{
			name: "rub - valid",
			r:    ratesFromFile(t, filepath.Join("testdata", "2024-02-10-all.json")),
			args: args{
				code: currencies.RUB,
			},
			want: Currency{
				Code:          currencies.RUB,
				Quantity:      100,
				RateFormated:  "2.9213",
				DiffFormated:  "0.0117",
				Rate:          2.9213,
				Name:          "Russian Ruble",
				Diff:          0.0117,
				Date:          "2024-02-09T17:45:14.052Z",
				ValidFromDate: "2024-02-10T00:00:00.000Z",
			},
			wantErr: assert.NoError,
		},
		{
			name: "not_exist - error",
			r:    ratesFromFile(t, filepath.Join("testdata", "2024-02-10-all.json")),
			args: args{
				code: "NOT_EXIST",
			},
			want:    Currency{},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.CurrencyByCode(tt.args.code)
			if !tt.wantErr(t, err, fmt.Sprintf("CurrencyByCode(%v)", tt.args.code)) {
				return
			}

			assert.Equalf(t, tt.want, got, "CurrencyByCode(%v)", tt.args.code)
		})
	}
}
