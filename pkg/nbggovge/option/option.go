package option

import (
	"time"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/internal"
)

// RatesOption represents optional pattern for nbgogvge rates client.
type RatesOption interface {
	Apply(params *internal.RatesParams)
}

type withDate time.Time

func (w withDate) Apply(p *internal.RatesParams) {
	p.Date = time.Time(w)
}

// WithDate adds date to options.
func WithDate(date time.Time) RatesOption {
	return withDate(date)
}

type withCurrency string

func (w withCurrency) Apply(p *internal.RatesParams) {
	c := string(w)
	if len(p.CurrencyCodes) == 0 {
		p.CurrencyCodes = []string{c}

		return
	}

	var exist bool

	for _, code := range p.CurrencyCodes {
		if code == c {
			exist = true

			break
		}
	}

	if !exist {
		p.CurrencyCodes = append(p.CurrencyCodes, c)
	}
}

// WithCurrency adds currency code to options.
func WithCurrency(code string) RatesOption {
	return withCurrency(code)
}
