package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/savioxavier/termlink"
	"github.com/urfave/cli/v3"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/dateutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/moneyutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

func createLink(text, url string) {
	if !termlink.SupportsHyperlinks() {
		return
	}

	fmt.Println(termlink.Link(text, url))
}

func menuCalcTaxes(ctx context.Context, _ *cli.Command) error {
	createLink("Declarations", "https://decl.rs.ge/decls.aspx")

	var req service.CalculateRequest

	taxq, err := makeTaxTypeQuestion("tax_type", "Select your taxes type")
	if err != nil {
		return fmt.Errorf("failed to create tax type question: %w", err)
	}

	yincQ := makeMoneyAmountQuestion("year_income", "Income from the beginning of a calendar year (GEL)")

	if err = survey.Ask([]*survey.Question{taxq, yincQ}, &req); err != nil {
		return fmt.Errorf("failed to ask questions abot tax rate and year income: %w", err)
	}

	income, err := getIncomeRequest()
	if err != nil {
		return fmt.Errorf("failed to get income request: %w", err)
	}

	req.Income = income

	svc := service.New()

	resp, err := svc.Calculate(ctx, req)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(resp)
	fmt.Println()

	return nil
}

type incomeAnswers struct {
	service.Income
	AddMore bool `survey:"add_more"`
}

func getIncomeRequest() ([]service.Income, error) {
	var (
		isCorrect bool
		err       error
		income    []service.Income
	)

	for !isCorrect {
		income = make([]service.Income, 0) // reset slice

		answers := incomeAnswers{
			Income:  service.Income{},
			AddMore: true, // to start loop
		}

		for answers.AddMore {
			answers.DateRequest, err = getDateRequest()
			if err != nil {
				return nil, fmt.Errorf("failed to get date request: %w", err)
			}

			questions := []*survey.Question{
				makeMoneyAmountQuestion("amount", "Input amount of income"),
				makeCurrencyQuestion("currency", "Select currency of income"),
				makeConfirmQuestion("add_more", "Add more?"),
			}

			if err = survey.Ask(questions, &answers); err != nil {
				return nil, fmt.Errorf("failed to ask questions about income: %w", err)
			}

			income = append(income, answers.Income)
		}

		confirmQ := makeConfirmQuestion("confirm", "Are your answers correct?")

		if err = survey.Ask([]*survey.Question{confirmQ}, &isCorrect); err != nil {
			return nil, fmt.Errorf("failed to ask questions about confirmation: %w", err)
		}
	}

	return income, nil
}

func menuConvert(ctx context.Context, _ *cli.Command) error {
	type convertAnswers struct {
		service.ConvertRequest
		IsCorrect bool `survey:"confirm"`
	}

	var answers convertAnswers

	for !answers.IsCorrect {
		datereq, err := getDateRequest()
		if err != nil {
			return fmt.Errorf("failed to get date request: %w", err)
		}

		answers.DateRequest = datereq

		questions := []*survey.Question{
			makeMoneyAmountQuestion("amount", "Input amount to convert"),
			makeCurrencyQuestion("currency_from", "Select currency of conversion 'from'"),
			makeCurrencyQuestion("currency_to", "Select currency of conversion 'to'"),
		}

		questions = append(questions, makeConfirmQuestion("confirm", "Are your answers correct?"))
		if err = survey.Ask(questions, &answers); err != nil {
			return err
		}
	}

	resp, err := service.New().Convert(ctx, answers.ConvertRequest)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(resp)
	fmt.Println()

	return nil
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
				return fmt.Errorf("failed to parse answer to string: %w", err)
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
	}

	if err := survey.Ask(questions, &datereq); err != nil {
		return service.DateRequest{}, err
	}

	mq, err := makeMonthMenu(datereq)
	if err != nil {
		return service.DateRequest{}, err
	}

	questions = []*survey.Question{
		{
			Name:      "month",
			Prompt:    mq,
			Validate:  nil,
			Transform: nil,
		},
	}

	if err = survey.Ask(questions, &datereq); err != nil {
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

	if err = survey.Ask(questions, &datereq); err != nil {
		return service.DateRequest{}, err
	}

	return datereq, nil
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
				return fmt.Errorf("failed to parse option tax type: %w", err)
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

	qs := &survey.Select{
		Renderer:      survey.Renderer{},
		Message:       msg,
		Options:       titles,
		Default:       taxes.TaxTypeSmallBusiness.String(),
		Help:          "",
		PageSize:      0,
		VimMode:       false,
		FilterMessage: "",
		Filter:        nil,
		Description: func(_ string, index int) string {
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

func makeMonthMenu(p service.DateRequest) (survey.Prompt, error) {
	parseYear, err := dateutils.ParseYear(p.Year)
	if err != nil {
		return nil, fmt.Errorf("parse year: %w", err)
	}

	now := time.Now()

	months := dateutils.GetMonthsInYearTillDate(parseYear, now)

	msg := "Select month of income"

	var defval []string

	if now.Year() == parseYear {
		defval = append(defval, time.Now().Month().String())
	}

	return makeSurveySelect(msg, months, defval...), nil
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

	now := time.Now()

	days := dateutils.DaysList(dateutils.DaysInMonthTillDate(parseMonth, parseYear, now))

	msg := "Select day of income"

	var defval []string

	if now.Year() == parseYear && now.Month() == parseMonth {
		defval = append(defval, days[now.Day()-1])
	}

	return makeSurveySelect(msg, days, defval...), nil
}

func makeSurveySelect(msg string, items []string, defaultVal ...string) survey.Prompt {
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
