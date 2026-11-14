package main

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/dateutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/moneyutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

const (
	callbackPrefix = "cb:"
	confirmYes     = "yes"
	confirmNo      = "no"
)

// errNoKeyboard returned when the current step expects text instead of a button.
var errNoKeyboard = errors.New("step has no keyboard")

// keyboardOption is one inline keyboard button: the text shown and the value sent back.
type keyboardOption struct {
	text  string
	value string
}

func optionsFromValues(values []string) []keyboardOption {
	opts := make([]keyboardOption, len(values))
	for i, v := range values {
		opts[i] = keyboardOption{text: v, value: v}
	}

	return opts
}

// buildInlineKeyboard builds an inline keyboard for step.
// Items are laid out in rows of maxPerRow columns.
func buildInlineKeyboard(step string, opts []keyboardOption, maxPerRow int) telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	var row []telego.InlineKeyboardButton

	for _, opt := range opts {
		row = append(row, telego.InlineKeyboardButton{
			Text:         opt.text,
			CallbackData: callbackPrefix + step + ":" + opt.value,
		})

		if len(row) == maxPerRow {
			rows = append(rows, row)
			row = nil
		}
	}

	if len(row) > 0 {
		rows = append(rows, row)
	}

	return telego.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// parseCallbackData splits callback data built by buildInlineKeyboard into step and value.
func parseCallbackData(data string) (step, value string, ok bool) {
	rest, ok := strings.CutPrefix(data, callbackPrefix)
	if !ok {
		return "", "", false
	}

	return strings.Cut(rest, ":")
}

// callbackValue returns the value of callback data when it was sent by the keyboard of the current step
// and is one of the options this step offers now. Buttons of old keyboards are rejected this way.
func callbackValue(sess *session, data string) (value string, ok bool, err error) {
	step, value, ok := parseCallbackData(data)
	if !ok || step == stepWithoutKeyboard || step != sess.stepID() {
		return "", false, nil
	}

	opts, _, err := stepOptions(sess)
	if err != nil {
		return "", false, err
	}

	ok = slices.ContainsFunc(opts, func(o keyboardOption) bool {
		return o.value == value
	})

	return value, ok, nil
}

// stepKeyboard builds the keyboard of the current step.
func stepKeyboard(sess *session) (*telego.InlineKeyboardMarkup, error) {
	opts, maxPerRow, err := stepOptions(sess)
	if err != nil {
		return nil, err
	}

	kb := buildInlineKeyboard(sess.stepID(), opts, maxPerRow)

	return &kb, nil
}

// stepOptions returns the buttons of the current step and how many fit in one row.
func stepOptions(sess *session) ([]keyboardOption, int, error) {
	const (
		oneColumn      = 1
		confirmPerRow  = 2
		defaultPerRow  = 4
		weekDaysPerRow = 7
	)

	switch sess.stepID() {
	case stepCalcTaxType:
		opts, err := taxTypeOptions()

		return opts, oneColumn, err
	case stepCalcYear:
		// One calculation covers one calendar year, fixed by the first income.
		if len(sess.incomes) > 0 {
			return optionsFromValues([]string{sess.incomes[0].Year}), defaultPerRow, nil
		}

		return yearOptions(), defaultPerRow, nil
	case stepConvertYear:
		return yearOptions(), defaultPerRow, nil
	case stepCalcMonth:
		opts, err := monthOptions(sess.currentInc.Year)

		return opts, defaultPerRow, err
	case stepConvertMonth:
		opts, err := monthOptions(sess.convertReq.Year)

		return opts, defaultPerRow, err
	case stepCalcDay:
		opts, err := dayOptions(sess.currentInc.Year, sess.currentInc.Month)

		return opts, weekDaysPerRow, err
	case stepConvertDay:
		opts, err := dayOptions(sess.convertReq.Year, sess.convertReq.Month)

		return opts, weekDaysPerRow, err
	case stepCalcCurrency, stepConvertFrom, stepConvertTo:
		return currencyOptions(), defaultPerRow, nil
	case stepCalcAddMore, stepCalcConfirm, stepConvertConfirm:
		return confirmOptions(), confirmPerRow, nil
	default:
		return nil, 0, errNoKeyboard
	}
}

// confirmOptions returns Yes/No buttons.
func confirmOptions() []keyboardOption {
	return []keyboardOption{
		{text: "✅ Yes", value: confirmYes},
		{text: "❌ No", value: confirmNo},
	}
}

// taxTypeOptions returns tax type buttons; the text shows the rate, the value is the tax type name.
func taxTypeOptions() ([]keyboardOption, error) {
	rates, err := taxes.AllTaxRates()
	if err != nil {
		return nil, fmt.Errorf("get tax rates: %w", err)
	}

	sort.Slice(rates, func(i, j int) bool {
		return rates[i].Type > rates[j].Type
	})

	const toPercentage float64 = 100

	opts := make([]keyboardOption, len(rates))
	for i, r := range rates {
		pct := moneyutils.Multiply(r.Rate, toPercentage)
		opts[i] = keyboardOption{
			text:  fmt.Sprintf("%s (%s%%)", r.Type.String(), moneyutils.ToString(pct)),
			value: r.Type.String(),
		}
	}

	return opts, nil
}

// yearOptions returns years from the current one down to 2016.
func yearOptions() []keyboardOption {
	now := service.Now()
	var years []string

	for i := now.Year(); i >= service.FirstYear; i-- {
		years = append(years, strconv.Itoa(i))
	}

	return optionsFromValues(years)
}

// monthOptions returns months of the given year up to today.
func monthOptions(year string) ([]keyboardOption, error) {
	y, err := dateutils.ParseYear(year)
	if err != nil {
		return nil, fmt.Errorf("parse year: %w", err)
	}

	return optionsFromValues(dateutils.GetMonthsInYearTillDate(y, service.Now())), nil
}

// dayOptions returns days of the given year and month up to today.
func dayOptions(year, month string) ([]keyboardOption, error) {
	y, err := dateutils.ParseYear(year)
	if err != nil {
		return nil, fmt.Errorf("parse year: %w", err)
	}

	m, err := dateutils.ParseMonth(month)
	if err != nil {
		return nil, fmt.Errorf("parse month: %w", err)
	}

	return optionsFromValues(dateutils.DaysList(dateutils.DaysInMonthTillDate(m, y, service.Now()))), nil
}

// currencyOptions returns all supported currencies, sorted.
func currencyOptions() []keyboardOption {
	currs := currencies.All()
	sort.Strings(currs)

	return optionsFromValues(currs)
}
