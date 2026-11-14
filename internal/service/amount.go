package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/obalunenko/georgia-tax-calculator/pkg/moneyutils"
)

// ErrInvalidAmount returned when an amount is empty, is not a number, or is out of range.
var ErrInvalidAmount = errors.New("invalid amount")

// ParseAmount parses an income or conversion amount. The amount must be greater than 0.
// The CLI, the bot, and the service use it so that all of them accept the same input.
func ParseAmount(raw string) (float64, error) {
	v, err := parseMoney(raw)
	if err != nil {
		return 0, err
	}

	if v <= 0 {
		return 0, fmt.Errorf("%w: must be greater than 0", ErrInvalidAmount)
	}

	return v, nil
}

// ParseYearIncome parses the income since the start of the calendar year in GEL.
// It must be 0 or greater: 0 is valid for the first income of the year.
func ParseYearIncome(raw string) (float64, error) {
	v, err := parseMoney(raw)
	if err != nil {
		return 0, err
	}

	if v < 0 {
		return 0, fmt.Errorf("%w: must be 0 or greater", ErrInvalidAmount)
	}

	return v, nil
}

func parseMoney(raw string) (float64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, fmt.Errorf("%w: value is required", ErrInvalidAmount)
	}

	v, err := moneyutils.Parse(s)
	if err != nil {
		return 0, fmt.Errorf("%w: %q is not a number", ErrInvalidAmount, s)
	}

	return v, nil
}
