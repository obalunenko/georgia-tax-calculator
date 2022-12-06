package nbggovge

import (
	"encoding/json"
	"errors"
	"strings"
)

var (
	// ErrCodeNotFound returned when specified code could not be found in set.
	ErrCodeNotFound = errors.New("code not found in set")
)

// Currency represents rates for one currency.
type Currency struct {
	Code          string  `json:"code"`
	Quantity      int64   `json:"quantity"`
	RateFormated  string  `json:"rateFormated"`
	DiffFormated  string  `json:"diffFormated"`
	Rate          float64 `json:"rate"`
	Name          string  `json:"name"`
	Diff          float64 `json:"diff"`
	Date          string  `json:"date"`
	ValidFromDate string  `json:"validFromDate"`
}

// Rates represents set of rates.
type Rates struct {
	Date       string     `json:"date"`
	Currencies []Currency `json:"currencies"`
}

// CurrencyByCode returns Currency from set by specified code.
// When no currency in set - ErrCodeNotFound returned.
func (r Rates) CurrencyByCode(code string) (Currency, error) {
	for _, currency := range r.Currencies {
		if strings.EqualFold(currency.Code, code) {
			return currency, nil
		}
	}

	return Currency{}, ErrCodeNotFound
}

// ratesResponse represents response.
type ratesResponse []Rates

func (r ratesResponse) Rates() Rates {
	if len(r) == 0 {
		return Rates{}
	}

	return r[0]
}

// unmarshalRatesResponse parses json to ratesResponse.
func unmarshalRatesResponse(data []byte) (ratesResponse, error) {
	var r ratesResponse

	err := json.Unmarshal(data, &r)

	return r, err
}
