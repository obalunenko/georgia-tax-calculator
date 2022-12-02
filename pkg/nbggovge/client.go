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
	// Rates returns ratesResponse for today by default for a list of currency codes set up by options.
	Rates(ctx context.Context, opts ...option.RatesOption) (Rates, error)
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

const (
	basePath        = "https://nbg.gov.ge/gw/api/ct/monetarypolicy/currencies/en/json"
	currenciesParam = "currencies"
	dateParam       = "date"
	dateLayout      = "2006-01-02"
)

// Rates fetches rates, list of currencies and date could be set by optional option.RatesOption.
// By default, it fetches all currencies for today.
func (c client) Rates(ctx context.Context, opts ...option.RatesOption) (Rates, error) {
	var (
		params internal.RatesParams
	)

	for _, opt := range opts {
		opt.Apply(&params)
	}

	if params.Date.IsZero() {
		params.Date = time.Now()
	}

	u, err := url.Parse(basePath)
	if err != nil {
		return Rates{}, fmt.Errorf("parse base url: %w", err)
	}

	q := u.Query()

	for _, code := range params.CurrencyCodes {
		q.Add(currenciesParam, code)
	}

	q.Add(dateParam, params.Date.Format(dateLayout))

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return Rates{}, fmt.Errorf("create request: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return Rates{}, fmt.Errorf("send request: %w", err)
	}

	defer func() {
		if err = res.Body.Close(); err != nil {
			log.WithError(ctx, err).Error("Failed to close response body.")
		}
	}()

	if res.StatusCode != http.StatusOK {
		return Rates{}, fmt.Errorf("invalid response status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Rates{}, fmt.Errorf("read response body: %w", err)
	}

	resp, err := unmarshalRatesResponse(body)
	if err != nil {
		return Rates{}, fmt.Errorf("unmarshal body to rates: %w", err)
	}

	return resp.Rates(), nil
}
