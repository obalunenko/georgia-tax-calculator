package nbggovge

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

	"github.com/obalunenko/getenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

type doerMock struct{}

const (
	NOTEXIST = "NOT_EXIST"
)

func (d doerMock) Do(req *http.Request) (*http.Response, error) {
	q := req.URL.Query()

	cur := q.Get(currenciesParam)
	date := q.Get("date")

	rates := ratesResponse{
		{
			Date: date,
			Currencies: []Currency{
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
	if !getenv.EnvOrDefault("INTEGRATION_TESTS", false) {
		t.Skip("skipping integration tests")
	}

	ctx := context.Background()

	type fields struct {
		httpClient HTTPClient
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
					option.WithDate(time.Date(2024, time.February, 10, 0, 0, 0, 0, time.UTC)),
					option.WithCurrency(currencies.EUR),
				},
			},
			wantPath: filepath.Join("testdata", "2024-02-10-eur.json"),
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
					option.WithDate(time.Date(2024, time.February, 10, 0, 0, 0, 0, time.UTC)),
					option.WithCurrency(currencies.USD),
					option.WithCurrency(currencies.GBP),
				},
			},
			wantPath: filepath.Join("testdata", "2024-02-10-usd-gbp.json"),
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
					option.WithDate(time.Date(2024, time.February, 10, 0, 0, 0, 0, time.UTC)),
				},
			},
			wantPath: filepath.Join("testdata", "2024-02-10-all.json"),
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
					option.WithDate(time.Date(2022, time.November, 25, 0, 0, 0, 0, time.UTC)),
					option.WithCurrency(NOTEXIST),
				},
			},
			wantPath: filepath.Join("testdata", "2024-02-10-notexist.json"),
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
					option.WithDate(time.Date(2022, time.November, 25, 0, 0, 0, 0, time.UTC)),
					option.WithCurrency(NOTEXIST),
				},
			},
			wantPath: filepath.Join("testdata", "empty.json"),
			wantErr:  require.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := ratesFromFile(t, tt.wantPath)

			cli := NewWithHTTPClient(tt.fields.httpClient)

			got, err := cli.Rates(tt.args.ctx, tt.args.opts...)
			tt.wantErr(t, err)

			assert.Equal(t, want, got)
		})
	}
}
