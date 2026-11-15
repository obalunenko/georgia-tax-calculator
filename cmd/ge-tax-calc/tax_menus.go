package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

func runTaxDetailsMenu() (service.CalculateRequest, error) {
	model := newTaxDetailsModel()
	if _, err := newProgram(model).Run(); err != nil {
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

func runIncomeMenu(taxType string) ([]service.Income, error) {
	model := newIncomeModel(taxType)
	if _, err := newProgram(model).Run(); err != nil {
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

	if key, ok := msg.(tea.KeyPressMsg); ok {
		if key.String() == keyCtrlC {
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

func (m *taxDetailsModel) View() tea.View {
	if m.err != nil {
		return tea.NewView(fmt.Sprintf("error: %v\n", m.err))
	}

	var b strings.Builder

	if m.step == taxStepYearIncome {
		b.WriteString(fmt.Sprintf("Selected tax type: %s\n\n", m.taxType))
	}

	b.WriteString(m.prompt.View())

	return tea.NewView(b.String())
}

func (m *taxDetailsModel) advance() tea.Cmd {
	switch m.step {
	case taxStepTaxType:
		m.taxType = m.prompt.Value()
		m.step = taxStepYearIncome

		return m.setPrompt(yearIncomePrompt(m.taxType))
	case taxStepYearIncome:
		m.yearIncome = m.prompt.Value()
		m.step = taxStepDone

		return tea.Quit
	default:
		return tea.Quit
	}
}

// yearIncomePrompt asks for the income since the start of the year. The Small Business categories enter
// field (15) of the previous month's declaration, so this month's incomes are not counted twice.
func yearIncomePrompt(taxType string) *promptModel {
	tt, err := taxes.ParseTaxType(taxType)
	if err != nil || !tt.IsSmallBusiness() {
		return newInputPrompt("Income from the beginning of a calendar year (GEL)", "0.00", "", validateYearIncomeInput)
	}

	prompt := newInputPrompt(
		"Field (15) of your previous month's declaration: income since the start of the year (GEL)",
		"0.00",
		"",
		validateYearIncomeInput,
	)
	prompt.SetNote("Enter 0 for your first declaration of the year.")

	return prompt
}

func (m *taxDetailsModel) setPrompt(p *promptModel) tea.Cmd {
	p.inheritHeight(m.prompt)
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
	incomeStepType
	incomeStepAddMore
	incomeStepConfirm
	incomeStepDone
)

type incomeModel struct {
	step    incomeStep
	prompt  *promptModel
	incomes []service.Income
	current service.Income
	// smallBusiness is set for the categories declared on the Small Business monthly declaration:
	// they ask for the income type of each income and keep all incomes in one month.
	smallBusiness bool
	err           error
}

func newIncomeModel(taxType string) *incomeModel {
	m := &incomeModel{
		incomes: make([]service.Income, 0),
	}

	tt, err := taxes.ParseTaxType(taxType)
	if err != nil {
		m.err = err

		return m
	}

	m.smallBusiness = tt.IsSmallBusiness()

	return m
}

func (m *incomeModel) Init() tea.Cmd {
	if m.err != nil {
		return tea.Quit
	}

	if m.prompt == nil {
		return m.setPrompt(m.yearPrompt())
	}

	return m.prompt.Init()
}

func (m *incomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		return m, tea.Quit
	}

	if key, ok := msg.(tea.KeyPressMsg); ok {
		if key.String() == keyCtrlC {
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

func (m *incomeModel) View() tea.View {
	if m.err != nil {
		return tea.NewView(fmt.Sprintf("error: %v\n", m.err))
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

	return tea.NewView(b.String())
}

// yearPrompt asks for the income year. After the first income only its year is offered,
// because one calculation covers one calendar year.
func (m *incomeModel) yearPrompt() *promptModel {
	if len(m.incomes) > 0 {
		year := m.incomes[0].Year

		return newSelectPrompt("Select year of income", valuesToOptions([]string{year}), year)
	}

	return newSelectPrompt("Select year of income", yearOptions(), defaultYearValue())
}

// monthPrompt asks for the income month. In the Small Business categories only the first income's month is
// offered after the first income, because the Small Business declaration is monthly.
func (m *incomeModel) monthPrompt() (*promptModel, error) {
	if m.smallBusiness && len(m.incomes) > 0 {
		month := m.incomes[0].Month

		return newSelectPrompt("Select month of income", valuesToOptions([]string{month}), month), nil
	}

	opts, err := monthOptions(m.current.Year)
	if err != nil {
		return nil, err
	}

	return newSelectPrompt("Select month of income", opts, defaultMonthValue(m.current.Year)), nil
}

func (m *incomeModel) advance() tea.Cmd {
	switch m.step {
	case incomeStepYear:
		m.current.Year = m.prompt.Value()
		m.step = incomeStepMonth

		prompt, err := m.monthPrompt()
		if err != nil {
			m.err = err
			return tea.Quit
		}

		return m.setPrompt(prompt)
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

		if m.smallBusiness {
			m.step = incomeStepType

			return m.setPrompt(m.incomeTypePrompt())
		}

		return m.addCurrent()
	case incomeStepType:
		m.current.Type = m.prompt.Value()

		return m.addCurrent()
	case incomeStepAddMore:
		if m.prompt.Value() == confirmYes {
			m.step = incomeStepYear

			return m.setPrompt(m.yearPrompt())
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

		return m.setPrompt(m.yearPrompt())
	default:
		return tea.Quit
	}
}

// incomeTypePrompt asks for the income type; the previous income's type is preselected.
func (m *incomeModel) incomeTypePrompt() *promptModel {
	var previous string
	if n := len(m.incomes); n > 0 {
		previous = m.incomes[n-1].Type
	}

	return newSelectPrompt("Select income type (declaration field)", incomeTypeOptions(), previous)
}

// addCurrent captures the current income and asks whether to add another one.
func (m *incomeModel) addCurrent() tea.Cmd {
	m.incomes = append(m.incomes, m.current)
	m.current = service.Income{}
	m.step = incomeStepAddMore

	prompt := newConfirmPrompt("Add another income entry?")
	prompt.SetNote("Choose 'No' when you are done adding incomes.")

	return m.setPrompt(prompt)
}

func (m *incomeModel) setPrompt(p *promptModel) tea.Cmd {
	p.inheritHeight(m.prompt)
	m.prompt = p

	return m.prompt.Init()
}

func formatIncomeDate(income service.Income) string {
	return fmt.Sprintf("%s-%s-%s", income.Year, income.Month, income.Day)
}

func renderIncomeList(incomes []service.Income) string {
	var b strings.Builder

	b.WriteString("Captured incomes:\n")

	if len(incomes) == 0 {
		b.WriteString("  none yet\n\n")
		return b.String()
	}

	for i := range incomes {
		date := formatIncomeDate(incomes[i])
		fmt.Fprintf(&b, "  %d) %s — %s %s", i+1, date, incomes[i].Amount, incomes[i].Currency)

		if label := taxes.IncomeTypeLabel(incomes[i].Type); label != "" {
			b.WriteString(" — " + label)
		}

		b.WriteByte('\n')
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
		b.WriteString(fmt.Sprintf("   Date: %s\n", formatIncomeDate(incomes[i])))
		b.WriteString(fmt.Sprintf("   Amount: %s %s\n", incomes[i].Amount, incomes[i].Currency))

		if label := taxes.IncomeTypeLabel(incomes[i].Type); label != "" {
			fmt.Fprintf(&b, "   Type: %s\n", label)
		}

		b.WriteByte('\n')
	}

	return b.String()
}
