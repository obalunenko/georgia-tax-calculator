package main

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

// addIncome answers the questions of one income entry.
func addIncome(t *testing.T, m *incomeModel, amount string) {
	t.Helper()

	send(m, enterKey(), enterKey(), enterKey())
	require.Equal(t, incomeStepAmount, m.step)

	send(m, typeKeys(amount)...)
	send(m, enterKey(), enterKey())
	require.Equal(t, incomeStepAddMore, m.step)
}

func TestTaxDetailsModelFlow(t *testing.T) {
	m := newTaxDetailsModel()
	require.NoError(t, m.err)

	m.Init()

	send(m, enterKey())
	require.Equal(t, taxStepYearIncome, m.step)
	assert.Contains(t, m.View().Content, "Selected tax type: "+taxes.TaxTypeSmallBusiness.String())
	assert.Contains(t, m.View().Content,
		"Field (15) of your previous month's declaration: income since the start of the year (GEL)")
	assert.Contains(t, m.View().Content, "Enter 0 for your first declaration of the year.")

	send(m, typeKeys("1000")...)
	assertQuit(t, send(m, enterKey()))

	assert.Equal(t, taxStepDone, m.step)
	assert.Equal(t, taxes.TaxTypeSmallBusiness.String(), m.taxType)
	assert.Equal(t, "1000", m.yearIncome)
}

func TestTaxDetailsModelEmploymentYearIncomePrompt(t *testing.T) {
	m := newTaxDetailsModel()
	m.Init()

	send(m, upKey(), enterKey())
	require.Equal(t, taxes.TaxTypeEmployment.String(), m.taxType)

	view := m.View().Content
	assert.Contains(t, view, "Income from the beginning of a calendar year (GEL)")
	assert.NotContains(t, view, "Field (15)")
}

func TestIncomeModelCollectsEntries(t *testing.T) {
	m := newIncomeModel(taxes.TaxTypeEmployment.String())
	m.Init()

	addIncome(t, m, "100")
	send(m, enterKey())
	require.Equal(t, incomeStepYear, m.step)

	addIncome(t, m, "200")
	send(m, downKey(), enterKey())
	require.Equal(t, incomeStepConfirm, m.step)

	view := m.View().Content

	assert.Contains(t, view, "Captured incomes:")
	assert.Contains(t, view, "Review your answers:")
	assert.Contains(t, view, "Amount: 200 "+currencies.USD)

	assertQuit(t, send(m, enterKey()))

	assert.Equal(t, incomeStepDone, m.step)
	require.Len(t, m.incomes, 2)
	assert.Equal(t, "100", m.incomes[0].Amount)
	assert.Equal(t, "200", m.incomes[1].Amount)
	assert.Equal(t, currencies.USD, m.incomes[1].Currency)
}

func TestIncomeModelDeclinedRestarts(t *testing.T) {
	m := newIncomeModel(taxes.TaxTypeEmployment.String())
	m.Init()

	addIncome(t, m, "100")
	send(m, downKey(), enterKey())
	send(m, downKey(), enterKey())

	assert.Equal(t, incomeStepYear, m.step)
	assert.Empty(t, m.incomes)
	assert.Equal(t, service.Income{}, m.current)
}

func TestIncomeModelOffersOnlyFirstIncomeYear(t *testing.T) {
	m := newIncomeModel(taxes.TaxTypeEmployment.String())
	m.Init()

	require.Greater(t, len(m.prompt.options), 1, "the first income can be in any year")

	addIncome(t, m, "100")
	send(m, enterKey())
	require.Equal(t, incomeStepYear, m.step)

	require.Len(t, m.prompt.options, 1)
	assert.Equal(t, m.incomes[0].Year, m.prompt.options[0].Value)

	send(m, enterKey())
	assert.Equal(t, m.incomes[0].Year, m.current.Year)
}

func TestIncomeModelDeclinedRestartOffersAllYears(t *testing.T) {
	m := newIncomeModel(taxes.TaxTypeEmployment.String())
	m.Init()

	addIncome(t, m, "100")
	send(m, downKey(), enterKey())
	send(m, downKey(), enterKey())
	require.Equal(t, incomeStepYear, m.step)

	assert.Greater(t, len(m.prompt.options), 1)
}

// enterIncomeUpToType answers the date, amount and currency of one Small Business income.
func enterIncomeUpToType(t *testing.T, m *incomeModel, amount string) {
	t.Helper()

	send(m, enterKey(), enterKey(), enterKey())
	send(m, typeKeys(amount)...)
	send(m, enterKey(), enterKey())
	require.Equal(t, incomeStepType, m.step)
}

func TestIncomeModelAsksIncomeTypeForSmallBusiness(t *testing.T) {
	m := newIncomeModel(taxes.TaxTypeSmallBusiness.String())
	m.Init()

	enterIncomeUpToType(t, m, "100")

	labels := make([]string, 0, len(m.prompt.options))
	for _, o := range m.prompt.options {
		labels = append(labels, o.Label)
	}

	assert.Equal(t, []string{"Cash (18)", "POS terminal (19)", "Non-cash (20)", "Other (21)", "Unpaid supply (22)"}, labels)
	assert.Equal(t, 0, m.prompt.cursor, "the first income starts on the first type")
	assert.Contains(t, m.View().Content, "Non-cash (20) — bank transfers")

	send(m, downKey(), downKey(), enterKey())
	require.Equal(t, incomeStepAddMore, m.step)
	assert.Equal(t, taxes.IncomeTypeNonCash.String(), m.incomes[0].Type)

	send(m, enterKey())
	enterIncomeUpToType(t, m, "50")
	assert.Equal(t, taxes.IncomeTypeNonCash.String(), m.prompt.options[m.prompt.cursor].Value,
		"the previous income's type is preselected")

	send(m, upKey(), upKey(), enterKey())
	send(m, downKey(), enterKey())
	require.Equal(t, incomeStepConfirm, m.step)

	view := m.View().Content
	assert.Contains(t, view, "— 100 "+currencies.USD+" — Non-cash (20)\n")
	assert.Contains(t, view, "— 50 "+currencies.USD+" — Cash (18)\n")
	assert.Contains(t, view, "   Type: Non-cash (20)\n")
	assert.Equal(t, taxes.IncomeTypeCash.String(), m.incomes[1].Type)
}

func TestIncomeTypeOptionsFitEightyColumns(t *testing.T) {
	const terminalWidth = 80

	for _, o := range incomeTypeOptions() {
		// The prompt renders "> <Label> — <Description>"; one option must stay one row.
		line := "> " + o.Label + " — " + o.Description
		assert.LessOrEqual(t, utf8.RuneCountInString(line), terminalWidth, line)
	}
}

func TestIncomeModelEmploymentHasNoIncomeType(t *testing.T) {
	m := newIncomeModel(taxes.TaxTypeEmployment.String())
	m.Init()

	addIncome(t, m, "100")

	assert.Empty(t, m.incomes[0].Type)
	assert.NotContains(t, m.View().Content, "Type:")
}

func TestIncomeModelRejectsUnknownTaxType(t *testing.T) {
	m := newIncomeModel("Individual Entrepreneur")

	require.ErrorIs(t, m.err, taxes.ErrInvalidTaxType)
	assertQuit(t, m.Init())
}

// addFebruaryIncome enters an income on February 1 of the previous year, which has all 12 months,
// answering the income type step when the model asks for it.
func addFebruaryIncome(t *testing.T, m *incomeModel) {
	t.Helper()

	if len(m.incomes) == 0 {
		send(m, downKey(), enterKey())
		require.Len(t, m.prompt.options, 12)
		send(m, downKey(), enterKey(), enterKey())
	} else {
		send(m, enterKey(), enterKey(), enterKey())
	}

	send(m, typeKeys("100")...)
	send(m, enterKey(), enterKey())

	if m.step == incomeStepType {
		send(m, enterKey())
	}

	require.Equal(t, incomeStepAddMore, m.step)
}

func TestIncomeModelMonthAfterFirstIncome(t *testing.T) {
	tests := []struct {
		name       string
		taxType    taxes.TaxType
		wantMonths int
	}{
		{name: "small business keeps the first income's month", taxType: taxes.TaxTypeSmallBusiness, wantMonths: 1},
		{name: "employment offers all months", taxType: taxes.TaxTypeEmployment, wantMonths: 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newIncomeModel(tt.taxType.String())
			m.Init()

			addFebruaryIncome(t, m)
			require.Equal(t, "February", m.incomes[0].Month)

			send(m, enterKey(), enterKey())
			require.Equal(t, incomeStepMonth, m.step)
			require.Len(t, m.prompt.options, tt.wantMonths)

			if tt.wantMonths == 1 {
				assert.Equal(t, "February", m.prompt.options[0].Value)
			}
		})
	}
}

func TestIncomeModelDeclinedRestartOffersAllMonths(t *testing.T) {
	m := newIncomeModel(taxes.TaxTypeSmallBusiness.String())
	m.Init()

	addFebruaryIncome(t, m)
	send(m, downKey(), enterKey())
	send(m, downKey(), enterKey())
	require.Equal(t, incomeStepYear, m.step)

	send(m, downKey(), enterKey())
	assert.Len(t, m.prompt.options, 12)
}

func TestRenderIncomesWithoutEntries(t *testing.T) {
	assert.Contains(t, renderIncomeList(nil), "none yet")
	assert.Equal(t, "No income entries captured.", renderIncomeDetails(nil))
}
