package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

// fillConvertForm answers every question up to the confirmation.
func fillConvertForm(t *testing.T, m *convertModel) {
	t.Helper()

	m.Init()

	send(m, enterKey(), enterKey(), enterKey())
	require.Equal(t, convertStepAmount, m.step)

	send(m, typeKeys("100.5")...)
	send(m, enterKey(), enterKey(), enterKey())
	require.Equal(t, convertStepConfirm, m.step)
}

func TestConvertModelConfirmed(t *testing.T) {
	m := newConvertModel()
	fillConvertForm(t, m)

	assert.Contains(t, m.View().Content, "Review your answers:")

	assertQuit(t, send(m, enterKey()))

	assert.Equal(t, convertStepDone, m.step)
	assert.NotEmpty(t, m.req.Year)
	assert.NotEmpty(t, m.req.Month)
	assert.NotEmpty(t, m.req.Day)
	assert.Equal(t, "100.5", m.req.Amount)
	assert.Equal(t, currencies.USD, m.req.CurrencyFrom)
	assert.Equal(t, currencies.GEL, m.req.CurrencyTo)
}

func TestConvertModelDeclinedRestarts(t *testing.T) {
	m := newConvertModel()
	fillConvertForm(t, m)

	send(m, downKey(), enterKey())

	assert.Equal(t, convertStepYear, m.step)
	assert.Equal(t, service.ConvertRequest{}, m.req)
}

func TestConvertModelShowsCurrentInput(t *testing.T) {
	m := newConvertModel()
	m.Init()

	send(m, enterKey(), enterKey(), enterKey())

	assert.Contains(t, m.View().Content, "Current input:\nDate: ")
}

func TestConvertModelNewPromptsFitTerminal(t *testing.T) {
	inner := newConvertModel()
	m := &fitHeightModel{Model: inner}

	m.Init()

	send(m, windowSize(testHeight))
	send(m, enterKey(), enterKey(), enterKey())
	send(m, typeKeys("5")...)
	send(m, enterKey())
	require.Equal(t, convertStepCurrencyFrom, inner.step)

	content := m.View().Content

	assert.LessOrEqual(t, strings.Count(content, "\n")+1, testHeight)
	assert.Contains(t, content, "Select currency of conversion 'from'")
	assert.Contains(t, content, "> "+currencies.USD+"\n")

	send(m, enterKey())

	content = m.View().Content

	assert.LessOrEqual(t, strings.Count(content, "\n")+1, testHeight)
	assert.Contains(t, content, "> "+currencies.GEL+"\n")
}
