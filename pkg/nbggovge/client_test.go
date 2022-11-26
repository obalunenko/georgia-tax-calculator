package nbggovge_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

type doerMock struct{}

const (
	NOTEXIST = "NOT_EXIST"
)

func (d doerMock) Do(req *http.Request) (*http.Response, error) {
	q := req.URL.Query()

	cur := q.Get("currencies")
	date := q.Get("date")

	rates := nbggovge.RatesResponse{
		{
			Date: date,
			Currencies: []nbggovge.Currency{
				{
					Code:          cur,
					Quantity:      1,
					RateFormated:  "2.02",
					DiffFormated:  "0",
					Rate:          2.02,
					Name:          "MOCK",
					Diff:          0,
					Date:          date,
					ValidFromDate: date,
				},
			},
		},
	}

	body, err := json.Marshal(rates)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	switch cur {
	case NOTEXIST:
		status := http.StatusNotFound

		return &http.Response{
			Status:           http.StatusText(status),
			StatusCode:       status,
			Proto:            "",
			ProtoMajor:       0,
			ProtoMinor:       0,
			Header:           nil,
			Body:             io.NopCloser(nil),
			ContentLength:    0,
			TransferEncoding: nil,
			Close:            false,
			Uncompressed:     false,
			Trailer:          nil,
			Request:          req,
			TLS:              nil,
		}, nil

	default:
		status := http.StatusOK

		return &http.Response{
			Status:           http.StatusText(status),
			StatusCode:       status,
			Proto:            "",
			ProtoMajor:       0,
			ProtoMinor:       0,
			Header:           nil,
			Body:             io.NopCloser(bytes.NewReader(body)),
			ContentLength:    0,
			TransferEncoding: nil,
			Close:            false,
			Uncompressed:     false,
			Trailer:          nil,
			Request:          req,
			TLS:              nil,
		}, nil
	}
}

func TestClient_Rates(t *testing.T) {
	ctx := context.Background()

	type fields struct {
		httpClient nbggovge.HTTPClient
	}

	type args struct {
		ctx  context.Context
		opts []option.RatesOption
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		wantPath string
		wantErr  require.ErrorAssertionFunc
	}{
		{
			name: "integration: one currency EUR",
			fields: fields{
				httpClient: http.DefaultClient,
			},
			args: args{
				ctx: ctx,
				opts: []option.RatesOption{
					option.WithDate(time.Date(2022, 11, 25, 0, 0, 0, 0, time.UTC)),
					option.WithCurrency(currencies.EUR),
				},
			},
			wantPath: filepath.Join("testdata", "2022-11-25-eur.json"),
			wantErr:  require.NoError,
		},
		{
			name: "integration: two currencies: USD, GBP",
			fields: fields{
				httpClient: http.DefaultClient,
			},
			args: args{
				ctx: ctx,
				opts: []option.RatesOption{
					option.WithDate(time.Date(2022, 11, 25, 0, 0, 0, 0, time.UTC)),
					option.WithCurrency(currencies.USD),
					option.WithCurrency(currencies.GBP),
				},
			},
			wantPath: filepath.Join("testdata", "2022-11-25-usd-gbp.json"),
			wantErr:  require.NoError,
		},
		{
			name: "integration: no currencies - all currencies",
			fields: fields{
				httpClient: http.DefaultClient,
			},
			args: args{
				ctx: ctx,
				opts: []option.RatesOption{
					option.WithDate(time.Date(2022, 11, 25, 0, 0, 0, 0, time.UTC)),
				},
			},
			wantPath: filepath.Join("testdata", "2022-11-25-all.json"),
			wantErr:  require.NoError,
		},
		{
			name: "integration: not existed currency - empty",
			fields: fields{
				httpClient: http.DefaultClient,
			},
			args: args{
				ctx: ctx,
				opts: []option.RatesOption{
					option.WithDate(time.Date(2022, 11, 25, 0, 0, 0, 0, time.UTC)),
					option.WithCurrency("NOT_EXIST"),
				},
			},
			wantPath: filepath.Join("testdata", "2022-11-25-notexist.json"),
			wantErr:  require.NoError,
		},
		{
			name: "mocked: not existed currency - empty",
			fields: fields{
				httpClient: doerMock{},
			},
			args: args{
				ctx: ctx,
				opts: []option.RatesOption{
					option.WithDate(time.Date(2022, 11, 25, 0, 0, 0, 0, time.UTC)),
					option.WithCurrency(NOTEXIST),
				},
			},
			wantPath: filepath.Join("testdata", "empty.json"),
			wantErr:  require.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := ratesResponseFromFile(t, tt.wantPath)

			cli := nbggovge.NewWithHTTPClient(tt.fields.httpClient)

			got, err := cli.Rates(tt.args.ctx, tt.args.opts...)
			tt.wantErr(t, err)

			assert.Equal(t, want, got)
		})
	}
}
