package nbggovge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/obalunenko/getenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

type doerMock struct {
	tb testing.TB
}

const (
	NOTEXIST = "NOT_EXIST"
)

func (d doerMock) Do(req *http.Request) (*http.Response, error) {
	d.tb.Helper()

	q := req.URL.Query()

	cur := q.Get(currenciesParam)
	date := q.Get("date")

	status := http.StatusOK

	var rates ratesResponse

	switch strings.TrimSpace(cur) {
	case NOTEXIST:
		d.tb.Log("received request for not existed currency, returning empty response")

		rates = newRatesResponse("", nil)
	case "":
		d.tb.Log("received request with empty currency, bad request")

		status = http.StatusBadRequest
	default:
		rates = newRatesResponse(date, []Currency{
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
		})
	}

	body, err := json.Marshal(rates)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

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

func TestClient_Rates(t *testing.T) {
	integration := getenv.EnvOrDefault("INTEGRATION_TESTS", false)

	ctx := t.Context()

	type fields struct {
		httpClient HTTPClient
	}

	type args struct {
		ctx  context.Context
		opts []option.RatesOption
	}

	tests := []struct {
		name        string
		integration bool
		fields      fields
		args        args
		wantPath    string
		wantErr     require.ErrorAssertionFunc
	}{
		{
			name:        "integration: one currency EUR",
			integration: true,
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
			name:        "integration: two currencies: USD, GBP",
			integration: true,
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
			name:        "integration: no currencies - all currencies",
			integration: true,
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
			name:        "integration: not existed currency - empty",
			integration: true,
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
			wantPath: filepath.Join("testdata", "empty.json"),
			wantErr:  require.NoError,
		},
		{
			name: "mocked: not existed currency - empty",
			fields: fields{
				httpClient: doerMock{
					tb: t,
				},
			},
			args: args{
				ctx: ctx,
				opts: []option.RatesOption{
					option.WithDate(time.Date(2022, time.November, 25, 0, 0, 0, 0, time.UTC)),
					option.WithCurrency(NOTEXIST),
				},
			},
			wantPath: filepath.Join("testdata", "empty.json"),
			wantErr:  require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.integration && !integration {
				t.Skip("skipping integration test")
			}

			want := ratesFromFile(t, tt.wantPath)

			cli := NewWithHTTPClient(tt.fields.httpClient)

			got, err := cli.Rates(tt.args.ctx, tt.args.opts...)
			tt.wantErr(t, err)

			assert.Equal(t, want, got)
		})
	}
}

type doerFunc func(req *http.Request) (*http.Response, error)

func (f doerFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("connection reset")
}

func httpResponse(req *http.Request, status int, body io.Reader) *http.Response {
	return &http.Response{
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		StatusCode: status,
		Body:       io.NopCloser(body),
		Request:    req,
	}
}

func TestClient_Rates_Errors(t *testing.T) {
	date := time.Date(2024, time.February, 10, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		do      doerFunc
		wantErr string
	}{
		{
			name: "transport error",
			do: func(*http.Request) (*http.Response, error) {
				return nil, errors.New("dial tcp: no route to host")
			},
			wantErr: "send request: dial tcp: no route to host",
		},
		{
			name: "status other than 200",
			do: func(req *http.Request) (*http.Response, error) {
				return httpResponse(req, http.StatusServiceUnavailable, strings.NewReader("")), nil
			},
			wantErr: "invalid response status: 503 Service Unavailable",
		},
		{
			name: "unreadable body",
			do: func(req *http.Request) (*http.Response, error) {
				return httpResponse(req, http.StatusOK, failingReader{}), nil
			},
			wantErr: "read response body: connection reset",
		},
		{
			name: "invalid JSON",
			do: func(req *http.Request) (*http.Response, error) {
				return httpResponse(req, http.StatusOK, strings.NewReader("{")), nil
			},
			wantErr: "unmarshal body to rates:",
		},
		{
			name: "more than one rate set",
			do: func(req *http.Request) (*http.Response, error) {
				return httpResponse(req, http.StatusOK, strings.NewReader(`[{"date":"a"},{"date":"b"}]`)), nil
			},
			wantErr: "response contains more than one rate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWithHTTPClient(tt.do).Rates(t.Context(), option.WithDate(date))

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
			assert.Equal(t, Rates{}, got)
		})
	}
}

func TestClient_Rates_Request(t *testing.T) {
	var got *http.Request

	do := doerFunc(func(req *http.Request) (*http.Response, error) {
		got = req

		return httpResponse(req, http.StatusOK, strings.NewReader("[]")), nil
	})

	_, err := NewWithHTTPClient(do).Rates(t.Context(),
		option.WithDate(time.Date(2024, time.February, 10, 0, 0, 0, 0, time.UTC)),
		option.WithCurrency(currencies.USD),
		option.WithCurrency(currencies.EUR),
	)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, http.MethodGet, got.Method)
	assert.Equal(t, basePath, got.URL.Scheme+"://"+got.URL.Host+got.URL.Path)
	assert.Equal(t, "2024-02-10", got.URL.Query().Get(dateParam))
	assert.Equal(t, []string{currencies.USD, currencies.EUR}, got.URL.Query()[currenciesParam])
}

func TestClient_Rates_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	do := doerFunc(func(req *http.Request) (*http.Response, error) {
		return nil, req.Context().Err()
	})

	_, err := NewWithHTTPClient(do).Rates(ctx)

	assert.ErrorIs(t, err, context.Canceled)
}
