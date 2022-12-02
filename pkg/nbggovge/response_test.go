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
			r:    ratesFromFile(t, filepath.Join("testdata", "2022-11-25-usd-gbp.json")),
			args: args{
				code: currencies.USD,
			},
			want: Currency{
				Code:          currencies.USD,
				Quantity:      1,
				RateFormated:  "2.7117",
				DiffFormated:  "0.0075",
				Rate:          2.7117,
				Name:          "US Dollar",
				Diff:          -0.0075,
				Date:          "2022-11-24T17:45:14.293Z",
				ValidFromDate: "2022-11-25T00:00:00.000Z",
			},
			wantErr: assert.NoError,
		},
		{
			name: "not_exist - error",
			r:    ratesFromFile(t, filepath.Join("testdata", "2022-11-25-usd-gbp.json")),
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
