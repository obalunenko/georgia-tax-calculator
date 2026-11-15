package service

import (
	"errors"
	"fmt"
	"time"
)

// FirstYear is the first year both interfaces offer for income and conversion dates.
const FirstYear = 2016

// tbilisiUTCOffset is Georgia's time zone: UTC+4 without daylight saving time since 2005.
// A fixed zone avoids depending on tzdata, which the bot's Alpine image does not ship.
const tbilisiUTCOffset = 4 * time.Hour

var tbilisi = time.FixedZone("Asia/Tbilisi", int(tbilisiUTCOffset.Seconds()))

// now returns the current time in Georgia; tests replace it.
var now = func() time.Time {
	return time.Now().In(tbilisi)
}

// Now returns the current time in Georgia. "Today" for income and conversion dates is the Georgian date,
// whatever the time zone of the process.
func Now() time.Time {
	return now()
}

// ErrInvalidDate returned when a date does not exist, is before FirstYear, or is after today in Georgia.
var ErrInvalidDate = errors.New("invalid date")

// newDate returns the UTC date of year, month, and day after checking it against ErrInvalidDate rules.
func newDate(year int, month time.Month, day int) (time.Time, error) {
	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	// time.Date normalizes, e.g. 31 April becomes 1 May.
	if date.Year() != year || date.Month() != month || date.Day() != day {
		return time.Time{}, fmt.Errorf("%w: %d %s %d does not exist", ErrInvalidDate, day, month, year)
	}

	if year < FirstYear {
		return time.Time{}, fmt.Errorf("%w: %s is before %d", ErrInvalidDate, date.Format(layout), FirstYear)
	}

	t := Now()
	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

	if date.After(today) {
		return time.Time{}, fmt.Errorf("%w: %s is in the future", ErrInvalidDate, date.Format(layout))
	}

	return date, nil
}
