package nbggovge

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	log "github.com/obalunenko/logger"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/internal"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

// Client is a contract for nbg.gov.ge API.
type Client interface {
	// Rates returns RatesResponse for today by default for a list of currency codes set up by options.
	Rates(ctx context.Context, opts ...option.RatesOption) (RatesResponse, error)
}

// HTTPClient is and interface for mocking sending http requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// New returns nbg.gov.ge API client.
func New() Client {
	return NewWithHTTPClient(http.DefaultClient)
}

// NewWithHTTPClient returns nbg.gov.ge API client with specified http client.
func NewWithHTTPClient(cli HTTPClient) Client {
	return client{
		HTTPClient: cli,
	}
}

type client struct {
	HTTPClient
}

// Rates fetches rates, list of currencies and date could be set by optional option.RatesOption.
// By default, it fetches all currencies for today.
func (c client) Rates(ctx context.Context, opts ...option.RatesOption) (RatesResponse, error) {
	var (
		params internal.RatesParams
	)

	for _, opt := range opts {
		opt.Apply(&params)
	}

	if params.Date.IsZero() {
		params.Date = time.Now()
	}

	const (
		basePath        = "https://nbg.gov.ge/gw/api/ct/monetarypolicy/currencies/en/json"
		currenciesParam = "currencies"
		dateParam       = "date"
		dateLayout      = "2006-01-02"
	)

	u, err := url.Parse(basePath)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	q := u.Query()

	for _, code := range params.CurrencyCodes {
		q.Add(currenciesParam, code)
	}

	q.Add(dateParam, params.Date.Format(dateLayout))

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return NilRatesResponse, fmt.Errorf("create request: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return NilRatesResponse, fmt.Errorf("send request: %w", err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.WithError(ctx, err).Error("Failed to close response body.")
		}
	}()

	if res.StatusCode != http.StatusOK {
		return NilRatesResponse, fmt.Errorf("invalid response status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return NilRatesResponse, fmt.Errorf("read response body: %w", err)
	}

	resp, err := UnmarshalRatesResponse(body)
	if err != nil {
		return NilRatesResponse, fmt.Errorf("unmarshal body to rates: %w", err)
	}

	return resp, nil
}
