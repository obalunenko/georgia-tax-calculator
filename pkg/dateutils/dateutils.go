package dateutils

import (
	"time"
)

// DaysInMonth return number of days in Month for specific year.
// The reason it works is that we generate a date one month on from the target one (m+1), but set the day of month to 0.
// Days are 1-indexed, so this has the effect of rolling back one day to the last day of the previous month
// (our target month of m). Calling Day() then procures the number we want.
func DaysInMonth(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
