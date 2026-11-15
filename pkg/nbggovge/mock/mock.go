// Package mock provides mock implementations of nbggovge.Client.
package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/internal"
)

// NewClient creates a new mock client that serves data through the real client code path.
// For each request it returns the rate set whose date matches the requested date,
// limited to the requested currencies (all currencies when none are requested).
// Like NBG, it returns an empty response when no rate set matches the date
// or none of the requested currencies is in the set. data is not modified.
func NewClient(data []nbggovge.Rates) nbggovge.Client {
	return nbggovge.NewWithHTTPClient(mockRatesHTTPClient{data: data})
}

type mockRatesHTTPClient struct {
	data []nbggovge.Rates
}

func (d mockRatesHTTPClient) Do(req *http.Request) (*http.Response, error) {
	q := req.URL.Query()

	resp := make([]nbggovge.Rates, 0, 1)

	if r, ok := d.ratesFor(q.Get(internal.DateParam), q[internal.CurrencyCodesParam]); ok {
		resp = append(resp, r)
	}

	body, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

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

// ratesFor returns the rate set for date with only the requested codes.
// It reports false when there is nothing to return.
func (d mockRatesHTTPClient) ratesFor(date string, codes []string) (nbggovge.Rates, bool) {
	idx := slices.IndexFunc(d.data, func(r nbggovge.Rates) bool {
		return date != "" && strings.HasPrefix(r.Date, date)
	})
	if idx < 0 {
		return nbggovge.Rates{}, false
	}

	r := d.data[idx]

	if len(codes) == 0 {
		return r, true
	}

	currs := make([]nbggovge.Currency, 0, len(codes))

	for _, c := range r.Currencies {
		if slices.Contains(codes, c.Code) {
			currs = append(currs, c)
		}
	}

	if len(currs) == 0 {
		return nbggovge.Rates{}, false
	}

	r.Currencies = currs

	return r, true
}
