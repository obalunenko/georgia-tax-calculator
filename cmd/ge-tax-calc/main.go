package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

func main() {
	// TODO:
	// 	- get currency rates for date
	// 	- calc sum in gel for date rates
	// 	- get sum of taxes

	resp, err := getCurrencyRatesByDate(time.Now(), currCodeEUR, currCodeUSD)
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.MarshalIndent(&resp, " ", " ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
}

type Rates []RatesResponse

func UnmarshalRates(data []byte) (Rates, error) {
	var r Rates

	err := json.Unmarshal(data, &r)

	return r, err
}

func (r *Rates) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type RatesResponse struct {
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

type currCode string

func (c currCode) String() string {
	return string(c)
}

const (
	currCodeUSD = "usd"
	currCodeEUR = "eur"
	currCodeGBP = "gbp"
	currCodeBYN = "byn"
)

func getCurrencyRatesByDate(date time.Time, codes ...currCode) (Rates, error) {
	const (
		basePath        = "https://nbg.gov.ge/gw/api/ct/monetarypolicy/currencies/en/json"
		currenciesParam = "currencies"
		dateParam       = "date"
		dateLayout      = "2006-01-02"
	)

	u, err := url.Parse(basePath)
	if err != nil {
		return nil, err
	}

	q := u.Query()

	if len(codes) < 1 {
		return nil, fmt.Errorf("at least one currency code should be passed")
	}

	for i := range codes {
		q.Add(currenciesParam, codes[i].String())
	}

	q.Add(dateParam, date.Format(dateLayout))
	u.RawQuery = q.Encode()

	method := http.MethodGet

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	client := http.DefaultClient

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(body))

	resp, err := UnmarshalRates(body)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

	fmt.Println(string(b))
}
