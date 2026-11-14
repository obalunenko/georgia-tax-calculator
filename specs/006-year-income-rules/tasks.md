---

description: "Task list for year income rules in tax calculation"
---

# Tasks: Year Income Rules for Tax Calculation

**Input**: Design documents from `/specs/006-year-income-rules/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/calculation-output.md, quickstart.md

**Tests**: Behavior changes in `internal/taxes` and `internal/service` MUST include tests (constitution VI): testify `assert`/`require`, table-driven `t.Run`, `-race`. Service tests use `mockConverter` (1:1 conversion to GEL). Interface tests cover the CLI form model and bot keyboards; no test calls NBG.

**Organization**: Grouped by user story. US1–US3 all change `service.Calculate` in `internal/service/service.go`, so their service tasks run in order T004 → T008 → T012; tests and interface tasks in other files can run in parallel.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: User story from spec.md (US1–US4)

## Phase 1: Feature Preparation

No dependency, fixture, or tooling changes are needed.

---

## Phase 2: Foundational (Blocking Prerequisites)

- [X] T001 Create `internal/service/year_income.go` with the package's file header conventions (package comment lives in `service.go`) to hold the year and limit rules, and `internal/service/year_income_test.go` for their tests

**Checkpoint**: Files exist; nothing else depends on new code yet.

---

## Phase 3: User Story 1 - Limit warning (Priority: P1) 🎯 MVP

**Goal**: For Small Business, the result warns when the reported year income exceeds the annual turnover limit; the tax is unchanged.

**Independent Test**: Small Business, year income `499000`, one income 2000 GEL → tax 20 GEL and the warning "Year income 501000 GEL exceeds the Small Business limit of 500000 GEL by 1000 GEL…"; total exactly 500000 → no warning.

- [X] T002 [US1] Add the exported constant `SmallBusinessTurnoverLimit float64 = 500_000` ("Small Business annual turnover limit, GEL per calendar year"; doc comment says it is pending confirmation against the Tax Code, 003 FR-015) in `internal/taxes/calculator.go`
- [X] T003 [US1] In `internal/service/year_income.go` add `LimitWarning` (`Limit`, `YearIncome`, `Excess` as `models.Money` in GEL) and `newLimitWarning(tt taxes.TaxType, yearIncome float64) *LimitWarning` returning a warning only when `tt == taxes.TaxTypeSmallBusiness` and `yearIncome > taxes.SmallBusinessTurnoverLimit` ("Present only when the category is Small Business and `YearIncome > Limit`"); `Excess` = `moneyutils.Add(yearIncome, -limit)`; `String()` returns exactly `Year income <YearIncome> GEL exceeds the Small Business limit of <Limit> GEL by <Excess> GEL. The tax above is calculated at the Small Business rate.` using `models.Money.String()`
- [X] T004 [US1] Add `LimitWarning *LimitWarning` to `CalculateResponse`, set it from the reported year income in `Calculate`, and append `"\n⚠️ " + LimitWarning.String()` after the `Taxes:` line in `CalculateResponse.String` in `internal/service/service.go`
- [X] T005 [P] [US1] Table tests in `internal/service/year_income_test.go`: Small Business above (501000 → excess 1000), equal (500000 → nil), below, Employment and Small Business (over limit) above (nil), and the exact `String()` text from `contracts/calculation-output.md`
- [X] T006 [P] [US1] Tests in `internal/service/year_income_test.go` (`TestCalculateYearIncomeRules`; was planned for `service_test.go`): `Calculate` with Small Business, year income `499000`, income 2000 → `Tax` 20 GEL and non-nil `LimitWarning`; `CalculateResponse.String` ends with the warning line
- [X] T007 [P] [US1] Append `"\n⚠️ " + resp.LimitWarning.String()` to the result in `formatCalcResult` when the warning is present in `cmd/bot/handlers.go`; test both cases in `cmd/bot/handlers_test.go`

**Checkpoint**: US1 works in both interfaces (CLI prints `CalculateResponse.String`).

---

## Phase 4: User Story 2 - One calendar year (Priority: P1)

**Goal**: Later incomes can only be in the first income's year; the service rejects mixed years.

**Independent Test**: After an income in 2025, only 2025 is offered (CLI and bot); a service request with 2024 and 2025 fails with `ErrIncomesSpanYears`.

- [X] T008 [US2] In `internal/service/year_income.go` add `ErrIncomesSpanYears = errors.New("incomes span more than one calendar year")` and `checkSingleYear(incomes []Income) error` comparing years parsed with `dateutils.ParseYear` ("All incomes have the same parsed year"); call it in `Calculate` in `internal/service/service.go` after parsing the year income and before any conversion, returning it wrapped (`failed to check incomes: %w`)
- [X] T009 [P] [US2] Tests in `internal/service/year_income_test.go`: no incomes, one year, two years → `ErrIncomesSpanYears`, invalid year → error; and a `Calculate` case with a converter that fails the test if called, proving no conversion runs for mixed years
- [X] T010 [P] [US2] Add `yearPrompt()` to `incomeModel` in `cmd/ge-tax-calc/tax_menus.go`: with captured incomes, a select prompt with only `incomes[0].Year`, preselected; otherwise `yearOptions()` and `defaultYearValue()`; use it in `Init`, the "add another" branch, and the review-"No" restart (which clears incomes first); test in `cmd/ge-tax-calc/tax_menus_test.go`
- [X] T011 [P] [US2] In `stepOptions` for `stepCalcYear` in `cmd/bot/keyboards.go` return only `sess.incomes[0].Year` when incomes exist; tests in `cmd/bot/keyboards_test.go`: one option after the first income, a callback with another year is rejected, all years before the first income

**Checkpoint**: US2 works in both interfaces and in the service.

---

## Phase 5: User Story 3 - Tax on the total (Priority: P2)

**Goal**: Tax = `round(total converted income × rate, 2)`, computed once.

**Independent Test**: Three incomes of 0.50 GEL at 1% → tax 0.02 GEL.

- [X] T012 [US3] In `Calculate` in `internal/service/service.go` sum the converted incomes with `moneyutils.Add` inside the loop and call `taxes.Calc` once on the total after the loop; remove the per-income tax sum
- [X] T013 [P] [US3] Tests in `internal/service/year_income_test.go` (`TestCalculateYearIncomeRules`) and the existing cases in `internal/service/service_test.go`: three incomes of `0.50` GEL with Small Business → `Tax` 0.02 GEL; existing Employment case (1000 + 200, year income 67.99 → tax 240, year income 1267.99) still passes

**Checkpoint**: US3 done; results match a single rounding of the total.

---

## Phase 6: User Story 4 - Renamed over-limit category (Priority: P3)

**Goal**: The 3% category is "Small Business (over limit)"; "Individual Entrepreneur" is gone.

**Independent Test**: Category lists show Employment, Small Business, Small Business (over limit); parsing "Individual Entrepreneur" fails.

- [X] T014 [US4] Rename `TaxTypeIndividualEntrepreneur` to `TaxTypeSmallBusinessOverLimit` with the line comment `// Small Business (over limit)` and its doc comment, keeping enum value 1, and update `stringToTaxType` and `taxrates` in `internal/taxes/calculator.go`
- [X] T015 [US4] Run `make codegen` to regenerate `internal/taxes/taxtype_string.go` (constitution VII; do not edit by hand)
- [X] T016 [P] [US4] Update `internal/taxes/calculator_test.go` for the new name and add: `ParseTaxType("Individual Entrepreneur")` returns `ErrInvalidTaxType`, `ParseTaxType("small business (over limit)")` returns the new type, `TaxRate.String()` is `Small Business (over limit) 3 %`
- [X] T017 [P] [US4] Add to `cmd/bot/keyboards_test.go` that `taxTypeOptions` contains `{text: "Small Business (over limit) (3%)", value: "Small Business (over limit)"}` and its callback data fits 64 bytes; add to `cmd/ge-tax-calc/prompts_test.go` that `taxTypeOptions` lists Employment, Small Business, Small Business (over limit) in this order

**Checkpoint**: US4 done in both interfaces.

---

## Phase N: Polish & Cross-Cutting Concerns

- [X] T018 [P] Update `specs/003-tax-calculation/spec.md` (FR-001 category name; FR-004, FR-013, FR-014 implemented by 006), `plan.md`, and `tasks.md` gaps 3 and 5
- [X] T019 [P] Document the limit warning, the one-year rule, and the category names in the "Usage" section of `README.md`
- [X] T020 Run `go vet ./...`, `go test -race ./...`, `make lint-pipeline`, full `.golangci.yml` on changed lines, and quickstart scenarios 1–5 with `go run ./cmd/ge-tax-calc run` (`specs/006-year-income-rules/quickstart.md`). Done on 2026-11-15: all automated checks pass (full lint: only `goconst` in tests); scenarios 1–3 and the mixed-year error were run through `service.New()` with the live NBG client; the interactive TUI and bot runs were not done by hand (covered by form and keyboard tests)

---

## Dependencies & Execution Order

- **Phase 2** (T001) → all stories.
- **US1**: T002 → T003 → T004; T005–T007 after T003/T004.
- **US2**: T008 after T004 (same function, `service.go`); T009 after T008; T010 and T011 touch only interface files and can start any time after T001.
- **US3**: T012 after T008 (same function); T013 after T012.
- **US4**: T014 → T015 → T016/T017; independent of US1–US3 except that T002 and T014 edit the same file (do them in order).
- **Polish**: after all stories.

Story completion order: US1 → US2 → US3 → US4 (US4 can also go first).

## Parallel Examples

- US1: after T004, run T005 (`year_income_test.go`), T006 (`service_test.go`), and T007 (`cmd/bot`) together.
- US2: T010 (`cmd/ge-tax-calc`) and T011 (`cmd/bot`) together; T009 after T008.
- US4: T016 (`internal/taxes`) and T017 (`cmd/`) together after T015.

## Implementation Strategy

- **MVP**: US1 (limit warning) — it is the reason the year income is collected.
- Then US2 (one year), which makes the warning meaningful for multi-income calculations; US3 (rounding); US4 (rename).
- Each story is shippable on its own; run `go test -race ./...` after each.
- When the Tax Code text arrives (003 FR-015), only `SmallBusinessTurnoverLimit`, the rates, and the Assumptions need updating.
