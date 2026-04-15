package main

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/mymmrac/telego"

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

// buildInlineKeyboard builds an inline keyboard with a list of string options.
// Items are laid out in rows of maxPerRow columns.
func buildInlineKeyboard(items []string, maxPerRow int) telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	var row []telego.InlineKeyboardButton

	for _, item := range items {
		row = append(row, telego.InlineKeyboardButton{
			Text:         item,
			CallbackData: callbackPrefix + item,
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

// buildConfirmKeyboard builds a Yes/No inline keyboard.
func buildConfirmKeyboard() telego.InlineKeyboardMarkup {
	return telego.InlineKeyboardMarkup{
		InlineKeyboard: [][]telego.InlineKeyboardButton{
			{
				{Text: "✅ Yes", CallbackData: callbackPrefix + confirmYes},
				{Text: "❌ No", CallbackData: callbackPrefix + confirmNo},
			},
		},
	}
}

// taxTypeKeyboard builds the keyboard for tax type selection.
func taxTypeKeyboard() (telego.InlineKeyboardMarkup, error) {
	rates, err := taxes.AllTaxRates()
	if err != nil {
		return telego.InlineKeyboardMarkup{}, fmt.Errorf("get tax rates: %w", err)
	}

	sort.Slice(rates, func(i, j int) bool {
		return rates[i].Type > rates[j].Type
	})

	const toPercentage float64 = 100

	items := make([]string, len(rates))
	for i, r := range rates {
		pct := moneyutils.Multiply(r.Rate, toPercentage)
		items[i] = fmt.Sprintf("%s (%s%%)", r.Type.String(), moneyutils.ToString(pct))
	}

	return buildInlineKeyboard(items, 1), nil
}

// taxTypeFromCallback extracts the tax type string from an item string like "Small Business (1%)".
func taxTypeFromItem(item string) string {
	// The item is in the format "TaxType (X%)" - we need just the type name.
	for _, tt := range taxes.AllTaxTypes() {
		if len(item) >= len(tt.String()) && item[:len(tt.String())] == tt.String() {
			return tt.String()
		}
	}

	return item
}

// yearKeyboard builds the year selection keyboard.
func yearKeyboard() telego.InlineKeyboardMarkup {
	const (
		begin     = 2016
		maxPerRow = 4
	)

	now := time.Now()
	var years []string

	for i := now.Year(); i >= begin; i-- {
		years = append(years, strconv.Itoa(i))
	}

	return buildInlineKeyboard(years, maxPerRow)
}

// monthKeyboard builds the month selection keyboard for the given year.
func monthKeyboard(year string) (telego.InlineKeyboardMarkup, error) {
	y, err := dateutils.ParseYear(year)
	if err != nil {
		return telego.InlineKeyboardMarkup{}, fmt.Errorf("parse year: %w", err)
	}

	months := dateutils.GetMonthsInYearTillDate(y, time.Now())

	const maxPerRow = 4

	return buildInlineKeyboard(months, maxPerRow), nil
}

// dayKeyboard builds the day selection keyboard for the given year and month.
func dayKeyboard(year, month string) (telego.InlineKeyboardMarkup, error) {
	y, err := dateutils.ParseYear(year)
	if err != nil {
		return telego.InlineKeyboardMarkup{}, fmt.Errorf("parse year: %w", err)
	}

	m, err := dateutils.ParseMonth(month)
	if err != nil {
		return telego.InlineKeyboardMarkup{}, fmt.Errorf("parse month: %w", err)
	}

	days := dateutils.DaysList(dateutils.DaysInMonthTillDate(m, y, time.Now()))

	const maxPerRow = 7

	return buildInlineKeyboard(days, maxPerRow), nil
}

// currencyKeyboard builds the currency selection keyboard.
func currencyKeyboard() telego.InlineKeyboardMarkup {
	currs := currencies.All()
	sort.Strings(currs)

	const maxPerRow = 4

	return buildInlineKeyboard(currs, maxPerRow)
}
