![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/obalunenko/georgia-tax-calculator)
[![GoDoc](https://godoc.org/github.com/obalunenko/georgia-tax-calculator?status.svg)](https://godoc.org/github.com/obalunenko/georgia-tax-calculator)
[![Latest release artifacts](https://img.shields.io/github/v/release/obalunenko/georgia-tax-calculator)](https://github.com/obalunenko/georgia-tax-calculator/releases/latest)
[![Go [lint, test]](https://github.com/obalunenko/georgia-tax-calculator/actions/workflows/go.yml/badge.svg)](https://github.com/obalunenko/georgia-tax-calculator/actions/workflows/go.yml)
[![Lint & Test & Build & Release](https://github.com/obalunenko/georgia-tax-calculator/actions/workflows/release.yml/badge.svg)](https://github.com/obalunenko/georgia-tax-calculator/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/obalunenko/georgia-tax-calculator)](https://goreportcard.com/report/github.com/obalunenko/georgia-tax-calculator)
[![codecov](https://codecov.io/gh/obalunenko/georgia-tax-calculator/branch/master/graph/badge.svg)](https://codecov.io/gh/obalunenko/georgia-tax-calculator)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=obalunenko_georgia-tax-calculator&metric=alert_status)](https://sonarcloud.io/summary/overall?id=obalunenko_georgia-tax-calculator)
![coverbadger-tag-do-not-edit](https://img.shields.io/badge/coverage-74.65%25-brightgreen?longCache=true&style=flat)

# georgia-tax-calculator

Calculates income taxes in Georgia.

- Fetches official rates from the [nbg.gov.ge](https://nbg.gov.ge) for the date of income.
- Converts income to GEL.
- Calculate taxes amount according to specified Taxes Category.
- **Includes caching** to reduce HTTP requests and improve performance.

## Features

- **Tax Calculation**: Calculate Georgian income taxes based on official rates
- **Currency Conversion**: Convert between currencies using NBG official rates  
- **Smart Caching**: Automatic caching of currency rates to minimize API calls
- **Interactive CLI**: User-friendly command-line interface
- **Telegram Bot**: Interactive Telegram bot interface for tax calculations
- **Multi-tax Categories**: Support for different Georgian tax categories

## Usage

1. Download binary
   from [![Latest release artifacts](https://img.shields.io/badge/artifacts-download-blue.svg)](https://github.com/obalunenko/georgia-tax-calculator/releases/latest)

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

#### Taxes calculations

[![asciicast](https://asciinema.org/a/rqN2ZwN72LNAfRQoGdmJmV4j5.svg)](https://asciinema.org/a/rqN2ZwN72LNAfRQoGdmJmV4j5)

#### Cuurency conversion

[![asciicast](https://asciinema.org/a/DhjfE2pOQa2PteSqKgOIYMBA5.svg)](https://asciinema.org/a/DhjfE2pOQa2PteSqKgOIYMBA5)

## Telegram Bot

The `ge-bot` binary provides a Telegram bot interface with the same functionality as the CLI.

### Running the Bot

#### Option 1 — Binary

1. Create a bot via [@BotFather](https://t.me/BotFather) and get your bot token.

2. Set the `TELEGRAM_BOT_TOKEN` environment variable:

   ```sh
   export TELEGRAM_BOT_TOKEN=your_bot_token_here
   ```

3. Run the bot:

   ```sh
   ge-bot
   ```

#### Option 2 — Docker Compose (local deployment)

1. Create a bot via [@BotFather](https://t.me/BotFather) and get your bot token.

2. Copy the example env file and fill in your token:

   ```sh
   cp deployments/docker-compose/.env.example deployments/docker-compose/.env
   # edit .env and set TELEGRAM_BOT_TOKEN=<your token>
   ```

3. Start the bot:

   ```sh
   make bot-up
   ```

4. Other useful commands:

   | Command | Description |
   |---------|-------------|
   | `make bot-up` | Build the image and start the bot in the background |
   | `make bot-down` | Stop and remove the bot container |
   | `make bot-logs` | Follow the bot's log output |

### Bot Commands

- `/start` — Welcome message and command overview
- `/calculate` — Start the tax calculation flow (guided step-by-step)
- `/convert` — Start the currency conversion flow (guided step-by-step)
- `/cancel` — Cancel the current operation
- `/help` — Show available commands

## For Developers

### NBG API Client with Caching

The application includes a caching layer for the National Bank of Georgia API client to improve performance:

```go
// Basic usage with caching (1-hour TTL by default)
client := nbggovge.NewCached()

// Custom TTL
client := nbggovge.NewCachedWithTTL(time.Minute * 30)

// No expiration (cache until restart)
client := nbggovge.NewCachedWithTTL(0)
```

See [pkg/nbggovge/README_CACHE.md](pkg/nbggovge/README_CACHE.md) for detailed documentation and examples.
