package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

// setNow fixes the current time in Georgia for one test.
func setNow(t *testing.T, at time.Time) {
	t.Helper()

	prev := now
	now = func() time.Time { return at.In(tbilisi) }

	t.Cleanup(func() { now = prev })
}

func TestNow(t *testing.T) {
	_, offset := Now().Zone()

	assert.Equal(t, 4*60*60, offset, "Georgia is UTC+4")
}

func TestNewDate(t *testing.T) {
	// 22:30 UTC on 14 November is already 15 November in Tbilisi.
	setNow(t, time.Date(2026, time.November, 14, 22, 30, 0, 0, time.UTC))

	tests := []struct {
		name    string
		year    int
		month   time.Month
		day     int
		wantErr bool
	}{
		{name: "past date", year: 2022, month: time.December, day: 8},
		{name: "first allowed year", year: 2016, month: time.January, day: 1},
		{name: "leap day", year: 2024, month: time.February, day: 29},
		{name: "today in Georgia while it is yesterday in UTC", year: 2026, month: time.November, day: 15},
		{name: "tomorrow in Georgia", year: 2026, month: time.November, day: 16, wantErr: true},
		{name: "no leap day", year: 2023, month: time.February, day: 29, wantErr: true},
		{name: "31 April", year: 2024, month: time.April, day: 31, wantErr: true},
		{name: "day 0", year: 2024, month: time.April, day: 0, wantErr: true},
		{name: "before the first year", year: 2015, month: time.December, day: 31, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newDate(tt.year, tt.month, tt.day)
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrInvalidDate)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, time.Date(tt.year, tt.month, tt.day, 0, 0, 0, 0, time.UTC), got)
		})
	}
}

func TestConvertRejectsInvalidDate(t *testing.T) {
	setNow(t, time.Date(2026, time.November, 15, 12, 0, 0, 0, time.UTC))

	s := service{c: failingConverter{t: t}}

	_, err := s.Convert(t.Context(), ConvertRequest{
		Year: "2024", Month: time.April.String(), Day: "31",
		CurrencyFrom: currencies.EUR,
		CurrencyTo:   currencies.GEL,
		Amount:       "100",
	})

	assert.ErrorIs(t, err, ErrInvalidDate)
}
