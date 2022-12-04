package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/urfave/cli/v2"

	"github.com/obalunenko/georgia-tax-calculator/internal/moneyutils"
	"github.com/obalunenko/georgia-tax-calculator/internal/service"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/dateutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

func menu(ctx context.Context) cli.ActionFunc {
	return func(c *cli.Context) error {
		var (
			answers service.InputParams
		)

		isCorrect := []*survey.Question{
			{
				Name: "",
				Prompt: &survey.Confirm{

					Renderer: survey.Renderer{},
					Message:  "Are your answers correct?",
					Default:  true,
					Help:     "",
				},
				Validate:  nil,
				Transform: nil,
			},
		}

		var correct bool

		for !correct {
			questions := []*survey.Question{
				{
					Name:      "year",
					Prompt:    makeYearsMenu(),
					Validate:  nil,
					Transform: nil,
				},
				{
					Name:      "month",
					Prompt:    makeMonthMenu(),
					Validate:  nil,
					Transform: nil,
				},
			}

			if err := survey.Ask(questions, &answers); err != nil {
				return err
			}

			dq, err := makeDayMenu(answers)
			if err != nil {
				return err
			}

			questions = []*survey.Question{
				{
					Name:      "day",
					Prompt:    dq,
					Validate:  nil,
					Transform: nil,
				},
			}

			if err := survey.Ask(questions, &answers); err != nil {
				return err
			}

			taxMenu, err := makeTaxTypeMenu()
			if err != nil {
				return err
			}

			questions = []*survey.Question{
				{
					Name: "amount",
					Prompt: &survey.Input{
						Renderer: survey.Renderer{},
						Message:  "Input amount of income",
						Default:  "0",
						Help:     "",
						Suggest:  nil,
					},
					Validate: func(ans interface{}) error {
						s, ok := ans.(string)
						if !ok {
							return fmt.Errorf("failed to cast answer to string: [%T], %v", ans, ans)
						}

						_, err := moneyutils.Parse(s)
						if err != nil {
							return err
						}
						return nil
					},
					Transform: nil,
				},
				{
					Name:      "currency",
					Prompt:    makeCurrencyMenu(),
					Validate:  nil,
					Transform: nil,
				},
				{
					Name:   "tax_type",
					Prompt: taxMenu,
					Validate: func(ans interface{}) error {
						s, ok := ans.(core.OptionAnswer)
						if !ok {
							return fmt.Errorf("failed to cast answer to OptionAnswer: [%T], %v", ans, ans)
						}

						_, err := taxes.ParseTaxType(s.Value)
						if err != nil {
							return err
						}
						return nil
					},
					Transform: nil,
				},
			}

			if err := survey.Ask(questions, &answers); err != nil {
				return err
			}

			if err := survey.Ask(isCorrect, &correct); err != nil {
				return err
			}
		}

		svc := service.New()

		resp, err := svc.Calculate(ctx, answers)
		if err != nil {
			return err
		}

		fmt.Println()
		fmt.Println(resp)

		return nil
	}
}

func makeTaxTypeMenu() (survey.Prompt, error) {
	rates, err := taxes.AllTaxRates()
	if err != nil {
		return nil, err
	}

	sort.Slice(rates, func(i, j int) bool {
		return rates[i].Type > rates[i].Type
	})

	titles := make([]string, len(rates))
	for i, r := range rates {
		titles[i] = r.Type.String()
	}

	var qs = &survey.Select{
		Renderer:      survey.Renderer{},
		Message:       "Choose a tax type:",
		Options:       titles,
		Default:       nil,
		Help:          "",
		PageSize:      0,
		VimMode:       false,
		FilterMessage: "",
		Filter:        nil,
		Description: func(value string, index int) string {
			m := moneyutils.Multiply(rates[index].Rate, 100)

			return fmt.Sprintf("%s %%", moneyutils.ToString(m))
		},
	}

	return qs, nil
}
func makeCurrencyMenu() survey.Prompt {
	currs := []string{currencies.EUR, currencies.USD, currencies.GBP, currencies.BYN, currencies.GEL}

	items := makeMenuItemsList(currs)

	msg := "Select currency of income"

	return makeSurveySelect(msg, items)
}

func makeYearsMenu() survey.Prompt {
	years := getYears(time.Now())

	items := makeMenuItemsList(years)

	msg := "Select year of income"
	return makeSurveySelect(msg, items)
}

func makeMonthMenu() survey.Prompt {
	months := dateutils.GetMonths()

	items := makeMenuItemsList(months)

	msg := "Select month of income"

	return makeSurveySelect(msg, items)
}

func makeDayMenu(p service.InputParams) (survey.Prompt, error) {
	parseMonth, err := dateutils.ParseMonth(p.Month)
	if err != nil {
		return nil, fmt.Errorf("parse month: %w", err)
	}

	parseYear, err := dateutils.ParseYear(p.Year)
	if err != nil {
		return nil, fmt.Errorf("parse year: %w", err)
	}

	days := dateutils.DaysList(dateutils.DaysInMonth(parseMonth, parseYear))

	items := makeMenuItemsList(days)

	msg := "Select day of income"

	return makeSurveySelect(msg, items), nil
}

func makeSurveySelect(msg string, items []string) survey.Prompt {
	return &survey.Select{
		Renderer:      survey.Renderer{},
		Message:       msg,
		Options:       items,
		Default:       nil,
		Help:          "",
		PageSize:      len(items),
		VimMode:       false,
		FilterMessage: "",
		Filter:        nil,
		Description:   nil,
	}
}

func makeMenuItemsList(list []string, commands ...string) []string {
	items := make([]string, 0, len(list)+len(commands))

	items = append(items, list...)

	items = append(items, commands...)

	return items
}

// years
func getYears(now time.Time) []string {
	var years []string

	const begin = 2016

	for i := begin; i <= now.Year(); i++ {
		years = append(years, strconv.Itoa(i))
	}

	return years
}
