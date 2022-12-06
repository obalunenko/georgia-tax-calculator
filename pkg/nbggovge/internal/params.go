// Package internal holds internal logic and parameters for requests.
package internal

import (
	"time"
)

// RatesParams represents request parameters for nbggovge rates request.
type RatesParams struct {
	Date          time.Time
	CurrencyCodes []string
}
