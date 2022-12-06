// Package dateutils provides functionality for working with dates.
package dateutils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DaysInMonth return number of days in Month for specific year.
// The reason it works is that we generate a date one month on from the target one (m+1), but set the day of month to 0.
// Days are 1-indexed, so this has the effect of rolling back one day to the last day of the previous month
// (our target month of m). Calling Day() then procures the number we want.
func DaysInMonth(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// GetMonths returns list of all month.
func GetMonths() []string {
	const totalMonth = 12

	months := make([]string, 0, totalMonth)

	for i := time.January; i <= time.December; i++ {
		months = append(months, i.String())
	}

	return months
}

// ErrIncorrectMonth returned when month is incorrect.
var ErrIncorrectMonth = errors.New("incorrect month")

// ParseMonth parses month from string.
func ParseMonth(raw string) (time.Month, error) {
	for i := time.January; i <= time.December; i++ {
		if isMonth(raw, i) {
			return i, nil
		}
	}

	return 0, fmt.Errorf("%s: %w", raw, ErrIncorrectMonth)
}

// ParseDay parses day from string.
func ParseDay(raw string) (int, error) {
	d, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}

	if d < 0 {
		return 0, fmt.Errorf("should be >0: %w", ErrInvalidDay)
	}

	const maxdays = 31

	if d > maxdays {
		return 0, fmt.Errorf("should be <31: %w", ErrInvalidDay)
	}

	return d, nil
}

func isMonth(raw string, m time.Month) bool {
	raw = strings.TrimSpace(raw)

	return strings.EqualFold(raw, m.String())
}

var (
	// ErrInvalidYear returned when year is invalid.
	ErrInvalidYear = errors.New("invalid year")
	// ErrInvalidDay returned when day is invalid.
	ErrInvalidDay = errors.New("invalid day")
)

// ParseYear parses year from string.
func ParseYear(raw string) (int, error) {
	y, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}

	if y < 0 {
		return 0, fmt.Errorf("should be >0: %w", ErrInvalidYear)
	}

	return y, nil
}

// DaysList returns list of days with specified number of days.
func DaysList(num int) []string {
	res := make([]string, 0, num)

	for i := 1; i <= num; i++ {
		d := fmt.Sprintf("%02d", i)

		res = append(res, d)
	}

	return res
}
