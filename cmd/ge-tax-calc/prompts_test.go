package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testHeight = 24
	testTitle  = "title"
)

type staticModel struct {
	content string
}

func (m staticModel) Init() tea.Cmd {
	return nil
}

func (m staticModel) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m staticModel) View() tea.View {
	return tea.NewView(m.content)
}

func enterKey() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyEnter}
}

func downKey() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyDown}
}

func upKey() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyUp}
}

func ctrlCKey() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
}

func windowSize(height int) tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: 100, Height: height}
}

// typeKeys returns the key presses for typing the text.
func typeKeys(text string) []tea.Msg {
	msgs := make([]tea.Msg, 0, len(text))

	for _, r := range text {
		msgs = append(msgs, tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	return msgs
}

// send delivers the messages to the model and returns the last command.
func send(m tea.Model, msgs ...tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	for _, msg := range msgs {
		_, cmd = m.Update(msg)
	}

	return cmd
}

func assertQuit(t *testing.T, cmd tea.Cmd) {
	t.Helper()

	require.NotNil(t, cmd)
	assert.IsType(t, tea.QuitMsg{}, cmd())
}

func numberedOptions(n int) []option {
	opts := make([]option, n)

	for i := range opts {
		label := fmt.Sprintf("opt%02d", i)
		opts[i] = option{Label: label, Value: label}
	}

	return opts
}

func newAmountPrompt() *promptModel {
	return newInputPrompt(testTitle, "0.00", "", validateMoneyInput)
}

func TestFitHeightModelView(t *testing.T) {
	const content = "one\ntwo\nthree\n"

	tests := []struct {
		name   string
		height int
		want   string
	}{
		{name: "terminal height is unknown", height: 0, want: content},
		{name: "frame fits", height: 4, want: content},
		{name: "top lines are dropped", height: 2, want: "three\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &fitHeightModel{Model: staticModel{content: content}}

			send(m, windowSize(tt.height))

			assert.Equal(t, tt.want, m.View().Content)
		})
	}
}

func TestNewProgram(t *testing.T) {
	assert.NotNil(t, newProgram(staticModel{}))
}

func TestPromptVisibleOptions(t *testing.T) {
	tests := []struct {
		name      string
		height    int
		pos       int
		note      string
		wantStart int
		wantEnd   int
	}{
		{name: "height is not known yet", height: 0, pos: 5, wantStart: 0, wantEnd: 10},
		{name: "list fits", height: 15, pos: 5, wantStart: 0, wantEnd: 10},
		{name: "cursor at the top", height: 11, pos: 0, wantStart: 0, wantEnd: 6},
		{name: "cursor in the middle", height: 11, pos: 5, wantStart: 2, wantEnd: 8},
		{name: "cursor at the bottom", height: 11, pos: 9, wantStart: 4, wantEnd: 10},
		{name: "note takes rows", height: 11, pos: 0, note: "note", wantStart: 0, wantEnd: 4},
		{name: "tiny terminal keeps the cursor row", height: 3, pos: 7, wantStart: 7, wantEnd: 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newSelectPrompt(testTitle, numberedOptions(10), "")
			p.SetNote(tt.note)
			p.cursor = tt.pos
			p.Update(windowSize(tt.height))

			start, end := p.visibleOptions()

			assert.Equal(t, tt.wantStart, start)
			assert.Equal(t, tt.wantEnd, end)
		})
	}
}

func TestPromptViewKeepsCursorVisible(t *testing.T) {
	const (
		minHeight = 6
		maxHeight = 70
	)

	for _, total := range []int{1, 2, 12, 31, 44} {
		opts := numberedOptions(total)

		for height := minHeight; height <= maxHeight; height++ {
			for pos := range total {
				p := newSelectPrompt(testTitle, opts, "")
				p.cursor = pos
				p.Update(windowSize(height))

				frame := p.View()
				msg := fmt.Sprintf("total=%d height=%d cursor=%d", total, height, pos)

				require.LessOrEqual(t, strings.Count(frame, "\n")+1, height, msg)
				require.Contains(t, frame, "> "+opts[pos].Label+"\n", msg)
			}
		}
	}
}

func TestPromptViewReachesEveryOption(t *testing.T) {
	opts := currencyOptions()

	p := newSelectPrompt(testTitle, opts, "")
	p.Update(windowSize(testHeight))

	seen := make(map[string]bool, len(opts))

	for range opts {
		for line := range strings.SplitSeq(p.View(), "\n") {
			seen[strings.TrimSpace(strings.TrimPrefix(line, ">"))] = true
		}

		p.Update(downKey())
	}

	for i := range opts {
		assert.True(t, seen[opts[i].Label], opts[i].Label)
	}
}

func TestPromptInstructions(t *testing.T) {
	const selectText = "Use ↑/↓ to navigate, Enter to confirm."

	opts := numberedOptions(10)

	cut := newSelectPrompt(testTitle, opts, "opt03")
	cut.Update(windowSize(11))

	assert.Equal(t, selectText, newSelectPrompt(testTitle, opts, "").instructions())
	assert.Equal(t, selectText+" (4/10)", cut.instructions())
	assert.Equal(t, "Enter to confirm.", newAmountPrompt().instructions())
}

func TestSelectPromptNavigation(t *testing.T) {
	tests := []struct {
		name string
		msgs []tea.Msg
		want int
	}{
		{name: "down", msgs: []tea.Msg{downKey()}, want: 2},
		{name: "stops at the end", msgs: []tea.Msg{downKey(), downKey(), downKey()}, want: 2},
		{name: "up", msgs: []tea.Msg{upKey()}, want: 0},
		{name: "stops at the start", msgs: []tea.Msg{upKey(), upKey(), upKey()}, want: 0},
		{name: "j moves down", msgs: typeKeys("j"), want: 2},
		{name: "k moves up", msgs: typeKeys("k"), want: 0},
		{name: "other keys are ignored", msgs: typeKeys("x"), want: 1},
		{name: "other messages are ignored", msgs: []tea.Msg{windowSize(testHeight)}, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newSelectPrompt(testTitle, numberedOptions(3), "opt01")

			for _, msg := range tt.msgs {
				p.Update(msg)
			}

			assert.Equal(t, tt.want, p.cursor)
			assert.False(t, p.Completed())
		})
	}
}

func TestSelectPromptCompletes(t *testing.T) {
	p := newSelectPrompt(testTitle, numberedOptions(3), "opt02")

	p.Update(enterKey())

	assert.True(t, p.Completed())
	assert.Equal(t, "opt02", p.Value())
}

func TestSelectPromptWithoutOptions(t *testing.T) {
	p := newSelectPrompt(testTitle, nil, "")

	p.Update(enterKey())

	assert.False(t, p.Completed())
	assert.Contains(t, p.View(), "No options available")
	assert.Contains(t, p.View(), "Error: no options available")
}

func TestConfirmPrompt(t *testing.T) {
	tests := []struct {
		name string
		msgs []tea.Msg
		want string
	}{
		{name: "yes by default", msgs: []tea.Msg{enterKey()}, want: confirmYes},
		{name: "no", msgs: []tea.Msg{downKey(), enterKey()}, want: confirmNo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newConfirmPrompt(testTitle)

			for _, msg := range tt.msgs {
				p.Update(msg)
			}

			assert.Equal(t, tt.want, p.Value())
		})
	}
}

func TestInputPromptShowsWholePlaceholder(t *testing.T) {
	assert.Contains(t, newAmountPrompt().View(), ".00")
}

func TestInputPromptCompletes(t *testing.T) {
	p := newAmountPrompt()

	for _, msg := range typeKeys("12.5") {
		p.Update(msg)
	}

	p.Update(enterKey())

	assert.True(t, p.Completed())
	assert.Equal(t, "12.5", p.Value())
}

func TestInputPromptRejectsInvalidValue(t *testing.T) {
	p := newAmountPrompt()

	p.Update(enterKey())

	assert.False(t, p.Completed())
	assert.Contains(t, p.View(), "Error: value is required")
}

func TestInputPromptKeepsErrorUntilInput(t *testing.T) {
	tests := []struct {
		name      string
		msg       tea.Msg
		wantErr   bool
		wantValue string
	}{
		{name: "initial blink", msg: textinput.Blink(), wantErr: true},
		{name: "cursor blink", msg: cursor.BlinkMsg{}, wantErr: true},
		{name: "window resize", msg: windowSize(testHeight), wantErr: true},
		{name: "typing", msg: typeKeys("1")[0], wantValue: "1"},
		{name: "paste", msg: tea.PasteMsg{Content: "12"}, wantValue: "12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newAmountPrompt()

			p.Update(enterKey())
			require.Error(t, p.err)

			p.Update(tt.msg)

			assert.Equal(t, tt.wantErr, p.err != nil)
			assert.Equal(t, tt.wantValue, p.textInput.Value())
		})
	}
}

func TestPromptInheritHeight(t *testing.T) {
	prev := newSelectPrompt(testTitle, nil, "")
	prev.Update(windowSize(testHeight))

	next := newSelectPrompt(testTitle, nil, "")

	next.inheritHeight(prev)
	assert.Equal(t, testHeight, next.height)

	next.inheritHeight(nil)
	assert.Equal(t, testHeight, next.height)

	var missing *promptModel

	assert.NotPanics(t, func() { missing.inheritHeight(prev) })
}

func TestModelsAbortOnCtrlC(t *testing.T) {
	convert := newConvertModel()
	details := newTaxDetailsModel()
	income := newIncomeModel()

	tests := []struct {
		name  string
		model tea.Model
		err   func() error
	}{
		{name: "convert", model: convert, err: func() error { return convert.err }},
		{name: "tax details", model: details, err: func() error { return details.err }},
		{name: "income", model: income, err: func() error { return income.err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.model.Init()

			assertQuit(t, send(tt.model, ctrlCKey()))
			assert.ErrorIs(t, tt.err(), errUserAborted)
		})
	}
}

func TestModelsStopOnError(t *testing.T) {
	errBoom := errors.New("boom")

	tests := []struct {
		name  string
		model tea.Model
	}{
		{name: "convert form", model: &convertModel{err: errBoom}},
		{name: "tax details form", model: &taxDetailsModel{err: errBoom}},
		{name: "income form", model: &incomeModel{err: errBoom}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertQuit(t, tt.model.Init())
			assertQuit(t, send(tt.model, enterKey()))
			assert.Equal(t, "error: boom\n", tt.model.View().Content)
		})
	}
}
