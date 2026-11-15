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

Amounts are plain numbers with a dot as the decimal separator (e.g. `1500.50`).
Income and conversion amounts must be greater than 0; income since the start
of the year can be `0`. The CLI and the bot apply the same rules.

Tax calculation rules:

- Tax categories: Employment (20%), Small Business (1%), and Small Business
  (over limit) (3%, formerly "Individual Entrepreneur").
- One calculation covers one calendar year: after the first income, only its
  year can be chosen.
- The tax is calculated once on the total income converted to GEL.
- For Small Business, if income since the start of the year plus the new
  incomes exceeds 500 000 GEL, the result shows a warning; the tax is still
  calculated at 1%. By the Tax Code (Article 90(2)), income from the start
  of the month in which the limit is exceeded until the end of the year is
  taxed at 3%; use Small Business (over limit) for those incomes. Rates and
  the limit follow the Tax Code as of 30.10.2026. The warning also tells you to
  recalculate the month's declaration with Small Business (over limit), where
  field (16) is 3.00.

Small Business monthly declaration (both Small Business categories):

- One calculation covers one month, as the declaration does: after the first
  income, only its month can be chosen. Run one calculation per month.
- For income since the start of the year, enter field (15) of your previous
  month's declaration (0 for your first declaration of the year). It is rounded
  to the tetri.
- Each income gets a type, which selects its declaration field:

  | Type | Field | Covers |
  |------|-------|--------|
  | Cash | (18) | cash payments (cash register turnover) |
  | POS terminal | (19) | card payments through a POS terminal |
  | Non-cash | (20) | bank transfers and other non-cash payments (not POS) |
  | Other | (21) | services not taxed at source, barter, and other income |
  | Unpaid supply | (22) | goods or services supplied, not paid for this month |

- The result lists the values to enter in the declaration, with two decimals.
  The Telegram bot shows each value in monospace, so a tap copies it.

  ```text
  Declaration fields (Small Business monthly declaration):
  (15) Income since the start of the year: 35942.00
  (16) Tax rate, %: 1.00
  (17) Taxable income for the month: 4150.00
  (18) Cash: 3950.00
  (19) POS terminal: 0.00
  (20) Non-cash: 200.00
  (21) Other: 0.00
  (22) Unpaid supply: 0.00
  ```

Employment calculations do not ask for the type or limit the month, and show
no declaration fields.

All available flags, commands and usage (`ge-tax-calc --help`; the `VERSION`
section with the build information of your binary is omitted here, see
`ge-tax-calc --version`):

```text
NAME:
   ge-tax-calc - A command line tool helper for preparing tax declaration in Georgia 

USAGE:
   ge-tax-calc [global options] [command [command options]]

DESCRIPTION:
   Helper tool for preparing tax declarations in Georgia.
   It gets the income amount in the received currency, converts it to GEL at the
   official NBG rate on the date of income, and calculates the tax amount for the
   selected tax category.

AUTHOR:
   Oleg Balunenko <oleg.balunenko@gmail.com>

COMMANDS:
   run      Runs taxes calculations
   convert  Runs currency converter
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

An unknown command prints `No help topic for '<command>'` and exits with
status 3.

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

   Optional logging level:

   ```sh
   export TELEGRAM_BOT_LOG_LEVEL=debug # debug|info|warn|error|fatal
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
   # optional: TELEGRAM_BOT_LOG_LEVEL=debug
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

   The container runs as a non-root user (UID 10001). Known chats are stored in
   the `bot-data` volume mounted at `/data`, so they survive `make bot-down` and
   image rebuilds; `docker compose down -v` removes them.

### Bot Commands

- `/start` — Welcome message and command overview
- `/calculate` — Start the tax calculation flow (guided step-by-step)
- `/convert` — Start the currency conversion flow (guided step-by-step)
- `/cancel` — Cancel the current operation
- `/help` — Show available commands

The bot sets this list, except `/start`, as its Telegram command menu at every
start, so Telegram suggests the commands when you type `/`. You do not need
BotFather's `/setcommands`; a list set there is replaced on the next start.

Only the buttons of the latest message are accepted: pressing a button from an
earlier message replies with "Unexpected input" and keeps the current step.

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
