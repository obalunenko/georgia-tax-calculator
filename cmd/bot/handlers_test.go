package main

import (
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/mymmrac/telego"
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

	text, _ := formatCalcResult(resp)
	assert.NotContains(t, text, "⚠️")

	resp.LimitWarning = &service.LimitWarning{
		Limit:      models.NewMoney(500_000, currencies.GEL),
		YearIncome: models.NewMoney(501_000, currencies.GEL),
		Excess:     models.NewMoney(1000, currencies.GEL),
	}

	got, _ := formatCalcResult(resp)

	assert.True(t, strings.HasSuffix(got, "Taxes to Pay: 20 GEL\n⚠️ "+resp.LimitWarning.String()), got)

	resp.Declaration = testDeclaration()

	got, _ = formatCalcResult(resp)

	assert.True(t, strings.HasSuffix(got, "(22) Unpaid supply: 0.00\n⚠️ "+resp.LimitWarning.String()), got,
		"the warning stays after the declaration block")
}

// testDeclaration returns the declaration of spec US1.1.
func testDeclaration() *service.Declaration {
	fields := []service.DeclarationField{
		{Number: 15, Label: "Income since the start of the year", Value: 35942},
		{Number: 16, Label: "Tax rate, %", Value: 1},
		{Number: 17, Label: "Taxable income for the month", Value: 4150},
		{Number: 18, Label: "Cash", Value: 3950},
		{Number: 19, Label: "POS terminal"},
		{Number: 20, Label: "Non-cash", Value: 200},
		{Number: 21, Label: "Other"},
		{Number: 22, Label: "Unpaid supply"},
	}

	return &service.Declaration{Fields: fields}
}

// entityTexts returns the text each entity covers, cut by UTF-16 offsets as Telegram does.
func entityTexts(text string, entities []telego.MessageEntity) []string {
	units := utf16.Encode([]rune(text))

	out := make([]string, 0, len(entities))
	for _, e := range entities {
		out = append(out, string(utf16.Decode(units[e.Offset:e.Offset+e.Length])))
	}

	return out
}

func TestFormatCalcResultDeclaration(t *testing.T) {
	resp := &service.CalculateResponse{
		TaxRate:    taxes.TaxRate{Type: taxes.TaxTypeSmallBusiness, Rate: 0.01},
		YearIncome: models.NewMoney(35942, currencies.GEL),
		Incomes: []service.IncomeResponse{{
			Amount:    models.NewMoney(3950, currencies.GEL),
			Converted: models.NewMoney(3950, currencies.GEL),
			Rate:      models.NewMoney(1, ""),
			Type:      taxes.IncomeTypeCash,
		}},
		TotalIncomeConverted: models.NewMoney(4150, currencies.GEL),
		Tax:                  models.NewMoney(41.5, currencies.GEL),
		Declaration:          testDeclaration(),
	}

	text, entities := formatCalcResult(resp)

	assert.Contains(t, text, "  1) 3950 GEL → 3950 GEL (rate: 1) — Cash (18)\n")
	assert.True(t, strings.HasSuffix(text, "Taxes to Pay: 41.5 GEL\n\n"+resp.Declaration.String()), text,
		"the bot prints the same block lines as the CLI")

	require.Len(t, entities, 8)

	for _, e := range entities {
		assert.Equal(t, telego.EntityTypeCode, e.Type)
	}

	assert.Equal(t,
		[]string{"35942.00", "1.00", "4150.00", "3950.00", "0.00", "200.00", "0.00", "0.00"},
		entityTexts(text, entities),
		"entity offsets count the emoji before them in UTF-16 code units",
	)
}

func TestFormatCalcResultEmployment(t *testing.T) {
	resp := &service.CalculateResponse{
		TaxRate:    taxes.TaxRate{Type: taxes.TaxTypeEmployment, Rate: 0.2},
		YearIncome: models.NewMoney(1100, currencies.GEL),
		Incomes: []service.IncomeResponse{{
			Amount:    models.NewMoney(100, currencies.USD),
			Converted: models.NewMoney(270, currencies.GEL),
			Rate:      models.NewMoney(2.7, ""),
		}},
		TotalIncomeConverted: models.NewMoney(270, currencies.GEL),
		Tax:                  models.NewMoney(54, currencies.GEL),
	}

	text, entities := formatCalcResult(resp)

	assert.Equal(t, "🧾 Tax Calculation Result\n\n"+
		"Tax Rate: Employment 20 %\n"+
		"Year Income: 1100 GEL\n\n"+
		"Incomes converted to GEL:\n"+
		"  1) 100 USD → 270 GEL (rate: 2.7)\n\n"+
		"Total Income (GEL): 270 GEL\n"+
		"Taxes to Pay: 54 GEL", text, "the Employment result keeps the pre-feature format")
	assert.Empty(t, entities)
}

func TestUTF16Len(t *testing.T) {
	assert.Equal(t, 0, utf16Len(""))
	assert.Equal(t, 4, utf16Len("1.00"))
	assert.Equal(t, 2, utf16Len("🧾"), "a character outside the BMP is two code units")
	assert.Equal(t, 2, utf16Len("⚠️"), "U+26A0 and the variation selector")
}
