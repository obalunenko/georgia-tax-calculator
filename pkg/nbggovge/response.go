package nbggovge

import (
	"encoding/json"
	"errors"
	"strings"
)

type RatesResponse []Rates

var NilRatesResponse = RatesResponse{}

var (
	ErrCodeNotFound = errors.New("code not found in set")
)

func (r RatesResponse) CurrencyByCode(code string) (Currency, error) {
	rates := r[0]

	for _, currency := range rates.Currencies {
		if strings.EqualFold(currency.Code, code) {
			return currency, nil
		}
	}

	return Currency{}, ErrCodeNotFound
}

func UnmarshalRatesResponse(data []byte) (RatesResponse, error) {
	var r RatesResponse

	err := json.Unmarshal(data, &r)

	return r, err
}

func (r *Rates) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Rates struct {
	Date       string     `json:"date"`
	Currencies []Currency `json:"currencies"`
}

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
