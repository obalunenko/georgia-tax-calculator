package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/internal/models"
	"github.com/obalunenko/georgia-tax-calculator/internal/service"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
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

func TestFormatCalcResultLimitWarning(t *testing.T) {
	resp := &service.CalculateResponse{
		TaxRate:              taxes.TaxRate{Type: taxes.TaxTypeSmallBusiness, Rate: 0.01},
		YearIncome:           models.NewMoney(501_000, currencies.GEL),
		TotalIncomeConverted: models.NewMoney(2000, currencies.GEL),
		Tax:                  models.NewMoney(20, currencies.GEL),
	}

	assert.NotContains(t, formatCalcResult(resp), "⚠️")

	resp.LimitWarning = &service.LimitWarning{
		Limit:      models.NewMoney(500_000, currencies.GEL),
		YearIncome: models.NewMoney(501_000, currencies.GEL),
		Excess:     models.NewMoney(1000, currencies.GEL),
	}

	got := formatCalcResult(resp)

	assert.True(t, strings.HasSuffix(got, "Taxes to Pay: 20 GEL\n⚠️ "+resp.LimitWarning.String()), got)
}
