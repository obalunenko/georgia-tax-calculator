package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/dateutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/moneyutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

var errUserAborted = errors.New("input aborted")

const (
	confirmYes = "yes"
	confirmNo  = "no"

	keyCtrlC = "ctrl+c"

	// inputWidth has to be set: with zero width textinput v2 renders only
	// the first rune of the placeholder.
	inputWidth = 32
)

// fitHeightModel drops the top lines of a frame that is taller than the
// terminal. Bubble Tea v2 redraws such inline frames on every renderer tick,
// flooding the terminal scrollback.
type fitHeightModel struct {
	tea.Model
	height int
}

func newProgram(model tea.Model) *tea.Program {
	return tea.NewProgram(&fitHeightModel{Model: model})
}

func (m *fitHeightModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.height = size.Height
	}

	var cmd tea.Cmd

	m.Model, cmd = m.Model.Update(msg)

	return m, cmd
}

func (m *fitHeightModel) View() tea.View {
	view := m.Model.View()

	lines := strings.Split(view.Content, "\n")
	if m.height > 0 && len(lines) > m.height {
		view.SetContent(strings.Join(lines[len(lines)-m.height:], "\n"))
	}

	return view
}

type option struct {
	Label       string
	Value       string
	Description string
}

type promptMode int

const (
	promptModeInput promptMode = iota
	promptModeSelect
)

type promptModel struct {
	title     string
	note      string
	mode      promptMode
	options   []option
	cursor    int
	textInput textinput.Model
	validator func(string) error
	value     string
	done      bool
	err       error
	height    int
}

func newInputPrompt(title, placeholder, defaultValue string, validator func(string) error) *promptModel {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = placeholder
	ti.SetWidth(inputWidth)
	ti.SetValue(defaultValue)
	ti.Focus()

	return &promptModel{
		title:     title,
		mode:      promptModeInput,
		textInput: ti,
		validator: validator,
	}
}

func newSelectPrompt(title string, opts []option, defaultValue string) *promptModel {
	cursor := 0
	if defaultValue != "" {
		for i := range opts {
			if opts[i].Value == defaultValue {
				cursor = i
				break
			}
		}
	}

	return &promptModel{
		title:   title,
		mode:    promptModeSelect,
		options: opts,
		cursor:  cursor,
	}
}

func newConfirmPrompt(title string) *promptModel {
	opts := []option{
		{Label: "Yes", Value: confirmYes},
		{Label: "No", Value: confirmNo},
	}

	return newSelectPrompt(title, opts, confirmYes)
}

func (p *promptModel) Init() tea.Cmd {
	if p == nil {
		return nil
	}

	if p.mode == promptModeInput {
		return textinput.Blink
	}

	return nil
}

func (p *promptModel) Update(msg tea.Msg) tea.Cmd {
	if p == nil || p.done {
		return nil
	}

	if size, ok := msg.(tea.WindowSizeMsg); ok {
		p.height = size.Height
	}

	switch p.mode {
	case promptModeInput:
		return p.updateInput(msg)
	case promptModeSelect:
		p.updateSelect(msg)
	}

	return nil
}

func (p *promptModel) updateInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.Code == tea.KeyEnter {
			p.err = nil
			val := strings.TrimSpace(p.textInput.Value())
			if p.validator != nil {
				if err := p.validator(val); err != nil {
					p.err = err
					return cmd
				}
			}

			p.value = val
			p.done = true
			return cmd
		}

		p.err = nil
	case tea.PasteMsg:
		p.err = nil
	case tea.WindowSizeMsg:
		// No-op.
	}

	p.textInput, cmd = p.textInput.Update(msg)

	return cmd
}

func (p *promptModel) updateSelect(msg tea.Msg) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return
	}

	switch key.String() {
	case "up", "k":
		if p.cursor > 0 {
			p.cursor--
		}
	case "down", "j":
		if p.cursor < len(p.options)-1 {
			p.cursor++
		}
	case "enter":
		if len(p.options) == 0 {
			p.err = errors.New("no options available")
			return
		}

		p.value = p.options[p.cursor].Value
		p.done = true
	}
}

func (p *promptModel) View() string {
	if p == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(p.title)
	b.WriteString("\n\n")

	switch p.mode {
	case promptModeInput:
		b.WriteString(p.textInput.View())
		b.WriteString("\n")
	case promptModeSelect:
		if len(p.options) == 0 {
			b.WriteString("No options available\n")
		} else {
			start, end := p.visibleOptions()

			for i := start; i < end; i++ {
				cursor := " "
				if i == p.cursor {
					cursor = ">"
				}

				b.WriteString(cursor)
				b.WriteByte(' ')
				b.WriteString(p.options[i].Label)

				if desc := strings.TrimSpace(p.options[i].Description); desc != "" {
					b.WriteString(" — ")
					b.WriteString(desc)
				}

				b.WriteByte('\n')
			}
		}
	}

	b.WriteString(p.notesView())

	b.WriteByte('\n')
	b.WriteString(p.instructions())
	b.WriteByte('\n')

	return b.String()
}

func (p *promptModel) notesView() string {
	var b strings.Builder

	if note := strings.TrimSpace(p.note); note != "" {
		b.WriteByte('\n')
		b.WriteString(note)
		b.WriteByte('\n')
	}

	if p.err != nil {
		b.WriteString("\nError: ")
		b.WriteString(p.err.Error())
		b.WriteByte('\n')
	}

	return b.String()
}

// visibleOptions returns the range of options that fits the terminal and keeps
// the cursor in view.
func (p *promptModel) visibleOptions() (start, end int) {
	// Title, instructions, the blank lines after and before them and the
	// final empty line of the frame.
	const (
		chromeLines = 5
		half        = 2
	)

	total := len(p.options)
	rows := p.height - chromeLines - strings.Count(p.title+p.notesView(), "\n")

	if p.height <= 0 || rows >= total {
		return 0, total
	}

	rows = max(rows, 1)
	start = min(max(p.cursor-rows/half, 0), total-rows)

	return start, start + rows
}

func (p *promptModel) instructions() string {
	switch p.mode {
	case promptModeInput:
		return "Enter to confirm."
	case promptModeSelect:
		const text = "Use ↑/↓ to navigate, Enter to confirm."

		if start, end := p.visibleOptions(); end-start < len(p.options) {
			return fmt.Sprintf("%s (%d/%d)", text, p.cursor+1, len(p.options))
		}

		return text
	default:
		return ""
	}
}

func (p *promptModel) Completed() bool {
	if p == nil {
		return false
	}

	return p.done
}

func (p *promptModel) Value() string {
	if p == nil {
		return ""
	}

	return p.value
}

func (p *promptModel) SetNote(note string) {
	if p == nil {
		return
	}

	p.note = note
}

// inheritHeight carries over the terminal height: the window size message is
// sent only on start and resize, not for every new prompt.
func (p *promptModel) inheritHeight(prev *promptModel) {
	if p == nil || prev == nil {
		return
	}

	p.height = prev.height
}

func validateMoneyInput(val string) error {
	if strings.TrimSpace(val) == "" {
		return errors.New("value is required")
	}

	if _, err := moneyutils.Parse(val); err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	return nil
}

func taxTypeOptions() ([]option, error) {
	rates, err := taxes.AllTaxRates()
	if err != nil {
		return nil, err
	}

	sort.Slice(rates, func(i, j int) bool {
		return rates[i].Type > rates[j].Type
	})

	const toPercentage float64 = 100

	opts := make([]option, len(rates))
	for i := range rates {
		m := moneyutils.Multiply(rates[i].Rate, toPercentage)
		opts[i] = option{
			Label:       rates[i].Type.String(),
			Value:       rates[i].Type.String(),
			Description: fmt.Sprintf("%s %%", moneyutils.ToString(m)),
		}
	}

	return opts, nil
}

func currencyOptions() []option {
	currs := currencies.All()

	sort.Strings(currs)

	return valuesToOptions(currs)
}

func yearOptions() []option {
	years := getYears(time.Now())

	sort.Slice(years, func(i, j int) bool {
		return years[i] > years[j]
	})

	return valuesToOptions(years)
}

func monthOptions(year string) ([]option, error) {
	parseYear, err := dateutils.ParseYear(year)
	if err != nil {
		return nil, fmt.Errorf("parse year: %w", err)
	}

	months := dateutils.GetMonthsInYearTillDate(parseYear, time.Now())

	return valuesToOptions(months), nil
}

func dayOptions(year, month string) ([]option, error) {
	parseYear, err := dateutils.ParseYear(year)
	if err != nil {
		return nil, fmt.Errorf("parse year: %w", err)
	}

	parseMonth, err := dateutils.ParseMonth(month)
	if err != nil {
		return nil, fmt.Errorf("parse month: %w", err)
	}

	days := dateutils.DaysList(dateutils.DaysInMonthTillDate(parseMonth, parseYear, time.Now()))

	return valuesToOptions(days), nil
}

func valuesToOptions(values []string) []option {
	opts := make([]option, len(values))
	for i := range values {
		opts[i] = option{
			Label: values[i],
			Value: values[i],
		}
	}

	return opts
}

func defaultYearValue() string {
	return strconv.Itoa(time.Now().Year())
}

func defaultMonthValue(year string) string {
	now := time.Now()
	if strconv.Itoa(now.Year()) == year {
		return now.Month().String()
	}

	return ""
}

func defaultDayValue(year, month string) string {
	now := time.Now()
	if strconv.Itoa(now.Year()) != year {
		return ""
	}

	monthParsed := now.Month().String()
	if !strings.EqualFold(monthParsed, month) {
		return ""
	}

	days := dateutils.DaysList(dateutils.DaysInMonthTillDate(now.Month(), now.Year(), now))

	day := now.Day()
	if day-1 < len(days) {
		return days[day-1]
	}

	return ""
}

func getYears(now time.Time) []string {
	var years []string

	const begin = 2016

	for i := begin; i <= now.Year(); i++ {
		years = append(years, strconv.Itoa(i))
	}

	return years
}
