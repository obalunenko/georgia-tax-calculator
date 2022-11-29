package converter

import (
	"context"
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
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

type mockRatesClient struct {
	data nbggovge.RatesResponse
}

func newMockRatesClient(t testing.TB) mockRatesClient {
	bytes, err := os.ReadFile(filepath.Join("testdata", "2022-11-25-all.json"))
	require.NoError(t, err)

	resp, err := nbggovge.UnmarshalRatesResponse(bytes)
	require.NoError(t, err)

	return mockRatesClient{
		data: resp,
	}
}

func (m mockRatesClient) Rates(_ context.Context, _ ...option.RatesOption) (nbggovge.RatesResponse, error) {
	return m.data, nil
}

func TestConverter_ToGel(t *testing.T) {
	ctx := context.Background()

	type fields struct {
		client nbggovge.Client
	}

	type args struct {
		ctx  context.Context
		m    models.Money
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
			name: "EUR, no error",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   2678.27,
					Currency: currencies.EUR,
				},
				date: time.Now(),
			},
			want: Response{
				Money: models.Money{
					Amount:   7557.54,
					Currency: currencies.GEL,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "EUR, no error",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx: ctx,
				m: models.Money{
					Amount:   2678.27,
					Currency: currencies.GBP,
				},
				date: time.Now(),
			},
			want: Response{models.Money{
				Amount:   8802.4,
				Currency: currencies.GEL,
			},
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Converter{
				client: tt.fields.client,
			}

			got, err := c.ConvertToGel(tt.args.ctx, tt.args.m, tt.args.date)
			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConverter_Convert(t *testing.T) {
	ctx := context.Background()

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
					Amount:   2678.27,
					Currency: currencies.EUR,
				},
				to:   currencies.GEL,
				date: time.Now(),
			},
			want: Response{models.Money{
				Amount:   7557.54,
				Currency: currencies.GEL,
			},
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
					Amount:   2678.27,
					Currency: currencies.EUR,
				},
				to:   currencies.EUR,
				date: time.Now(),
			},
			want: Response{models.Money{
				Amount:   2678.27,
				Currency: currencies.EUR,
			},
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
					Amount:   2678.27,
					Currency: currencies.EUR,
				},
				to:   currencies.GBP,
				date: time.Now(),
			},
			want: Response{models.Money{
				Amount:   2299.50,
				Currency: currencies.GBP,
			},
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
					Amount:   2678.27,
					Currency: "",
				},
				to:   currencies.GBP,
				date: time.Now(),
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
					Amount:   2678.27,
					Currency: currencies.EUR,
				},
				to:   "",
				date: time.Now(),
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
					Amount:   2678.27,
					Currency: currencies.GEL,
				},
				to:   currencies.GEL,
				date: time.Now(),
			},
			want: Response{models.Money{
				Amount:   2678.27,
				Currency: currencies.GEL,
			},
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Converter{
				client: tt.fields.client,
			}

			got, err := c.Convert(tt.args.ctx, tt.args.m, tt.args.to, tt.args.date)
			if !tt.wantErr(t, err, fmt.Sprintf("Convert(%v, %v, %v, %v)", tt.args.ctx, tt.args.m, tt.args.to, tt.args.date)) {
				return
			}

			assert.Equalf(t, tt.want, got, "Convert(%v, %v, %v, %v)", tt.args.ctx, tt.args.m, tt.args.to, tt.args.date)
		})
	}
}
