package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

func runTaxDetailsMenu() (service.CalculateRequest, error) {
	model := newTaxDetailsModel()
	if _, err := tea.NewProgram(model).Run(); err != nil {
		return service.CalculateRequest{}, err
	}

	if model.err != nil {
		return service.CalculateRequest{}, model.err
	}

	return service.CalculateRequest{
		TaxType:    model.taxType,
		YearIncome: model.yearIncome,
	}, nil
}

func runIncomeMenu() ([]service.Income, error) {
	model := newIncomeModel()
	if _, err := tea.NewProgram(model).Run(); err != nil {
		return nil, err
	}

	if model.err != nil {
		return nil, model.err
	}

	return model.incomes, nil
}

type taxDetailsStep int

const (
	taxStepTaxType taxDetailsStep = iota
	taxStepYearIncome
	taxStepDone
)

type taxDetailsModel struct {
	step       taxDetailsStep
	prompt     *promptModel
	taxType    string
	yearIncome string
	err        error
}

func newTaxDetailsModel() *taxDetailsModel {
	opts, err := taxTypeOptions()
	if err != nil {
		return &taxDetailsModel{err: err}
	}

	prompt := newSelectPrompt("Select your taxes type", opts, taxes.TaxTypeSmallBusiness.String())
	prompt.SetNote("Bubble Tea UI • press Enter to confirm your selection.")

	return &taxDetailsModel{
		step:   taxStepTaxType,
		prompt: prompt,
	}
}

func (m *taxDetailsModel) Init() tea.Cmd {
	if m.err != nil {
		return tea.Quit
	}

	return m.prompt.Init()
}

func (m *taxDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *taxDetailsModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("error: %v\n", m.err)
	}

	var b strings.Builder

	if m.step == taxStepYearIncome {
		b.WriteString(fmt.Sprintf("Selected tax type: %s\n\n", m.taxType))
	}

	b.WriteString(m.prompt.View())

	return b.String()
}

func (m *taxDetailsModel) advance() tea.Cmd {
	switch m.step {
	case taxStepTaxType:
		m.taxType = m.prompt.Value()
		m.step = taxStepYearIncome

		prompt := newInputPrompt(
			"Income from the beginning of a calendar year (GEL)",
			"0.00",
			"",
			validateMoneyInput,
		)

		return m.setPrompt(prompt)
	case taxStepYearIncome:
		m.yearIncome = m.prompt.Value()
		m.step = taxStepDone

		return tea.Quit
	default:
		return tea.Quit
	}
}

func (m *taxDetailsModel) setPrompt(p *promptModel) tea.Cmd {
	m.prompt = p

	return m.prompt.Init()
}

type incomeStep int

const (
	incomeStepYear incomeStep = iota
	incomeStepMonth
	incomeStepDay
	incomeStepAmount
	incomeStepCurrency
	incomeStepAddMore
	incomeStepConfirm
	incomeStepDone
)

type incomeModel struct {
	step    incomeStep
	prompt  *promptModel
	incomes []service.Income
	current service.Income
	err     error
}

func newIncomeModel() *incomeModel {
	m := &incomeModel{
		incomes: make([]service.Income, 0),
	}

	return m
}

func (m *incomeModel) Init() tea.Cmd {
	if m.err != nil {
		return tea.Quit
	}

	if m.prompt == nil {
		return m.setPrompt(newSelectPrompt("Select year of income", yearOptions(), defaultYearValue()))
	}

	return m.prompt.Init()
}

func (m *incomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *incomeModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("error: %v\n", m.err)
	}

	var b strings.Builder

	b.WriteString(renderIncomeList(m.incomes))

	if m.step == incomeStepConfirm {
		b.WriteString("\nReview your answers:\n\n")
		b.WriteString(renderIncomeDetails(m.incomes))
		b.WriteByte('\n')
	}

	if m.prompt != nil {
		b.WriteString(m.prompt.View())
	}

	return b.String()
}

func (m *incomeModel) advance() tea.Cmd {
	switch m.step {
	case incomeStepYear:
		m.current.Year = m.prompt.Value()
		m.step = incomeStepMonth

		opts, err := monthOptions(m.current.Year)
		if err != nil {
			m.err = err
			return tea.Quit
		}

		return m.setPrompt(newSelectPrompt("Select month of income", opts, defaultMonthValue(m.current.Year)))
	case incomeStepMonth:
		m.current.Month = m.prompt.Value()
		m.step = incomeStepDay

		opts, err := dayOptions(m.current.Year, m.current.Month)
		if err != nil {
			m.err = err
			return tea.Quit
		}

		return m.setPrompt(newSelectPrompt("Select day of income", opts, defaultDayValue(m.current.Year, m.current.Month)))
	case incomeStepDay:
		m.current.Day = m.prompt.Value()
		m.step = incomeStepAmount

		return m.setPrompt(newInputPrompt("Input amount of income", "0.00", "", validateMoneyInput))
	case incomeStepAmount:
		m.current.Amount = m.prompt.Value()
		m.step = incomeStepCurrency

		return m.setPrompt(newSelectPrompt("Select currency of income", currencyOptions(), currencies.USD))
	case incomeStepCurrency:
		m.current.Currency = m.prompt.Value()
		m.incomes = append(m.incomes, m.current)
		m.current = service.Income{}
		m.step = incomeStepAddMore

		prompt := newConfirmPrompt("Add another income entry?")
		prompt.SetNote("Choose 'No' when you are done adding incomes.")

		return m.setPrompt(prompt)
	case incomeStepAddMore:
		if m.prompt.Value() == confirmYes {
			m.step = incomeStepYear

			return m.setPrompt(newSelectPrompt("Select year of income", yearOptions(), defaultYearValue()))
		}

		m.step = incomeStepConfirm

		prompt := newConfirmPrompt("Are your answers correct?")
		prompt.SetNote("Selecting 'No' will restart income entry.")

		return m.setPrompt(prompt)
	case incomeStepConfirm:
		if m.prompt.Value() == confirmYes {
			m.step = incomeStepDone
			return tea.Quit
		}

		m.incomes = make([]service.Income, 0)
		m.current = service.Income{}
		m.step = incomeStepYear

		return m.setPrompt(newSelectPrompt("Select year of income", yearOptions(), defaultYearValue()))
	default:
		return tea.Quit
	}
}

func (m *incomeModel) setPrompt(p *promptModel) tea.Cmd {
	m.prompt = p

	return m.prompt.Init()
}

func renderIncomeList(incomes []service.Income) string {
	var b strings.Builder

	b.WriteString("Captured incomes:\n")

	if len(incomes) == 0 {
		b.WriteString("  none yet\n\n")
		return b.String()
	}

	for i := range incomes {
		date := fmt.Sprintf("%s-%s-%s", incomes[i].Year, incomes[i].Month, incomes[i].Day)
		b.WriteString(fmt.Sprintf("  %d) %s — %s %s\n", i+1, date, incomes[i].Amount, incomes[i].Currency))
	}

	b.WriteByte('\n')

	return b.String()
}

func renderIncomeDetails(incomes []service.Income) string {
	if len(incomes) == 0 {
		return "No income entries captured."
	}

	var b strings.Builder

	for i := range incomes {
		b.WriteString(fmt.Sprintf("%d)\n", i+1))
		b.WriteString(fmt.Sprintf("   Date: %s-%s-%s\n", incomes[i].Year, incomes[i].Month, incomes[i].Day))
		b.WriteString(fmt.Sprintf("   Amount: %s %s\n", incomes[i].Amount, incomes[i].Currency))
		b.WriteByte('\n')
	}

	return b.String()
}
