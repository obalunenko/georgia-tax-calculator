![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/obalunenko/georgia-tax-calculator)
[![GoDoc](https://godoc.org/github.com/obalunenko/georgia-tax-calculator?status.svg)](https://godoc.org/github.com/obalunenko/georgia-tax-calculator)
[![Latest release artifacts](https://img.shields.io/github/v/release/obalunenko/georgia-tax-calculator)](https://github.com/obalunenko/georgia-tax-calculator/releases/latest)
[![Go [lint, test]](https://github.com/obalunenko/georgia-tax-calculator/actions/workflows/go.yml/badge.svg)](https://github.com/obalunenko/georgia-tax-calculator/actions/workflows/go.yml)
[![Lint & Test & Build & Release](https://github.com/obalunenko/georgia-tax-calculator/actions/workflows/release.yml/badge.svg)](https://github.com/obalunenko/georgia-tax-calculator/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/obalunenko/georgia-tax-calculator)](https://goreportcard.com/report/github.com/obalunenko/georgia-tax-calculator)
[![codecov](https://codecov.io/gh/obalunenko/georgia-tax-calculator/branch/master/graph/badge.svg)](https://codecov.io/gh/obalunenko/georgia-tax-calculator)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=obalunenko_georgia-tax-calculator&metric=alert_status)](https://sonarcloud.io/summary/overall?id=obalunenko_georgia-tax-calculator)
![coverbadger-tag-do-not-edit](https://img.shields.io/badge/coverage-72.96%25-brightgreen?longCache=true&style=flat)


# georgia-tax-calculator

Calculates income taxes in Georgia.

- Fetches official rates from the [nbg.gov.ge](https://nbg.gov.ge) for the date of income.
- Converts income to GEL.
- Calculate taxes amount according to specified Taxes Category.


## Usage

1. Download binary from [![Latest release artifacts](https://img.shields.io/badge/artifacts-download-blue.svg)](https://github.com/obalunenko/georgia-tax-calculator/releases/latest)

2. Run `ge-tax-calc run` and follow instructions

All available flags, commands and usage:

```text
NAME:
   ge-tax-calc - A command line tool helper for preparing tax declaration in Georgia 

USAGE:
   ge-tax-calc [global options] command [command options] [arguments...]

DESCRIPTION:
   Helper tool for preparing tax declarations in Georgia.
   It get income amount in received currency, converts it to GEL according to
   official rates on date of income and calculates tax amount
   according to selected taxes category.

AUTHOR:
   Oleg Balunenko <oleg.balunenko@gmail.com>

COMMANDS:
   run      Runs taxes calculations
   convert  Runs currency converter
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```


### Demo

[![asciicast](https://asciinema.org/a/rqN2ZwN72LNAfRQoGdmJmV4j5.svg)](https://asciinema.org/a/rqN2ZwN72LNAfRQoGdmJmV4j5)
