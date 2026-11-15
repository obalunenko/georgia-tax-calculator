# Implementation Plan: Income Type and Declaration Field Numbers

**Feature Directory**: `specs/007-income-type-declaration-fields` | **Branch**: `feat/income-type-declaration-fields`

**Date**: 2026-11-15 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/007-income-type-declaration-fields/spec.md`

## Summary

For the Small Business categories, each income gets a type (Cash, POS
terminal, Non-cash, Other, Unpaid supply) that maps to fields (18)–(22) of
the monthly declaration. `service.Calculate` validates the types and a
one-month rule before any conversion, and returns a `Declaration` with the
values of fields (15)–(22), formatted with two decimals, next to the existing
totals. The CLI prints the service text; the bot prints the same lines and
marks each value as a `code` entity so Telegram copies it with a tap. Both
interfaces ask for the type after the currency, limit later incomes to the
first income's month, ask for the year income as field (15) of the previous
declaration, and the limit warning gains a recalculation line. Employment is
unchanged.

## Technical Context

**Language/Version**: Go 1.27.3, single module
`github.com/obalunenko/georgia-tax-calculator` with committed `vendor/`.

**Primary Dependencies**: urfave/cli v3 and Bubble Tea v2 (CLI), telego (bot;
`MessageEntity` and `EntityTypeCode` from the vendored package),
shopspring/decimal (via `pkg/moneyutils`). No dependency changes.

**Storage**: None. The calculation month lives in the CLI income model and
the bot session (in memory), derived from the first captured income.

**Testing**: testify table tests in `internal/taxes/income_type_test.go`
(new), `pkg/moneyutils/moneyutils_test.go` (new),
`internal/service/declaration_test.go` (new), updated
`internal/service/service_test.go` and `year_income_test.go`,
`cmd/ge-tax-calc/tax_menus_test.go`, `cmd/bot/keyboards_test.go`,
`cmd/bot/handlers_test.go`, `cmd/bot/flows_test.go` (fake Telegram API, which
records `entities`). `go test -race ./...`. Service tests use
`mockConverter` (1:1), so field values are exact.

**Target Platform**: CLI on linux/darwin/windows; bot on linux/darwin and as a
Docker image.

**Performance Goals**: N/A. New checks run before conversions, so an invalid
request makes no NBG calls.

**Constraints**: Money sums and two-decimal formatting through
`pkg/moneyutils` (constitution III); interfaces call `internal/service` and
use `internal/taxes` only for value types (`IncomeType`, `TaxType`)
(constitution I); CLI/bot parity with one text source in the service
(constitution II); no new generated code.

**Scale/Scope**: One service method, one new enum, one new formatting helper,
two interface flows (`run`, `/calculate`); two of three categories; five
income types; eight declaration fields.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] I. The interfaces get validation, the declaration block, and the
      warning text from `internal/service`; they use `internal/taxes` for the
      `IncomeType` list and `TaxType.IsSmallBusiness`, as they already do for
      category lists. No new internal package is imported by `cmd/`.
- [x] II. Both interfaces change together: same type options and order, same
      month rule, same year-income prompt meaning, same block lines (from the
      service), same warning. Only the bot's monospace values differ, which
      is presentation (spec FR-014).
- [x] III. Per-type sums use `moneyutils.Add`; the rate percent uses
      `moneyutils.Multiply`; two-decimal output uses the new
      `moneyutils.ToStringFixed`. Rates come from the existing table;
      `TaxType` gains a method only, no enum or rate change.
- [x] IV. N/A (no NBG changes); tests keep using `mockConverter` and the
      bot's `fakeService`.
- [x] V. `pkg/moneyutils` gains a function that imports only
      `shopspring/decimal`.
- [x] VI. New behavior in `internal/taxes`, `internal/service`, and
      `pkg/moneyutils` gets table-driven testify tests passing with `-race`;
      interface behavior is tested in the CLI model tests and the bot
      keyboard, handler, and flow tests.
- [x] VII. No generated code changes: `IncomeType` uses a descriptor table
      instead of stringer ([research.md](research.md) R1); `vendor/` is
      untouched.
- [x] Known deviations are not extended.

Post-design re-check (after Phase 1): unchanged, all gates pass.

## Project Structure

### Documentation (this feature)

```text
specs/007-income-type-declaration-fields/
├── plan.md              # This file
├── research.md          # Phase 0 decisions
├── data-model.md        # Phase 1 entities
├── quickstart.md        # Phase 1 validation guide
├── contracts/
│   └── calculation-output.md   # Prompts, type options, result block, warning, errors
├── checklists/requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
pkg/moneyutils/
├── moneyutils.go          # ToStringFixed (new)
└── moneyutils_test.go     # (new)
internal/
├── taxes/
│   ├── income_type.go         # IncomeType, descriptor table, ParseIncomeType, AllIncomeTypes, ErrInvalidIncomeType (new)
│   ├── income_type_test.go    # (new)
│   ├── calculator.go          # TaxType.IsSmallBusiness
│   └── calculator_test.go
└── service/
    ├── service.go             # Income.Type; IncomeResponse; CalculateResponse.Incomes/Declaration and String(); Calculate wiring
    ├── declaration.go         # Declaration, DeclarationField, newDeclaration, checkIncomeTypes (new)
    ├── declaration_test.go    # (new)
    ├── year_income.go         # ErrIncomesSpanMonths, checkSingleMonth; LimitWarning recalculation line
    ├── year_income_test.go
    └── service_test.go        # Small Business cases get income types; Incomes as []IncomeResponse
cmd/
├── ge-tax-calc/
│   ├── menu.go                # pass the tax type to runIncomeMenu
│   ├── tax_menus.go           # year-income prompt per category; newIncomeModel(taxType); month limit; incomeStepType; type in lists and review
│   ├── prompts.go             # incomeTypeOptions
│   └── tax_menus_test.go
└── bot/
    ├── session.go             # calcStepIncomeType, step ID calc_inc_type
    ├── keyboards.go           # income type options; month limited after the first income
    ├── handlers.go            # year-income text per category; currency → type step; lists, summary, result with entities
    ├── send_message.go        # UTF-16 code entity helper (or a small new file next to it)
    ├── keyboards_test.go
    ├── handlers_test.go
    └── flows_test.go          # Small Business flow presses a type
README.md                      # income types, one-month rule, year-income prompt, declaration block (FR-011)
specs/006-year-income-rules/spec.md  # note: 007 extends FR-001 text and narrows FR-005 for Small Business
```

**Structure Decision**: The income type is a value type in `internal/taxes`
next to `TaxType`, because both interfaces need its list and constitution I
allows them that package. The declaration block and the new checks live in
`internal/service` (new `declaration.go`, month check next to the 006 year
check in `year_income.go`), so the text has one source. The bot entity
helper stays in `cmd/bot`, since only Telegram needs it. No new packages or
cross-package dependencies.

**Validation**: `go vet ./...`, `go test -race ./...`, `make lint-pipeline`
(full `.golangci.yml` on changed lines, mind funlen in `handleCalcCallback`
and `incomeModel.advance`); manual CLI and bot runs from
[quickstart.md](quickstart.md). No `make codegen` or `make sync-vendor`.

## Complexity Tracking

No constitution violations.
