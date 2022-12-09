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

func menuConvert(ctx context.Context) cli.ActionFunc {
	return func(c *cli.Context) error {
		type convertAnswers struct {
			service.ConvertRequest
			IsCorrect bool `survey:"confirm"`
		}

		var (
			answers convertAnswers
		)

		for !answers.IsCorrect {
			datereq, err := getDateRequest()
			if err != nil {
				return err
			}

			answers.DateRequest = datereq

			questions := []*survey.Question{
				makeMoneyAmountQuestion("amount", "Input amount to convert"),
				makeCurrencyQuestion("currency_from", "Select currency of conversion 'from'"),
				makeCurrencyQuestion("currency_to", "Select currency of conversion 'to'"),
			}

			questions = append(questions, makeConfirmQuestion("confirm", "Are your answers correct?"))
			if err := survey.Ask(questions, &answers); err != nil {
				return err
			}
		}

		resp, err := service.New().Convert(ctx, answers.ConvertRequest)
		if err != nil {
			return err
		}

		fmt.Println()
		fmt.Println(resp)

		return nil
	}
}

func makeMoneyAmountQuestion(fieldname, msg string) *survey.Question {
	return &survey.Question{
		Name: fieldname,
		Prompt: &survey.Input{
			Renderer: survey.Renderer{},
			Message:  msg,
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
	}
}

func makeCurrencyQuestion(fieldname, msg string) *survey.Question {
	return &survey.Question{
		Name:      fieldname,
		Prompt:    makeCurrencyMenu(msg),
		Validate:  nil,
		Transform: nil,
	}
}

func makeConfirmQuestion(fieldname, msg string) *survey.Question {
	return &survey.Question{
		Name: fieldname,
		Prompt: &survey.Confirm{

			Renderer: survey.Renderer{},
			Message:  msg,
			Default:  true,
			Help:     "",
		},
		Validate:  nil,
		Transform: nil,
	}
}

func getDateRequest() (service.DateRequest, error) {
	var datereq service.DateRequest

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

	if err := survey.Ask(questions, &datereq); err != nil {
		return service.DateRequest{}, err
	}

	dq, err := makeDayMenu(datereq)
	if err != nil {
		return service.DateRequest{}, err
	}

	questions = []*survey.Question{
		{
			Name:      "day",
			Prompt:    dq,
			Validate:  nil,
			Transform: nil,
		},
	}

	if err := survey.Ask(questions, &datereq); err != nil {
		return service.DateRequest{}, err
	}

	return datereq, nil
}

func menuCalcTaxes(ctx context.Context) cli.ActionFunc {
	return func(c *cli.Context) error {
		type calculateAnswers struct {
			service.CalculateRequest
			IsCorrect bool `survey:"confirm"`
		}

		var answers calculateAnswers

		taxq, err := makeTaxTypeQuestion("tax_type", "Select your taxes type")
		if err != nil {
			return err
		}

		for !answers.IsCorrect {
			answers.DateRequest, err = getDateRequest()
			if err != nil {
				return err
			}

			questions := []*survey.Question{
				makeMoneyAmountQuestion("amount", "Input amount of income"),
				makeCurrencyQuestion("currency", "Select currency of income"),
			}

			questions = append(questions, taxq, makeConfirmQuestion("confirm", "Are your answers correct?"))

			if err = survey.Ask(questions, &answers); err != nil {
				return err
			}
		}

		svc := service.New()

		resp, err := svc.Calculate(ctx, answers.CalculateRequest)
		if err != nil {
			return err
		}

		fmt.Println()
		fmt.Println(resp)

		return nil
	}
}

func makeTaxTypeQuestion(fieldname, msg string) (*survey.Question, error) {
	taxMenu, err := makeTaxTypeMenu(msg)
	if err != nil {
		return nil, err
	}

	return &survey.Question{
		Name:   fieldname,
		Prompt: taxMenu,
		Validate: func(ans interface{}) error {
			s, ok := ans.(core.OptionAnswer)
			if !ok {
				return fmt.Errorf("failed to cast answer to OptionAnswer: [%T], %v", ans, ans)
			}

			_, err = taxes.ParseTaxType(s.Value)
			if err != nil {
				return err
			}
			return nil
		},
		Transform: nil,
	}, nil
}

func makeTaxTypeMenu(msg string) (survey.Prompt, error) {
	rates, err := taxes.AllTaxRates()
	if err != nil {
		return nil, err
	}

	sort.Slice(rates, func(i, j int) bool {
		return rates[i].Type > rates[j].Type
	})

	titles := make([]string, len(rates))
	for i, r := range rates {
		titles[i] = r.Type.String()
	}

	var qs = &survey.Select{
		Renderer:      survey.Renderer{},
		Message:       msg,
		Options:       titles,
		Default:       taxes.TaxTypeSmallBusiness.String(),
		Help:          "",
		PageSize:      0,
		VimMode:       false,
		FilterMessage: "",
		Filter:        nil,
		Description: func(value string, index int) string {
			const toPercentage float64 = 100

			m := moneyutils.Multiply(rates[index].Rate, toPercentage)

			return fmt.Sprintf("%s %%", moneyutils.ToString(m))
		},
	}

	return qs, nil
}

func makeCurrencyMenu(msg string) survey.Prompt {
	currs := currencies.All()

	sort.Strings(currs)

	return makeSurveySelect(msg, currs, currencies.EUR)
}

func makeYearsMenu() survey.Prompt {
	years := getYears(time.Now())

	sort.Slice(years, func(i, j int) bool {
		return years[i] > years[j]
	})

	msg := "Select year of income"

	return makeSurveySelect(msg, years, strconv.Itoa(time.Now().Year()))
}

func makeMonthMenu() survey.Prompt {
	months := dateutils.GetMonths()

	msg := "Select month of income"

	return makeSurveySelect(msg, months, time.Now().Month().String())
}

func makeDayMenu(p service.DateRequest) (survey.Prompt, error) {
	parseMonth, err := dateutils.ParseMonth(p.Month)
	if err != nil {
		return nil, fmt.Errorf("parse month: %w", err)
	}

	parseYear, err := dateutils.ParseYear(p.Year)
	if err != nil {
		return nil, fmt.Errorf("parse year: %w", err)
	}

	days := dateutils.DaysList(dateutils.DaysInMonth(parseMonth, parseYear))

	msg := "Select day of income"

	return makeSurveySelect(msg, days), nil
}

func makeSurveySelect(msg string, items []string, defaultVal ...any) survey.Prompt {
	var defval any

	if len(defaultVal) == 1 {
		defval = defaultVal[0]
	}

	return &survey.Select{
		Renderer:      survey.Renderer{},
		Message:       msg,
		Options:       items,
		Default:       defval,
		Help:          "",
		PageSize:      len(items),
		VimMode:       false,
		FilterMessage: "",
		Filter:        nil,
		Description:   nil,
	}
}

func getYears(now time.Time) []string {
	var years []string

	const begin = 2016

	for i := begin; i <= now.Year(); i++ {
		years = append(years, strconv.Itoa(i))
	}

	return years
}
