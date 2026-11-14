package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMoney(t *testing.T) {
	tests := []struct {
		val     string
		wantErr bool
	}{
		{val: "1500.00"},
		{val: " 12 "},
		{val: "", wantErr: true},
		{val: "12abc", wantErr: true},
		{val: "1,500", wantErr: true},
		{val: "NaN", wantErr: true},
		{val: "-Inf", wantErr: true},
		{val: "0", wantErr: true},
		{val: "-5", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			assert.Equal(t, tt.wantErr, validateMoney(tt.val) != nil)
		})
	}
}

func TestValidateYearIncome(t *testing.T) {
	tests := []struct {
		val     string
		wantErr bool
	}{
		{val: "0"},
		{val: "0.00"},
		{val: "1500"},
		{val: "-1", wantErr: true},
		{val: "12abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			assert.Equal(t, tt.wantErr, validateYearIncome(tt.val) != nil)
		})
	}
}

func TestInvalidAmountText(t *testing.T) {
	err := validateMoney("0")
	require.Error(t, err)

	assert.Equal(t,
		"❌ Invalid amount: must be greater than 0\n\nPlease enter a valid number (e.g. 1500.00):",
		invalidAmountText(err),
	)
}
