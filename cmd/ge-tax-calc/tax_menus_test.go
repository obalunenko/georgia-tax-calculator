package main

import (
	"testing"

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

	send(m, typeKeys("1000")...)
	assertQuit(t, send(m, enterKey()))

	assert.Equal(t, taxStepDone, m.step)
	assert.Equal(t, taxes.TaxTypeSmallBusiness.String(), m.taxType)
	assert.Equal(t, "1000", m.yearIncome)
}

func TestIncomeModelCollectsEntries(t *testing.T) {
	m := newIncomeModel()
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
	m := newIncomeModel()
	m.Init()

	addIncome(t, m, "100")
	send(m, downKey(), enterKey())
	send(m, downKey(), enterKey())

	assert.Equal(t, incomeStepYear, m.step)
	assert.Empty(t, m.incomes)
	assert.Equal(t, service.Income{}, m.current)
}

func TestRenderIncomesWithoutEntries(t *testing.T) {
	assert.Contains(t, renderIncomeList(nil), "none yet")
	assert.Equal(t, "No income entries captured.", renderIncomeDetails(nil))
}
