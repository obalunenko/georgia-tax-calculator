![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/obalunenko/georgia-tax-calculator)
[![GoDoc](https://godoc.org/github.com/obalunenko/georgia-tax-calculator?status.svg)](https://godoc.org/github.com/obalunenko/georgia-tax-calculator)
[![Latest release artifacts](https://img.shields.io/github/v/release/obalunenko/georgia-tax-calculator)](https://github.com/obalunenko/georgia-tax-calculator/releases/latest)
[![Go [lint, test]](https://github.com/obalunenko/georgia-tax-calculator/actions/workflows/go.yml/badge.svg)](https://github.com/obalunenko/georgia-tax-calculator/actions/workflows/go.yml)
[![Lint & Test & Build & Release](https://github.com/obalunenko/georgia-tax-calculator/actions/workflows/release.yml/badge.svg)](https://github.com/obalunenko/georgia-tax-calculator/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/obalunenko/georgia-tax-calculator)](https://goreportcard.com/report/github.com/obalunenko/georgia-tax-calculator)
[![codecov](https://codecov.io/gh/obalunenko/georgia-tax-calculator/branch/master/graph/badge.svg)](https://codecov.io/gh/obalunenko/georgia-tax-calculator)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=obalunenko_georgia-tax-calculator&metric=alert_status)](https://sonarcloud.io/summary/overall?id=obalunenko_georgia-tax-calculator)



# georgia-tax-calculator

Calculates income taxes in Georgia.

- Fetches official rates from the [nbg.gov.ge](https://nbg.gov.ge) for the date of income
- Converts income to GEL
- Calculate taxes amount according to specified Taxes Category.
