package dateutils

import (
	"fmt"
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

func TestDaysList(t *testing.T) {
	type args struct {
		num int
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "",
			args: args{
				num: 5,
			},
			want: []string{"01", "02", "03", "04", "05"},
		},
		{
			name: "",
			args: args{
				num: 0,
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, DaysList(tt.args.num), "DaysList(%v)", tt.args.num)
		})
	}
}

func TestParseMonth(t *testing.T) {
	type args struct {
		raw string
	}

	tests := []struct {
		name    string
		args    args
		want    time.Month
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "",
			args: args{
				raw: "December",
			},
			want:    time.December,
			wantErr: assert.NoError,
		},
		{
			name: "",
			args: args{
				raw: "January",
			},
			want:    time.January,
			wantErr: assert.NoError,
		},
		{
			name: "",
			args: args{
				raw: "Jan",
			},
			want:    time.Month(0),
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMonth(tt.args.raw)
			if !tt.wantErr(t, err, fmt.Sprintf("ParseMonth(%v)", tt.args.raw)) {
				return
			}

			assert.Equalf(t, tt.want, got, "ParseMonth(%v)", tt.args.raw)
		})
	}
}

func TestGetMonths(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "",
			want: []string{
				"January", "February", "March", "April", "May", "June", "July", "August",
				"September", "October", "November", "December",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetMonths(), "GetMonths()")
		})
	}
}

func TestParseYear(t *testing.T) {
	type args struct {
		raw string
	}

	tests := []struct {
		name    string
		args    args
		want    int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "",
			args: args{
				raw: "1992",
			},
			want:    1992,
			wantErr: assert.NoError,
		},
		{
			name: "",
			args: args{
				raw: "1",
			},
			want:    1,
			wantErr: assert.NoError,
		},
		{
			name: "",
			args: args{
				raw: "-1992",
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "",
			args: args{
				raw: "19s92",
			},
			want:    0,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseYear(tt.args.raw)
			if !tt.wantErr(t, err, fmt.Sprintf("ParseYear(%v)", tt.args.raw)) {
				return
			}

			assert.Equalf(t, tt.want, got, "ParseYear(%v)", tt.args.raw)
		})
	}
}

func TestDaysInMonthTillDate(t *testing.T) {
	type args struct {
		m    time.Month
		year int
		now  time.Time
	}

	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "2022/10; now is 2022/12/18 - 31",
			args: args{
				m:    time.October,
				year: 2022,
				now:  time.Date(2022, 12, 18, 0, 0, 0, 0, time.Local),
			},
			want: 31,
		},
		{
			name: "2022/12; now is 2022/12/18 - 18",
			args: args{
				m:    time.December,
				year: 2022,
				now:  time.Date(2022, 12, 18, 0, 0, 0, 0, time.Local),
			},
			want: 18,
		},
		{
			name: "2021/12; now is 2022/12/18 - 18",
			args: args{
				m:    time.December,
				year: 2021,
				now:  time.Date(2022, 12, 18, 0, 0, 0, 0, time.Local),
			},
			want: 31,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DaysInMonthTillDate(tt.args.m, tt.args.year, tt.args.now)

			assert.Equalf(t, tt.want, got, "DaysInMonthTillDate(%v, %v, %v)", tt.args.m, tt.args.year, tt.args.now)
		})
	}
}
