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

type Rates []RatesResponse

func UnmarshalExpired(data []byte) (Rates, error) {
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

func main() {
	// TODO:
	// 	- get currency rates for date
	// 	- calc sum in gel for date rates
	// 	- get sum of taxes
	fmt.Println("hello")

	const (
		basePath   = "https://nbg.gov.ge/gw/api/ct/monetarypolicy/currencies/en/json"
		currencies = "currencies"
		eur        = "eur"
		usd        = "usd"
		gbp        = "gbp"
		byn        = "byn"
		date       = "date"
		layout     = "2006-01-02"
	)
	u, err := url.Parse(basePath)
	if err != nil {
		log.Fatal(err)
	}

	q := u.Query()
	q.Add(currencies, eur)
	q.Add(currencies, usd)
	q.Add(currencies, gbp)
	q.Add(currencies, byn)
	q.Add(date, time.Now().Format(layout))
	// q.Set(date, time.Now().AddDate(-5, 0, -15).Format(layout))
	u.RawQuery = q.Encode()

	method := http.MethodGet

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	client := http.DefaultClient

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(body))

	resp, err := UnmarshalExpired(body)
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.MarshalIndent(&resp, " ", " ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
}
