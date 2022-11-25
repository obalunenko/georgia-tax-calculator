package converter

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		ctx    context.Context
		amount float64
		from   string
		date   time.Time
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
				ctx:    ctx,
				amount: 2678.27,
				from:   currencies.EUR,
				date:   time.Now(),
			},
			want: Response{
				Amount:   7557.54,
				Currency: currencies.GEL,
			},
			wantErr: assert.NoError,
		},
		{
			name: "EUR, no error",
			fields: fields{
				client: newMockRatesClient(t),
			},
			args: args{
				ctx:    ctx,
				amount: 2678.27,
				from:   currencies.GBP,
				date:   time.Now(),
			},
			want: Response{
				Amount:   8802.4,
				Currency: currencies.GEL,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Converter{
				client: tt.fields.client,
			}
			got, err := c.ConvertToGel(tt.args.ctx, tt.args.amount, tt.args.from, tt.args.date)
			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResponse_String(t *testing.T) {
	type fields struct {
		Amount   float64
		Currency string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "",
			fields: fields{
				Amount:   25.26789,
				Currency: currencies.GEL,
			},
			want: "25.27 GEL",
		},
		{
			name: "",
			fields: fields{
				Amount:   25.21289,
				Currency: currencies.GEL,
			},
			want: "25.21 GEL",
		},
		{
			name: "",
			fields: fields{
				Amount:   25.21489,
				Currency: currencies.GEL,
			},
			want: "25.21 GEL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Response{
				Amount:   tt.fields.Amount,
				Currency: tt.fields.Currency,
			}

			assert.Equalf(t, tt.want, r.String(), "String()")
		})
	}
}
