package moneyutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToStringFixed(t *testing.T) {
	tests := []struct {
		name   string
		v      float64
		places int32
		want   string
	}{
		{name: "integer", v: 4150, places: 2, want: "4150.00"},
		{name: "zero", v: 0, places: 2, want: "0.00"},
		{name: "two decimals", v: 1500.55, places: 2, want: "1500.55"},
		{name: "rate", v: 3, places: 2, want: "3.00"},
		{name: "half rounds up", v: 0.005, places: 2, want: "0.01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ToStringFixed(tt.v, tt.places))
		})
	}
}
