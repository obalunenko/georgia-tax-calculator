package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

func runConvertMenu() (service.ConvertRequest, error) {
	model := newConvertModel()
	if _, err := tea.NewProgram(model).Run(); err != nil {
		return service.ConvertRequest{}, err
	}

	if model.err != nil {
		return service.ConvertRequest{}, model.err
	}

	return model.req, nil
}

type convertStep int

const (
	convertStepYear convertStep = iota
	convertStepMonth
	convertStepDay
	convertStepAmount
	convertStepCurrencyFrom
	convertStepCurrencyTo
	convertStepConfirm
	convertStepDone
)

type convertModel struct {
	step   convertStep
	prompt *promptModel
	req    service.ConvertRequest
	err    error
}

func newConvertModel() *convertModel {
	return &convertModel{}
}

func (m *convertModel) Init() tea.Cmd {
	if m.err != nil {
		return tea.Quit
	}

	if m.prompt == nil {
		return m.setPrompt(newSelectPrompt("Select year of conversion", yearOptions(), defaultYearValue()))
	}

	return m.prompt.Init()
}

func (m *convertModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		return m, tea.Quit
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		if key.Type == tea.KeyCtrlC {
			m.err = errUserAborted
			return m, tea.Quit
		}
	}

	cmd := m.prompt.Update(msg)
	if m.prompt.Completed() {
		return m, m.advance()
	}

	return m, cmd
}

func (m *convertModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("error: %v\n", m.err)
	}

	var b strings.Builder

	if summary := renderConvertSummary(m.req); summary != "" && m.step != convertStepYear && m.step != convertStepConfirm {
		b.WriteString("Current input:\n")
		b.WriteString(summary)
		b.WriteString("\n\n")
	}

	if m.step == convertStepConfirm {
		b.WriteString("Review your answers:\n\n")
		b.WriteString(renderConvertSummary(m.req))
		b.WriteString("\n\n")
	}

	if m.prompt != nil {
		b.WriteString(m.prompt.View())
	}

	return b.String()
}

func (m *convertModel) advance() tea.Cmd {
	switch m.step {
	case convertStepYear:
		m.req.Year = m.prompt.Value()
		m.step = convertStepMonth

		opts, err := monthOptions(m.req.Year)
		if err != nil {
			m.err = err
			return tea.Quit
		}

		return m.setPrompt(newSelectPrompt("Select month of conversion", opts, defaultMonthValue(m.req.Year)))
	case convertStepMonth:
		m.req.Month = m.prompt.Value()
		m.step = convertStepDay

		opts, err := dayOptions(m.req.Year, m.req.Month)
		if err != nil {
			m.err = err
			return tea.Quit
		}

		return m.setPrompt(newSelectPrompt("Select day of conversion", opts, defaultDayValue(m.req.Year, m.req.Month)))
	case convertStepDay:
		m.req.Day = m.prompt.Value()
		m.step = convertStepAmount

		return m.setPrompt(newInputPrompt("Input amount to convert", "0.00", "", validateMoneyInput))
	case convertStepAmount:
		m.req.Amount = m.prompt.Value()
		m.step = convertStepCurrencyFrom

		return m.setPrompt(newSelectPrompt("Select currency of conversion 'from'", currencyOptions(), currencies.USD))
	case convertStepCurrencyFrom:
		m.req.CurrencyFrom = m.prompt.Value()
		m.step = convertStepCurrencyTo

		return m.setPrompt(newSelectPrompt("Select currency of conversion 'to'", currencyOptions(), currencies.GEL))
	case convertStepCurrencyTo:
		m.req.CurrencyTo = m.prompt.Value()
		m.step = convertStepConfirm

		prompt := newConfirmPrompt("Are your answers correct?")
		prompt.SetNote("Selecting 'No' restarts the converter form.")

		return m.setPrompt(prompt)
	case convertStepConfirm:
		if m.prompt.Value() == confirmYes {
			m.step = convertStepDone
			return tea.Quit
		}

		m.req = service.ConvertRequest{}
		m.step = convertStepYear

		return m.setPrompt(newSelectPrompt("Select year of conversion", yearOptions(), defaultYearValue()))
	default:
		return tea.Quit
	}
}

func (m *convertModel) setPrompt(p *promptModel) tea.Cmd {
	m.prompt = p

	return m.prompt.Init()
}

func renderConvertSummary(req service.ConvertRequest) string {
	var b strings.Builder

	if req.Year != "" && req.Month != "" && req.Day != "" {
		b.WriteString(fmt.Sprintf("Date: %s-%s-%s\n", req.Year, req.Month, req.Day))
	}

	if strings.TrimSpace(req.Amount) != "" {
		b.WriteString("Amount: ")
		b.WriteString(req.Amount)
		if req.CurrencyFrom != "" {
			b.WriteByte(' ')
			b.WriteString(req.CurrencyFrom)
		}
		b.WriteByte('\n')
	}

	if req.CurrencyFrom != "" && strings.TrimSpace(req.Amount) == "" {
		b.WriteString(fmt.Sprintf("Currency from: %s\n", req.CurrencyFrom))
	}

	if req.CurrencyTo != "" {
		b.WriteString(fmt.Sprintf("Currency to: %s\n", req.CurrencyTo))
	}

	return strings.TrimRight(b.String(), "\n")
}
