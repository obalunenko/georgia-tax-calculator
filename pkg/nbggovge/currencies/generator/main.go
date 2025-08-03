// generator is a tool to generate currencies.go file.
package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"os"
	"slices"
	"time"

	log "github.com/obalunenko/logger"
	"github.com/obalunenko/version"
	"golang.org/x/tools/imports"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

const (
	templateName  = "codes.go.tmpl"
	generatorName = "currencies-generator"
)

//go:embed codes.go.tmpl
var templateFS embed.FS

type currencyInfo struct {
	Code string
	Name string
}

type versionInfo struct {
	Version string
	Name    string
	Go      string
}

type templateParams struct {
	VersionInfo  versionInfo
	PackageName  string
	CurrencyList []currencyInfo
	Date         string
	Source       string
}

func main() {
	ctx := context.Background()

	input := flag.String("input", "", "input file")
	useClient := flag.Bool("client", false, "use client")
	output := flag.String("output", "codes.go", "output file")
	pkg := flag.String("pkg", "currencies", "package name")
	flag.Parse()

	if *output == "" {
		log.Fatal(ctx, "Output file is required")
	}

	if *pkg == "" {
		log.Fatal(ctx, "Package name is required")
	}

	fileData, err := getInput(ctx, *useClient, *input)
	if err != nil {
		log.WithError(ctx, err).Fatal("Failed to get input")
	}

	currencies := make([]currencyInfo, 0, len(fileData[0].Currencies))

	for _, currency := range fileData[0].Currencies {
		currencies = append(currencies, currencyInfo{
			Code: currency.Code,
			Name: currency.Name,
		})
	}

	currencies = mutateCurrencyList(currencies)

	t, err := template.ParseFS(templateFS, templateName)
	if err != nil {
		log.WithError(ctx, err).Fatal("Failed to parse template")
	}

	tf, err := time.Parse("2006-01-02T15:04:05.000Z", fileData[0].Date)
	if err != nil {
		log.WithError(ctx, err).Fatal("Failed to parse date")
	}

	params := templateParams{
		VersionInfo: versionInfo{
			Version: version.GetVersion(),
			Name:    generatorName,
			Go:      version.GetGoVersion(),
		},
		PackageName:  *pkg,
		CurrencyList: currencies,
		Date:         tf.Format(time.DateOnly),
	}

	if *useClient {
		params.Source = "client"
	} else {
		params.Source = "file"
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, params); err != nil {
		log.WithError(ctx, err).Fatal("Failed to execute template")
	}

	formattedSrc, err := imports.Process(*output, buf.Bytes(), nil)
	if err != nil {
		log.WithError(ctx, err).Fatal("Failed to format source code")
	}

	const perm = 0o600

	if err = os.WriteFile(*output, formattedSrc, perm); err != nil {
		log.WithError(ctx, err).Fatal("Failed to write source code")
	}

	log.WithFields(ctx, log.Fields{
		"output": *output,
		"input":  *input,
		"pkg":    *pkg,
		"len":    len(currencies),
		"source": params.Source,
		"date":   params.Date,
	}).Info("Done")
}

func getInput(ctx context.Context, useClient bool, input string) ([]nbggovge.Rates, error) {
	if useClient {
		rates, err := nbggovge.New().Rates(ctx, option.WithDate(time.Now()))

		if err != nil || len(rates.Currencies) == 0 {
			return nil, fmt.Errorf("get rates from client: %w", err)
		}

		return []nbggovge.Rates{rates}, nil
	}

	if input == "" {
		return nil, errors.New("input file is required")
	}

	jsonData, err := os.ReadFile(input) //nolint:gosec // we don't care about file inclusion here.
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	var fileData []nbggovge.Rates
	if err = json.Unmarshal(jsonData, &fileData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	if len(fileData) != 1 {
		return nil, errors.New("invalid input file")
	}

	return fileData, nil
}

func mutateCurrencyList(currencies []currencyInfo) []currencyInfo {
	newV := slices.Clone(currencies)

	if !slices.ContainsFunc(newV, func(info currencyInfo) bool {
		return info.Code == "GEL"
	}) {
		newV = append(newV, currencyInfo{
			Code: "GEL",
			Name: "Georgian Lari",
		})
	}

	slices.SortFunc(newV, func(a, b currencyInfo) int {
		if a.Code == b.Code {
			return 0
		}

		if a.Code < b.Code {
			return -1
		}

		return 1
	})

	return newV
}
