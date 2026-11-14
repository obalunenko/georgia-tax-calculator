---

description: "Migrated task list for tax calculation"
---

# Tasks: Tax Calculation

**Input**: Design documents from `/specs/003-tax-calculation/`

**Prerequisites**: plan.md, spec.md; depends on `specs/002-currency-conversion`

**Status**: migrated — all tasks reflect code that already exists on `master`.
Gaps found during migration are listed at the end without checkboxes; they are
candidates for follow-up features or bug reports, not pending work for this
feature.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)

## Phase 1: Feature Preparation

- [x] T001 Add the `stringer` directive for `TaxType` and run `make codegen` to produce `internal/taxes/taxtype_string.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

- [x] T002 Define `TaxType` (Individual Entrepreneur, Small Business, Employment) with sentinel, `Valid`, `ParseTaxType`, and errors in `internal/taxes/calculator.go`
- [x] T003 Define `TaxRate`, the rate table (3%, 1%, 20%), `Rate`, `AllTaxRates`, and `AllTaxTypes` in `internal/taxes/calculator.go`

---

## Phase 3: User Story 1 - Calculate tax for incomes in any currency (Priority: P1) 🎯 MVP

### Tests for User Story 1

- [x] T004 [P] [US1] Table-driven `Calc`, `Rate`, and `AllTaxTypes` tests in `internal/taxes/calculator_test.go`
- [x] T005 [P] [US1] Service tests (one income, multiple incomes, invalid request, converter error) and `CalculateResponse.String` test in `internal/service/service_test.go`

### Implementation for User Story 1

- [x] T006 [US1] Implement `Calc` (rate × amount with `moneyutils`, rounded to 2 places) in `internal/taxes/calculator.go`
- [x] T007 [US1] Define `Income`, `CalculateRequest`, `CalculateResponse` and its `String` in `internal/service/service.go`
- [x] T008 [US1] Implement `service.Calculate`: parse category and year income, convert each income to GEL, calculate and sum taxes (`internal/service/service.go`)

**Checkpoint**: Tax calculation works through the service.

---

## Phase 4: User Story 2 - Calculate in the CLI (Priority: P2)

### Tests for User Story 2

- [x] T009 [P] [US2] Form tests (tax details flow, income collection, declined restart, empty list rendering) in `cmd/ge-tax-calc/tax_menus_test.go`

### Implementation for User Story 2

- [x] T010 [US2] Implement `taxTypeOptions` (category with rate, sorted) in `cmd/ge-tax-calc/prompts.go`
- [x] T011 [US2] Implement the tax details form and the income loop form with summaries and confirmation in `cmd/ge-tax-calc/tax_menus.go`
- [x] T012 [US2] Register the `run` command, print the declarations link, call the service, and print the result (`cmd/ge-tax-calc/commands.go`, `menuCalcTaxes` in `cmd/ge-tax-calc/menu.go`)

**Checkpoint**: `ge-tax-calc run` works end to end.

---

## Phase 5: User Story 3 - Calculate in the Telegram bot (Priority: P2)

### Implementation for User Story 3

- [x] T013 [P] [US3] Add `flowCalculate`, `calcStep`, `calcReq`, `incomes`, and `currentInc` to the session in `cmd/bot/session.go`
- [x] T014 [P] [US3] Build the tax type keyboard and parse the selected item in `cmd/bot/keyboards.go`
- [x] T015 [US3] Implement `/calculate`, text input for year income and amounts, step callbacks, add-more loop, confirmation, and result formatting in `cmd/bot/handlers.go`

**Checkpoint**: `/calculate` works end to end in the bot.

---

## Phase N: Polish & Cross-Cutting Concerns

- [x] T016 [P] Document the `run` command and bot `/calculate` in `README.md`

---

## Gaps Found During Migration

Not tasks of this feature. Track them with `/speckit-bugfix-report` or a new
`/speckit-specify` feature.

1. **Bot amount parsing** (`fmt.Sscanf`, `cmd/bot/handlers.go:697`) accepts
   `12abc` and `1,500` for year income and income amounts; the error appears
   only after confirmation. See 002 gap 1.
2. **Different restart scope.** On "No" at the final review, the CLI restarts
   only income entry (`cmd/ge-tax-calc/tax_menus.go`), while the bot restarts
   the whole flow including category and year income
   (`cmd/bot/handlers.go:393`).
3. **Year income has no effect and incomes are not limited to one year.**
   Year income is only added to the reported total; incomes from 2016 to the
   current year can be mixed. Open question recorded as FR-013.
4. **No bot tests and no end-to-end calculation test.** Service tests use
   `mockConverter` (1:1 conversion); no test runs `Calculate` with fixture NBG
   rates.
5. **Per-income rounding.** Tax is rounded per income before summing, so the
   total can differ by a few tetri from rounding the total once. Not verified
   against declaration rules.
6. **Declarations link only in the CLI.** `menuCalcTaxes` prints a link to
   `https://decl.rs.ge/decls.aspx` when the terminal supports hyperlinks; the
   bot does not show it.
7. **Error text for invalid categories in `taxes.Calc`** uses the stringer
   fallback (`TaxType(N): tax type not supported`).
8. **Inherited from 002**: spinner in the service, unused `survey` tags,
   negative and zero amounts accepted, date validation only in the interfaces.
