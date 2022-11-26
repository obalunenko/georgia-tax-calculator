package nbggovge

import (
	"encoding/json"
	"errors"
	"strings"
)

// RatesResponse represents response.
type RatesResponse []Rates

// NilRatesResponse is a shortcut for empty RatesResponse.
var NilRatesResponse = RatesResponse{}

var (
	// ErrCodeNotFound returned when specified code could not be found in set.
	ErrCodeNotFound = errors.New("code not found in set")
)

// CurrencyByCode returns Currency from set by specified code.
// When no currency in set - ErrCodeNotFound returned.
func (r RatesResponse) CurrencyByCode(code string) (Currency, error) {
	rates := r[0]

	for _, currency := range rates.Currencies {
		if strings.EqualFold(currency.Code, code) {
			return currency, nil
		}
	}

	return Currency{}, ErrCodeNotFound
}

// UnmarshalRatesResponse parses json to RatesResponse.
func UnmarshalRatesResponse(data []byte) (RatesResponse, error) {
	var r RatesResponse

	err := json.Unmarshal(data, &r)

	return r, err
}

// Marshal marshals data from an Rates to bytes.
func (r *Rates) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Rates represents set of rates.
type Rates struct {
	Date       string     `json:"date"`
	Currencies []Currency `json:"currencies"`
}

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
