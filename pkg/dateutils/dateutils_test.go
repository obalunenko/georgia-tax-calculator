package dateutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDaysInMonth(t *testing.T) {
	type args struct {
		m    time.Month
		year int
	}

	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "2022 november - 30",
			args: args{
				m:    time.November,
				year: 2022,
			},
			want: 30,
		},
		{
			name: "2022 december - 31",
			args: args{
				m:    time.December,
				year: 2022,
			},
			want: 31,
		},
		{
			name: "2022 february - 28",
			args: args{
				m:    time.February,
				year: 2022,
			},
			want: 28,
		},
		{
			name: "2024 february - 29",
			args: args{
				m:    time.February,
				year: 2024,
			},
			want: 29,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DaysInMonth(tt.args.m, tt.args.year)

			assert.Equal(t, tt.want, got)
		})
	}
}
