// Package mock provides mock implementations of nbggovge.Client.
package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/internal"
)

// NewClient creates a new mock client.
// It returns a client that returns data for each request.
// If currency is not found in data, it returns http.StatusNotFound.
// if date is empty, it returns current date in response.
func NewClient(data []nbggovge.Rates) nbggovge.Client {
	return nbggovge.NewWithHTTPClient(mockRatesHTTPClient{data: data})
}

type mockRatesHTTPClient struct {
	data []nbggovge.Rates
}

func (d mockRatesHTTPClient) Do(req *http.Request) (*http.Response, error) {
	q := req.URL.Query()

	cur := q.Get(internal.CurrencyCodesParam)
	date := q.Get(internal.CurrencyCodesParam)

	if date == "" {
		date = time.Now().Format(internal.DateLayout)
	}

	rates := d.data

	for i := range rates {
		r := rates[i]

		if r.Date == date {
			continue
		}

		r.Date = date
		for j := range r.Currencies {
			c := r.Currencies[j]
			c.Date = date
			c.ValidFromDate = date
			r.Currencies[j] = c
		}

		rates[i] = r
	}

	body, err := json.Marshal(rates)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	status := http.StatusOK

	if !slices.Contains(currencies.All(), cur) {
		status = http.StatusNotFound
		body = nil
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
