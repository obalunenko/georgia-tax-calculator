package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/obalunenko/georgia-tax-calculator/internal/spinner"

	"github.com/obalunenko/georgia-tax-calculator/internal/moneyutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/dateutils"

	"github.com/urfave/cli/v2"

	log "github.com/obalunenko/logger"

	"github.com/obalunenko/georgia-tax-calculator/internal/converter"
	"github.com/obalunenko/georgia-tax-calculator/internal/models"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

func main() {
	ctx := context.Background()

	ctx = log.ContextWithLogger(ctx, log.FromContext(ctx))

	app := cli.NewApp()
	app.Name = "ge-tax-calc"
	app.Description = "Helper tool for preparing tax declarations in Georgia." +
		"It get income amount in received currency, converts it to GEL according to" +
		"official rates on date of income and calculates tax amount" +
		"according to selected ta category."
	app.Usage = `A command line tool helper for preparing tax declaration in Georgia `
	app.Authors = []*cli.Author{
		{
			Name:  "Oleg Balunenko",
			Email: "oleg.balunenko@gmail.com",
		},
	}
	app.CommandNotFound = notFound(ctx)
	app.Commands = commands(ctx)
	app.Version = printVersion(ctx)
	app.Before = printHeader(ctx)
	app.After = onExit(ctx)

	if err := app.Run(os.Args); err != nil {

		log.WithError(ctx, err).Fatal("Run failed")
	}
}

func calc(ctx context.Context, p inputParams) (string, error) {
	stop := spinner.Start()
	defer stop()

	year, err := dateutils.ParseYear(p.Year)
	if err != nil {
		return "", err
	}

	month, err := dateutils.ParseMonth(p.Month)
	if err != nil {
		return "", err
	}

	day, err := dateutils.ParseDay(p.Day)
	if err != nil {
		return "", err
	}

	date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)

	mv, err := moneyutils.Parse(p.Amount)
	if err != nil {
		return "", err
	}

	income := models.NewMoney(mv, p.Currency)
	incomeOut := models.NewResultOutput("Income", income)

	converted, err := convert(ctx, convertParams{
		date:  date,
		m:     income,
		tocur: currencies.GEL,
	})
	if err != nil {
		return "", fmt.Errorf("failed to convert: %w", err)
	}

	convertedOut := models.NewResultOutput("Converted", converted.Money)

	tt, err := taxes.ParseTaxType(p.Taxtype)
	if err != nil {
		return "", fmt.Errorf("failed to parse tax type: %w", err)
	}

	tax, err := taxes.Calc(converted.Money, tt)
	if err != nil {
		return "", fmt.Errorf("failed to calc taxes: %w", err)
	}

	taxesOut := models.NewResultOutput("Taxes", tax.Money)

	const (
		layout = "2006-01-02"
	)

	var resp string
	resp += fmt.Sprintf("Date: %s\n", date.Format(layout))
	resp += fmt.Sprintf("Tax Rate: %s\n", tax.Rate.String())
	resp += fmt.Sprintf("%s\n", incomeOut.String())
	resp += fmt.Sprintf("%s\n", convertedOut.String())
	resp += fmt.Sprintf("%s\n", taxesOut.String())

	return resp, nil
}

type convertParams struct {
	date  time.Time
	m     models.Money
	tocur string
}

func convert(ctx context.Context, p convertParams) (converter.Response, error) {
	client := nbggovge.New()

	c := converter.NewConverter(client)

	resp, err := c.Convert(ctx, p.m, p.tocur, p.date)
	if err != nil {
		return converter.Response{}, err
	}

	return resp, nil
}
