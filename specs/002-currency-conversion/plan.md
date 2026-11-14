# Implementation Plan: Currency Conversion

**Feature Directory**: `specs/002-currency-conversion` | **Branch**: N/A (migrated)

**Date**: 2026-11-14 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/002-currency-conversion/spec.md`

**Note**: Reverse-engineered from existing code (`status: migrated`). It
describes what is implemented, not a proposal.

## Summary

Both interfaces collect a date (year, month name, day), an amount string, and
two currency codes, then call `service.Convert`. The service parses the input
with `pkg/dateutils` and `pkg/moneyutils` and delegates to
`internal/converter`, which fetches both currencies from the cached NBG client
for that date and computes a cross rate through GEL using per-unit rates.

## Technical Context

**Language/Version**: Go 1.27.3, single module with committed `vendor/`.

**Primary Dependencies**: `shopspring/decimal` (inside `pkg/moneyutils`);
`charm.land/bubbletea/v2` and `urfave/cli/v3` (CLI); `mymmrac/telego` (bot);
`pkg/nbggovge` (rates).

**Storage**: None. Bot sessions are in memory (see
[004-telegram-bot-platform](../004-telegram-bot-platform/plan.md)); NBG
responses are cached for 1 hour.

**Testing**: testify table-driven tests; `mock.NewClient` with
`internal/converter/testdata/2023-09-09-all.json`; `mockConverter` in
`internal/service/service_test.go`; Bubble Tea model tests in
`cmd/ge-tax-calc/convert_menu_test.go`. No bot tests.

**Target Platform**: CLI on linux/darwin/windows; bot on linux/darwin and as
a Docker image.

**Project Type**: Feature spanning both interfaces over the service layer.

**Performance Goals**: Not specified; one NBG request per uncached date and
currency pair.

**Constraints**: Money math through `pkg/moneyutils` (constitution III);
interfaces call only `internal/service` (constitution I); CLI/bot parity
(constitution II).

**Scale/Scope**: One service method, one converter, one CLI form, one bot
flow.

## Constitution Check

- [x] I. `cmd/ge-tax-calc` and `cmd/bot` call `service.Convert`; neither
      imports `internal/converter` or `pkg/nbggovge`.
- [ ] II. Flows match, but amount validation differs: the CLI uses
      `moneyutils.Parse`, the bot uses `fmt.Sscanf` (see tasks Gaps).
      BUG-001: resolved by one shared rule in `internal/service` (decision 6).
- [x] III. Rate and amount arithmetic and rounding use `pkg/moneyutils`.
- [x] IV. Rates come from `nbggovge.Client`; tests use `mock.NewClient` and
      fixtures.
- [x] V. `pkg/moneyutils` and `pkg/dateutils` do not import `internal/`.
- [ ] VI. Converter, service, and CLI form are tested; the bot flow and
      `pkg/moneyutils` are not (allowed by VI, recorded as gaps).
- [x] VII. No generated code involved.
- [ ] Known deviations: `service.Convert` starts the terminal spinner, also
      for bot requests (existing deviation, not extended).

## Project Structure

### Documentation (this feature)

```text
specs/002-currency-conversion/
├── spec.md
├── plan.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
├── ge-tax-calc/
│   ├── commands.go          # `convert` command → menuConvert
│   ├── menu.go              # menuConvert: run form, call service, print
│   ├── convert_menu.go      # Bubble Tea form: steps, summary, confirm/restart
│   └── prompts.go           # select/input/confirm prompts, year/month/day/currency options, validateMoneyInput
└── bot/
    ├── handlers.go          # handleConvert, handleConvertTextInput, handleConvertCallback, formatting, validateMoney
    ├── session.go           # flowConvert, convertStep, convertReq
    └── keyboards.go         # year/month/day/currency/confirm keyboards
internal/
├── service/service.go       # ConvertRequest/Response, Convert
├── converter/converter.go   # Converter, cross rate, rounding, ErrCurrencyNotSet
└── models/models.go         # Money
pkg/
├── moneyutils/moneyutils.go # Parse, Multiply, Div, Round, ToString
└── dateutils/dateutils.go   # ParseYear/Month/Day, months and days up to a date
```

**Structure Decision**: Interfaces own input collection and presentation;
`internal/service` owns parsing and orchestration; `internal/converter` owns
the rate math.

**Validation**: `go test -race ./internal/converter/... ./internal/service/...
./cmd/ge-tax-calc/... ./pkg/dateutils/...`; `make vet`, `make lint-pipeline`,
`make test`, `make build`.

## Technical Decisions

1. **Cross rate through GEL**: NBG publishes GEL per `Quantity` units of each
   currency; dividing per-unit rates gives any pair, with GEL as a synthetic
   currency at rate 1.
2. **Precision**: decimal arithmetic in `moneyutils`, values exposed as
   `float64`; amounts rounded to 2 places, rates to 4.
3. **String requests**: `ConvertRequest` carries strings so both interfaces
   pass raw input; parsing lives in the service.
4. **Date limits in interfaces**: year (2016 to now), month, and day options
   up to today come from `dateutils` and are offered as choices, so invalid
   dates cannot be selected.
5. **Confirmation step**: both interfaces show a summary and restart the whole
   form on "No".
6. **Shared amount rule** (BUG-001): `internal/service` exports
   `ParseAmount` (amount > 0) and `ParseYearIncome` (year income ≥ 0, used
   by 003), both trimming spaces, parsing with `moneyutils.Parse`, and
   wrapping a sentinel `ErrInvalidAmount`. `Convert` and `Calculate` call
   them; the CLI `validateMoneyInput` and the bot `validateMoney` call them,
   so all three apply one rule (constitution I and II). `cmd/` packages
   cannot share code with each other, so the rule cannot live in either of
   them. (FR-006, FR-010, FR-011, FR-014)

**Bugfix**: 2026-11-15 — BUG-001 Updated from bugfix patch

## Complexity Tracking

No justified violations. Partially met gates (II, VI) and the known spinner
deviation are recorded in [tasks.md](tasks.md).

| Factor | Assessment |
|--------|------------|
| Files | ~12 source files across 7 packages |
| Dependency depth | cmd → service → converter → nbggovge |
| State | CLI form model; bot per-user session |
| External coupling | NBG API through the cached client |
