package converter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/internal/models"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/mock"
)

func newMockRatesClient(t testing.TB) nbggovge.Client {
	var data []nbggovge.Rates

	for _, name := range []string{"2022-11-25-all.json", "2023-09-09-all.json"} {
		b, err := os.ReadFile(filepath.Join("testdata", name))
		require.NoError(t, err)

		var resp []nbggovge.Rates

		require.NoError(t, json.Unmarshal(b, &resp))

		data = append(data, resp...)
	}

	return mock.NewClient(data)
}

func errorIs(target error) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, msgAndArgs ...any) bool {
		return assert.ErrorIs(t, err, target, msgAndArgs...)
	}
}

func TestConverter_Convert(t *testing.T) {
	ctx := t.Context()

	rateDate := time.Date(2023, time.September, 9, 0, 0, 0, 0, time.UTC)

	type fields struct {
		client nbggovge.Client
	}

	type args struct {
		ctx  context.Context
		m    models.Money
		to   string
		date time.Time
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Response
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "EUR - GEL",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   2_678.27,
					Currency: currencies.EUR,
				},
				to:   currencies.GEL,
				date: rateDate,
			},
			want: Response{
				Money: models.Money{
					Amount:   7521.39,
					Currency: currencies.GEL,
				},
				Rate: 2.8083,
			},
			wantErr: assert.NoError,
		},
		{
			name: "EUR - EUR",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   2_678.27,
					Currency: currencies.EUR,
				},
				to:   currencies.EUR,
				date: rateDate,
			},
			want: Response{
				Money: models.Money{
					Amount:   2_678.27,
					Currency: currencies.EUR,
				},
				Rate: 1,
			},
			wantErr: assert.NoError,
		},
		{
			name: "EUR - GBP",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   2_678.27,
					Currency: currencies.EUR,
				},
				to:   currencies.GBP,
				date: rateDate,
			},
			want: Response{
				Money: models.Money{
					Amount:   2_297.45,
					Currency: currencies.GBP,
				},
				Rate: 0.8578,
			},
			wantErr: assert.NoError,
		},
		{
			name: "no from - error",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   2_678.27,
					Currency: "",
				},
				to:   currencies.GBP,
				date: rateDate,
			},
			want:    Response{},
			wantErr: assert.Error,
		},
		{
			name: "no to - error",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   2_678.27,
					Currency: currencies.EUR,
				},
				to:   "",
				date: rateDate,
			},
			want:    Response{},
			wantErr: assert.Error,
		},
		{
			name: "GEL - GEL",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   2_678.27,
					Currency: currencies.GEL,
				},
				to:   currencies.GEL,
				date: rateDate,
			},
			want: Response{
				Money: models.Money{
					Amount:   2_678.27,
					Currency: currencies.GEL,
				},
				Rate: 1,
			},
			wantErr: assert.NoError,
		},
		{
			name: "PLN - GEL",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   2_678.27,
					Currency: currencies.PLN,
				},
				to:   currencies.GEL,
				date: rateDate,
			},
			want: Response{
				Money: models.Money{
					Amount:   1_625.9,
					Currency: currencies.GEL,
				},
				Rate: 0.6071,
			},
			wantErr: assert.NoError,
		},
		{
			name: "BYN - GEL",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   2_678.27,
					Currency: currencies.BYN,
				},
				to:   currencies.GEL,
				date: rateDate,
			},
			want: Response{
				Money: models.Money{
					Amount:   2_792.1,
					Currency: currencies.GEL,
				},
				Rate: 1.0425,
			},
			wantErr: assert.NoError,
		},
		{
			name: "BYN - PLN",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   2_678.27,
					Currency: currencies.BYN,
				},
				to:   currencies.PLN,
				date: rateDate,
			},
			want: Response{
				Money: models.Money{
					Amount:   4_599.3,
					Currency: currencies.PLN,
				},
				Rate: 1.7173,
			},
			wantErr: assert.NoError,
		},
		{
			name: "PLN - BYN",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   2_678.27,
					Currency: currencies.PLN,
				},
				to:   currencies.BYN,
				date: rateDate,
			},
			want: Response{
				Money: models.Money{
					Amount:   1_559.61,
					Currency: currencies.BYN,
				},
				Rate: 0.5823,
			},
			wantErr: assert.NoError,
		},
		{
			name: "RUB - EUR",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   50_000,
					Currency: currencies.RUB,
				},
				to:   currencies.EUR,
				date: rateDate,
			},
			want: Response{
				Money: models.Money{
					Amount:   477.74,
					Currency: currencies.EUR,
				},
				Rate: 0.0096,
			},
			wantErr: assert.NoError,
		},
		{
			name: "EUR - RUB",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   550,
					Currency: currencies.EUR,
				},
				to:   currencies.RUB,
				date: rateDate,
			},
			want: Response{
				Money: models.Money{
					Amount:   57_562.14,
					Currency: currencies.RUB,
				},
				Rate: 104.6584,
			},
			wantErr: assert.NoError,
		},
		{
			name: "RUB - GEL",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   50_000,
					Currency: currencies.RUB,
				},
				to:   currencies.GEL,
				date: rateDate,
			},
			want: Response{
				Money: models.Money{
					Amount:   1_341.65,
					Currency: currencies.GEL,
				},
				Rate: 0.0268,
			},
			wantErr: assert.NoError,
		},
		{
			name: "RUB - TMT",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   50_000,
					Currency: currencies.RUB,
				},
				to:   currencies.TMT,
				date: rateDate,
			},
			want: Response{
				Money: models.Money{
					Amount:   1_788.8,
					Currency: currencies.TMT,
				},
				Rate: 0.0358,
			},
			wantErr: assert.NoError,
		},
		{
			name: "TMT - RUB",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   1_787.95,
					Currency: currencies.TMT,
				},
				to:   currencies.RUB,
				date: rateDate,
			},
			want: Response{
				Money: models.Money{
					Amount:   49976.38,
					Currency: currencies.RUB,
				},
				Rate: 27.9518,
			},
			wantErr: assert.NoError,
		},
		{
			name: "EUR - GEL uses the rate of the requested date",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   1_000,
					Currency: currencies.EUR,
				},
				to:   currencies.GEL,
				date: time.Date(2022, time.November, 25, 0, 0, 0, 0, time.UTC),
			},
			want: Response{
				Money: models.Money{
					Amount:   2_821.8,
					Currency: currencies.GEL,
				},
				Rate: 2.8218,
			},
			wantErr: assert.NoError,
		},
		{
			name: "no rates for the date - error",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   1_000,
					Currency: currencies.EUR,
				},
				to:   currencies.GEL,
				date: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			want:    Response{},
			wantErr: errorIs(nbggovge.ErrCodeNotFound),
		},
		{
			name: "unknown currency - error",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   1_000,
					Currency: "NOT_EXIST",
				},
				to:   currencies.GEL,
				date: rateDate,
			},
			want:    Response{},
			wantErr: errorIs(nbggovge.ErrCodeNotFound),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := converter{
				client: tt.fields.client,
			}

			got, err := c.Convert(tt.args.ctx, tt.args.m, tt.args.to, tt.args.date)
			if !tt.wantErr(t, err, fmt.Sprintf(
				"Convert(%v, %v, %v, %v)",
				tt.args.ctx,
				tt.args.m,
				tt.args.to,
				tt.args.date,
			)) {
				return
			}

			assert.Equalf(t, tt.want, got,
				"Convert(%v, %v, %v, %v)",
				tt.args.ctx,
				tt.args.m,
				tt.args.to,
				tt.args.date,
			)
		})
	}
}
