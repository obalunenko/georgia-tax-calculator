package main

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

func TestParseCallbackData(t *testing.T) {
	tests := []struct {
		data      string
		wantStep  string
		wantValue string
		wantOK    bool
	}{
		{data: "cb:calc_day:31", wantStep: stepCalcDay, wantValue: "31", wantOK: true},
		{data: "cb:calc_tax:Small Business", wantStep: stepCalcTaxType, wantValue: "Small Business", wantOK: true},
		{data: "cb:31", wantOK: false},
		{data: "31", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.data, func(t *testing.T) {
			step, value, ok := parseCallbackData(tt.data)

			assert.Equal(t, tt.wantOK, ok)

			if tt.wantOK {
				assert.Equal(t, tt.wantStep, step)
				assert.Equal(t, tt.wantValue, value)
			}
		})
	}
}

func calcSession(step calcStep, inc service.Income) *session {
	return &session{flow: flowCalculate, calcStep: step, currentInc: inc}
}

func convertSession(step convertStep, req service.ConvertRequest) *session {
	return &session{flow: flowConvert, convertStep: step, convertReq: req}
}

func TestCallbackValue(t *testing.T) {
	april2024 := service.Income{DateRequest: service.DateRequest{Year: "2024", Month: "April"}}

	tests := []struct {
		name   string
		sess   *session
		data   string
		wantOK bool
	}{
		{
			name:   "year of the current step",
			sess:   convertSession(convertStepYear, service.ConvertRequest{}),
			data:   "cb:conv_year:2024",
			wantOK: true,
		},
		{
			name:   "year before 2016",
			sess:   convertSession(convertStepYear, service.ConvertRequest{}),
			data:   "cb:conv_year:2015",
			wantOK: false,
		},
		{
			name:   "day that exists in the month",
			sess:   calcSession(calcStepDay, april2024),
			data:   "cb:calc_day:30",
			wantOK: true,
		},
		{
			name:   "day 31 from an older keyboard of the same step",
			sess:   calcSession(calcStepDay, april2024),
			data:   "cb:calc_day:31",
			wantOK: false,
		},
		{
			name:   "button of an earlier step",
			sess:   calcSession(calcStepCurrency, april2024),
			data:   "cb:calc_year:2024",
			wantOK: false,
		},
		{
			name:   "button of another flow",
			sess:   convertSession(convertStepConfirm, service.ConvertRequest{}),
			data:   "cb:calc_ok:yes",
			wantOK: false,
		},
		{
			name:   "confirmation",
			sess:   calcSession(calcStepConfirm, service.Income{}),
			data:   "cb:calc_ok:yes",
			wantOK: true,
		},
		{
			name:   "currency at the confirmation step",
			sess:   calcSession(calcStepConfirm, service.Income{}),
			data:   "cb:calc_ok:" + currencies.USD,
			wantOK: false,
		},
		{
			name:   "tax type name",
			sess:   calcSession(calcStepTaxType, service.Income{}),
			data:   "cb:calc_tax:" + taxes.TaxTypeSmallBusiness.String(),
			wantOK: true,
		},
		{
			name:   "tax type button text instead of value",
			sess:   calcSession(calcStepTaxType, service.Income{}),
			data:   "cb:calc_tax:Small Business (1%)",
			wantOK: false,
		},
		{
			name:   "unknown currency",
			sess:   convertSession(convertStepCurrencyFrom, service.ConvertRequest{}),
			data:   "cb:conv_from:XXX",
			wantOK: false,
		},
		{
			name:   "data in the format used before step IDs",
			sess:   convertSession(convertStepCurrencyFrom, service.ConvertRequest{}),
			data:   "cb:" + currencies.USD,
			wantOK: false,
		},
		{
			name:   "button while the step expects text",
			sess:   calcSession(calcStepAmount, april2024),
			data:   "cb:calc_day:30",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok, err := callbackValue(tt.sess, tt.data)
			require.NoError(t, err)

			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestCallbackValueRejectsFutureMonth(t *testing.T) {
	now := time.Now()
	if now.Month() == time.December {
		t.Skip("no future month in the current year")
	}

	sess := convertSession(convertStepMonth, service.ConvertRequest{
		DateRequest: service.DateRequest{Year: strconv.Itoa(now.Year())},
	})

	_, ok, err := callbackValue(sess, "cb:conv_month:"+time.December.String())
	require.NoError(t, err)

	assert.False(t, ok)
}

func TestStepKeyboard(t *testing.T) {
	inc := service.Income{DateRequest: service.DateRequest{Year: "2024", Month: "February"}}
	req := service.ConvertRequest{DateRequest: inc.DateRequest}

	sessions := []*session{
		calcSession(calcStepTaxType, inc),
		calcSession(calcStepYear, inc),
		calcSession(calcStepMonth, inc),
		calcSession(calcStepDay, inc),
		calcSession(calcStepCurrency, inc),
		calcSession(calcStepAddMore, inc),
		calcSession(calcStepConfirm, inc),
		convertSession(convertStepYear, req),
		convertSession(convertStepMonth, req),
		convertSession(convertStepDay, req),
		convertSession(convertStepCurrencyFrom, req),
		convertSession(convertStepCurrencyTo, req),
		convertSession(convertStepConfirm, req),
	}

	// Telegram limits callback_data to 64 bytes.
	const maxCallbackData = 64

	for _, sess := range sessions {
		t.Run(sess.stepID(), func(t *testing.T) {
			kb, err := stepKeyboard(sess)
			require.NoError(t, err)
			require.NotEmpty(t, kb.InlineKeyboard)

			for _, row := range kb.InlineKeyboard {
				for _, btn := range row {
					assert.LessOrEqual(t, len(btn.CallbackData), maxCallbackData, btn.CallbackData)

					_, ok, err := callbackValue(sess, btn.CallbackData)
					require.NoError(t, err)
					assert.True(t, ok, "every button of the current keyboard is accepted: %s", btn.CallbackData)
				}
			}
		})
	}
}

func TestStepKeyboardForTextStep(t *testing.T) {
	_, err := stepKeyboard(calcSession(calcStepYearIncome, service.Income{}))

	assert.ErrorIs(t, err, errNoKeyboard)
}

func TestTaxTypeOptions(t *testing.T) {
	opts, err := taxTypeOptions()
	require.NoError(t, err)

	assert.Contains(t, opts, keyboardOption{
		text:  "Small Business (1%)",
		value: taxes.TaxTypeSmallBusiness.String(),
	})
}

func TestDayOptions(t *testing.T) {
	opts, err := dayOptions("2024", "February")
	require.NoError(t, err)

	require.Len(t, opts, 29)
	assert.Equal(t, keyboardOption{text: "29", value: "29"}, opts[28])
}
