package main

import (
	"context"
	"fmt"

	"github.com/savioxavier/termlink"
	"github.com/urfave/cli/v3"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
	"github.com/obalunenko/georgia-tax-calculator/internal/spinner"
)

func createLink(text, url string) {
	if !termlink.SupportsHyperlinks() {
		return
	}

	fmt.Println(termlink.Link(text, url))
}

func menuCalcTaxes(ctx context.Context, _ *cli.Command) error {
	createLink("Declarations", service.DeclarationsURL)

	req, err := runTaxDetailsMenu()
	if err != nil {
		return fmt.Errorf("failed to collect tax details: %w", err)
	}

	incomes, err := runIncomeMenu()
	if err != nil {
		return fmt.Errorf("failed to collect incomes: %w", err)
	}

	req.Income = incomes

	svc := service.New()

	stop := spinner.Start(
		fmt.Sprintf("Calculating taxes for %d incomes", len(req.Income)),
		fmt.Sprintf("Calculated taxes for %d incomes", len(req.Income)),
	)

	resp, err := svc.Calculate(ctx, req)

	stop()

	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(resp)
	fmt.Println()

	return nil
}

func menuConvert(ctx context.Context, _ *cli.Command) error {
	req, err := runConvertMenu()
	if err != nil {
		return fmt.Errorf("failed to collect converter input: %w", err)
	}

	stop := spinner.Start(
		fmt.Sprintf("Converting %s%s to %s", req.Amount, req.CurrencyFrom, req.CurrencyTo),
		fmt.Sprintf("Converted %s%s to %s", req.Amount, req.CurrencyFrom, req.CurrencyTo),
	)

	resp, err := service.New().Convert(ctx, req)

	stop()

	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(resp)
	fmt.Println()

	return nil
}
