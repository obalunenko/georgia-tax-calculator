package main

import (
	"context"
	"fmt"

	"github.com/savioxavier/termlink"
	"github.com/urfave/cli/v3"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
)

func createLink(text, url string) {
	if !termlink.SupportsHyperlinks() {
		return
	}

	fmt.Println(termlink.Link(text, url))
}

func menuCalcTaxes(ctx context.Context, _ *cli.Command) error {
	createLink("Declarations", "https://decl.rs.ge/decls.aspx")

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

	resp, err := svc.Calculate(ctx, req)
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

	resp, err := service.New().Convert(ctx, req)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(resp)
	fmt.Println()

	return nil
}
