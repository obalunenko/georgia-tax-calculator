package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

func TestMoney_String(t *testing.T) {
	type fields struct {
		Amount   float64
		Currency string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "",
			fields: fields{
				Amount:   25.26789,
				Currency: currencies.GEL,
			},
			want: "25.26789 GEL",
		},
		{
			name: "",
			fields: fields{
				Amount:   25.21289,
				Currency: currencies.GEL,
			},
			want: "25.21289 GEL",
		},
		{
			name: "",
			fields: fields{
				Amount:   25.21489,
				Currency: currencies.GEL,
			},
			want: "25.21489 GEL",
		},
		{
			name: "",
			fields: fields{
				Amount:   25.21489,
				Currency: "",
			},
			want: "25.21489",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewMoney(tt.fields.Amount, tt.fields.Currency)

			assert.Equalf(t, tt.want, r.String(), "String()")
		})
	}
}
