---

description: "Task list for income type and declaration field numbers"
---

# Tasks: Income Type and Declaration Field Numbers

**Input**: Design documents from `/specs/007-income-type-declaration-fields/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/calculation-output.md, quickstart.md

**Tests**: Behavior changes in `pkg/moneyutils`, `internal/taxes`, and `internal/service` MUST include tests (constitution VI): testify `assert`/`require`, table-driven `t.Run`, `-race`. Service tests use `mockConverter` (1:1 conversion to GEL). Interface tests cover the CLI Bubble Tea models and the bot keyboards, handlers, and flows over the fake Telegram API; no test calls NBG.

**Organization**: Grouped by user story. US2 (choose the type) is implemented before US1 (see the fields) because the block needs the types; both are P1 and together form the MVP (spec US2 "Why this priority"). US1–US3 all change `Calculate` in `internal/service/service.go`, so their service tasks run in order T005 → T009 → T014; tests and interface tasks in other files run in parallel.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks); it does not itself request spawning additional agents
- **[Story]**: User story from spec.md (US1–US4)

## Phase 1: Feature Preparation

No dependency, fixture, generated-code, or tooling changes are needed (plan: no `make codegen`, no `make sync-vendor`).

---

## Phase 2: Foundational (Blocking Prerequisites)

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T001 [P] Add `ToStringFixed(v float64, places int32) string` with a doc comment to `pkg/moneyutils/moneyutils.go`, returning `decimal.NewFromFloat(v).StringFixed(places)`; create `pkg/moneyutils/moneyutils_test.go` with table cases `4150` → `4150.00`, `0` → `0.00`, `1500.55` → `1500.55`, `3` → `3.00`, `0.005` → `0.01` (2 places)
- [x] T002 [P] Create `internal/taxes/income_type.go`: `type IncomeType uint` with `incomeTypeUnknown` (0), `IncomeTypeCash` (1), `IncomeTypePOSTerminal` (2), `IncomeTypeNonCash` (3), `IncomeTypeOther` (4), `IncomeTypeUnpaidSupply` (5), `incomeTypeSentinel`; a descriptor table (no stringer, research R1) with name / field / description exactly: `Cash` / 18 / `cash payments (cash register turnover)`, `POS terminal` / 19 / `card payments through a POS terminal`, `Non-cash` / 20 / `bank transfers and other non-cash payments, except POS terminals`, `Other` / 21 / `services not taxed at source, barter, and other income`, `Unpaid supply` / 22 / `goods supplied or services provided but not paid for in this month`; methods `String()`, `Field() int`, `Label()` (`"<name> (<field>)"`, e.g. `Cash (18)`), `Description()`, `Valid()` ("true for values 1–5"); for invalid values (0, ≥ 6) `String()` returns `IncomeType(<n>)`, `Field()` returns 0, `Label()` and `Description()` return `""`; `AllIncomeTypes()` returning the five valid types in field order; `ErrInvalidIncomeType = errors.New("invalid income type")`; `ParseIncomeType(raw string) (IncomeType, error)` accepting the names case-insensitively after trimming spaces, otherwise `fmt.Errorf("%q: %w", raw, ErrInvalidIncomeType)`. Add table tests in `internal/taxes/income_type_test.go` (names, fields, labels, order, parsing of `non-cash` and ` Cash `, rejection of `""` and `Card` with `errors.Is`)
- [x] T003 [P] Add `func (i TaxType) IsSmallBusiness() bool` (doc comment: the categories declared on the Small Business monthly declaration) to `internal/taxes/calculator.go`, true for `TaxTypeSmallBusiness` and `TaxTypeSmallBusinessOverLimit`; table test in `internal/taxes/calculator_test.go` (both true, `TaxTypeEmployment` and the zero value false)
- [x] T004 In `internal/service/service.go` add `Type string` to `Income` (doc: income type name, "required for the Small Business categories, ignored for Employment"); add `IncomeResponse` embedding `ConvertResponse` with `Type taxes.IncomeType` and a `String()` that returns the `ConvertResponse` lines plus `\nType: <Label>` when `Type.Valid()`; change `CalculateResponse.Incomes` to `[]IncomeResponse` and build it in `Calculate` (type left zero for now); update the expected `Incomes` in `internal/service/service_test.go` (cases at the `Incomes: []ConvertResponse{` literals and `TestCalculateResponse_String`). Depends on T002

**Checkpoint**: `go test -race ./pkg/moneyutils/... ./internal/... ./cmd/...` passes; output is unchanged.

---

## Phase 3: User Story 2 - Choose the income type for each income (Priority: P1)

**Goal**: For the Small Business categories, both interfaces ask for the type after the currency, show it in lists and the review, and the service rejects a missing or unknown type.

**Independent Test**: In either interface choose Small Business, enter an income up to the currency, and check the five options `Cash (18)` … `Unpaid supply (22)`; finish two incomes with different types and check the review; a service request with a Small Business income without a type fails with `income 1: "": invalid income type` and makes no conversion.

- [x] T005 [US2] Create `internal/service/declaration.go` with `checkIncomeTypes(tt taxes.TaxType, incomes []Income) ([]taxes.IncomeType, error)`: for `tt.IsSmallBusiness()` parse every `Income.Type` with `taxes.ParseIncomeType` and wrap failures as `fmt.Errorf("income %d: %w", i+1, err)` (text `income 2: "Card": invalid income type`); for other categories return zero types without parsing (Employment ignores stray types). In `Calculate` in `internal/service/service.go` call it after `checkSingleYear` and before any conversion, wrapped `failed to check incomes: %w`, and set `IncomeResponse.Type` from the result. Depends on T003, T004
- [x] T006 [P] [US2] Tests in `internal/service/declaration_test.go` (new): missing type and unknown type at position 2 (exact text, `errors.Is(err, taxes.ErrInvalidIncomeType)`), lowercase `non-cash` accepted, Employment with `Card` accepted and `IncomeResponse.Type` zero, and a Small Business request with an invalid type using a converter that fails the test if called; give existing Small Business requests a type so they keep testing what they tested: the `converter returns error` case in `internal/service/service_test.go` (`Type: taxes.IncomeTypeNonCash.String()`) and the `gel()` helper in `TestCalculateYearIncomeRules` in `internal/service/year_income_test.go`. Depends on T005
- [x] T007 [P] [US2] CLI: in `cmd/ge-tax-calc/prompts.go` add `incomeTypeOptions() []option` from `taxes.AllIncomeTypes()` (`Label: t.Label()`, `Value: t.String()`, `Description: t.Description()`); in `cmd/ge-tax-calc/tax_menus.go` change `newIncomeModel()` to `newIncomeModel(taxType string)` storing whether `taxes.ParseTaxType(taxType)` is a Small Business category (parse error → `m.err`), change `runIncomeMenu()` to `runIncomeMenu(taxType string)` and pass `req.TaxType` in `menuCalcTaxes` in `cmd/ge-tax-calc/menu.go`; add `incomeStepType` after `incomeStepCurrency`: for Small Business the currency step opens `newSelectPrompt("Select income type (declaration field)", incomeTypeOptions(), <previous income's Type or "">)` and the type step sets `m.current.Type` before appending; Employment goes from currency to "Add another income entry?" as today; `renderIncomeList` appends ` — <Label>` and `renderIncomeDetails` adds `   Type: <Label>` when the income has a type (contract "Income lists and review"); extract helpers so `advance` stays within funlen. Tests in `cmd/ge-tax-calc/tax_menus_test.go`: existing tests use `newIncomeModel(taxes.TaxTypeEmployment.String())` unchanged; new Small Business test checks the five options in field order, picks `Non-cash`, checks the second income preselects `Non-cash`, the list line `— Non-cash (20)`, the review line `Type: Non-cash (20)`, and `incomes[i].Type`; Employment test checks there is no type step. Depends on T002
- [x] T008 [P] [US2] Bot: in `cmd/bot/session.go` add `calcStepIncomeType` between `calcStepCurrency` and `calcStepAddMore`, `stepCalcIncomeType = "calc_inc_type"`, and its `calcStepIDs` entry; in `cmd/bot/keyboards.go` add `isSmallBusiness(taxType string) bool` (via `taxes.ParseTaxType`) and `incomeTypeOptions()` (`text: t.Label()`, `value: t.String()`), returned for `stepCalcIncomeType` with one button per row; in `cmd/bot/handlers.go` at `calcStepCurrency` for Small Business move to `calcStepIncomeType` and send `✅ Currency: <currency>` / empty line / `🧾 Select the income type (declaration field):` / `• <Label> — <Description>` per type; add the `calcStepIncomeType` case that sets `currentInc.Type`, appends the income, and sends `✅ Income type: <Label>` / list / `➕ Add another income entry?`; Employment keeps the current currency path; `formatIncomeList` and `formatCalcSummary` append ` — <Label>` when the income has a type; extract an append-income helper so `handleCalcCallback` stays within funlen. Tests: `cmd/bot/keyboards_test.go` (five buttons, one per row, `cb:calc_inc_type:Unpaid supply` ≤ 64 bytes, unknown type and type buttons at another step rejected by `callbackValue`); `cmd/bot/flows_test.go` `TestBotCalculateFlow` presses `cb:calc_inc_type:Non-cash` after each currency, asserts the type message lists five types, the review shows `— Non-cash (20)`, text at the type step gets "Unexpected input", and the request carries `Type: "Non-cash"`; the Employment flow in `TestBotCalculateFailedAndPanicked` has no type step. Depends on T002, T003

**Checkpoint**: Types are collected and validated in both interfaces; output has no declaration block yet.

---

## Phase 4: User Story 1 - See which declaration field each amount goes to (Priority: P1) 🎯 MVP (with US2)

**Goal**: For the Small Business categories, the result lists fields (15)–(22) with two decimals; the warning gains the recalculation line; the year-income prompt asks for field (15) of the previous declaration; the bot values copy with a tap.

**Independent Test**: Small Business, year income `31792`, 3950 GEL Cash and 200 GEL Non-cash in one month → block (15) 35942.00, (16) 1.00, (17) 4150.00, (18) 3950.00, (19) 0.00, (20) 200.00, (21) 0.00, (22) 0.00, tax 41.5 GEL, identical in the CLI and the bot (bot values as `code` entities).

- [x] T009 [US1] In `internal/service/declaration.go` add `DeclarationField` (`Number int`, `Label string`, `Value float64`, `Formatted string`; `String()` = `(<Number>) <Label>: <Formatted>`), `Declaration` (`Fields []DeclarationField`, "fields 15–22, in this order, all present"; `String()` = `Declaration fields (Small Business monthly declaration):` then one line per field, joined with `\n`), and `newDeclaration(tr taxes.TaxRate, yearIncome, total float64, incomes []IncomeResponse) *Declaration` with labels exactly `Income since the start of the year`, `Tax rate, %`, `Taxable income for the month`, then the income type names; values: (15) `yearIncome`, (16) `moneyutils.Multiply(tr.Rate, 100)`, (17) `total`, (18)–(22) per-type sums of `Converted.Amount` with `moneyutils.Add` ("0 when none"); `Formatted` = `moneyutils.ToStringFixed(value, 2)`. In `internal/service/service.go` add `Declaration *Declaration` to `CalculateResponse`, set it in `Calculate` only when `tt.IsSmallBusiness()`, and in `CalculateResponse.String()` append `\n\n` + `Declaration.String()` after the `Taxes:` line and before the warning. Depends on T001, T005
- [x] T010 [P] [US1] Tests in `internal/service/declaration_test.go`: spec US1.1 (exact eight `Formatted` values, tax 41.5 GEL), US1.2 (two Non-cash incomes, `1000` USD and `500.55` EUR, which `mockConverter` converts 1:1 → (17) = (20) = `1500.55`, others `0.00`; covers "several currencies in one type"), US1.3 (over limit → (16) `3.00`), the invariant "field 17 = sum of fields 18–22 = `TotalIncomeConverted`; field 15 = `YearIncome`", Employment → `Declaration == nil`; add a `TestCalculateResponse_String` case in `internal/service/service_test.go` equal to the CLI example in `contracts/calculation-output.md`. Depends on T009
- [x] T011 [US1] In `internal/service/year_income.go` append to `LimitWarning.String()` the line `\nFor this month's declaration, recalculate with <TaxTypeSmallBusinessOverLimit.String()>: field (16) is <rate>.` where `<rate>` is `moneyutils.ToStringFixed(moneyutils.Multiply(rate, 100), 2)` of the over-limit rate from the rate table (`3.00` today); update `wantWarning` in `internal/service/year_income_test.go` and keep `TestCalculateYearIncomeRules` asserting the warning is the end of `String()`. Depends on T001
- [x] T012 [P] [US1] CLI year-income prompt in `taxDetailsModel.advance` in `cmd/ge-tax-calc/tax_menus.go`: for a Small Business category use the title `Field (15) of your previous month's declaration: income since the start of the year (GEL)` and `SetNote("Enter 0 for your first declaration of the year.")`; Employment keeps `Income from the beginning of a calendar year (GEL)`; tests in `cmd/ge-tax-calc/tax_menus_test.go` for both. (The CLI result already prints `CalculateResponse.String()` from `cmd/ge-tax-calc/menu.go`.) Depends on T003
- [x] T013 [P] [US1] Bot: in `cmd/bot/handlers.go` at `calcStepTaxType` send, for a Small Business category, `✅ Tax type: <name>` / empty line / `💰 Enter field (15) of your previous month's declaration: income since the start of the year in GEL:` / `(0.00 for your first declaration of the year)`, Employment unchanged; change `formatCalcResult` to return `(string, []telego.MessageEntity)`: income lines append ` — <Label>` when `inc.Type.Valid()`, after `Taxes to Pay: …` append `\n\n`, the header, and each `f.String()` line, adding one `telego.MessageEntity{Type: telego.EntityTypeCode, Offset, Length}` per value (the `Formatted` suffix of the line) with offset and length in UTF-16 code units, then the warning as today; add `utf16Len(s string) int` (with `unicode/utf16`) in `cmd/bot/send_message.go`; `runCalculation` sets `Entities` and no `ParseMode`. Tests: `cmd/bot/handlers_test.go` — for a response with the US1.1 declaration the eight entity substrings (decoded via UTF-16, with emoji before them) equal the eight values, Employment has no entities and no block, `TestFormatCalcResultLimitWarning` still ends with the warning; `cmd/bot/flows_test.go` — the Small Business year-income text asks for field (15) and the result call records eight `entities` of type `code`. Depends on T008, T009

**Checkpoint**: MVP — US1 + US2 work in both interfaces (quickstart CLI scenarios 1, 2, 4, 6; bot scenario 1).

---

## Phase 5: User Story 3 - Keep one Small Business calculation within one month (Priority: P1)

**Goal**: After the first income of a Small Business calculation only its month can be chosen; the service rejects Small Business incomes from two months; Employment keeps the one-year rule.

**Independent Test**: Small Business, first income 2026-10-05 → the second income offers only 2026 and October in both interfaces; a service request with September and October fails with `September and October: incomes span more than one calendar month`; Employment allows both months.

- [x] T014 [US3] In `internal/service/year_income.go` add `ErrIncomesSpanMonths = errors.New("incomes span more than one calendar month")` and `checkSingleMonth(incomes []Income) error` comparing months parsed with `dateutils.ParseMonth`, returning `fmt.Errorf("%s and %s: %w", first, month, ErrIncomesSpanMonths)` with `time.Month` names; in `Calculate` in `internal/service/service.go` call it between `checkSingleYear` and `checkIncomeTypes` (order of research R4) only when `tt.IsSmallBusiness()`, before any conversion, wrapped `failed to check incomes: %w`. Depends on T009
- [x] T015 [P] [US3] Tests in `internal/service/year_income_test.go`: no incomes, one month, September and October (exact text, `errors.Is`), invalid month → error; `Calculate` with Small Business and two months fails without calling the converter; Employment with two months of one year succeeds. Depends on T014
- [x] T016 [P] [US3] CLI: add `monthPrompt()` to `incomeModel` in `cmd/ge-tax-calc/tax_menus.go` (pattern of `yearPrompt`): for a Small Business category with captured incomes, a select prompt `Select month of income` with only `incomes[0].Month`, preselected; otherwise `monthOptions(year)` and `defaultMonthValue(year)`; use it at `incomeStepYear`. Tests in `cmd/ge-tax-calc/tax_menus_test.go`: Small Business second income offers one month equal to the first; review "No" restart offers all months again; Employment second income offers all months. Depends on T007
- [x] T017 [P] [US3] Bot: in `stepOptions` for `stepCalcMonth` in `cmd/bot/keyboards.go` return only `sess.incomes[0].Month` when `isSmallBusiness(sess.calcReq.TaxType)` and incomes exist. Tests: `cmd/bot/keyboards_test.go` (one month after the first Small Business income, another month's callback rejected, Employment offers all months); `cmd/bot/flows_test.go` (after "add another" the month keyboard has one button, an older month button gets "Unexpected input", review "No" offers all months, and `/cancel` at the type step followed by `/calculate` offers all months again). Depends on T008

**Checkpoint**: US3 works in both interfaces and in the service (quickstart CLI scenario 3; bot scenarios 2–3).

---

## Phase 6: User Story 4 - Employment calculation is unchanged (Priority: P2)

**Goal**: Employment steps and output stay exactly as before.

**Independent Test**: An Employment calculation shows the old year-income prompt, no type step, no month limit, and a result byte-identical to the pre-feature format.

- [x] T018 [P] [US4] Regression test in `internal/service/service_test.go`: an Employment `Calculate` with two incomes in different months returns `Declaration == nil`, zero `IncomeResponse.Type`, and a `String()` with no `Type:` line and no `Declaration fields` block, matching the pre-feature text. Depends on T009, T014
- [x] T019 [P] [US4] Regression test in `cmd/bot/handlers_test.go`: `formatCalcResult` for an Employment response equals a literal string in the pre-feature format and returns no entities. Depends on T013

**Checkpoint**: All stories done.

---

## Phase N: Polish & Cross-Cutting Concerns

- [x] T020 [P] Update the "Tax calculation rules" in the Usage section of `README.md` (FR-011): the five income types with field numbers, the one-month rule for Small Business, the year-income prompt (field (15) of the previous month's declaration), the declaration block with an example, the bot's tap-to-copy values, the recalculation line in the limit warning; Employment unchanged
- [x] T021 [P] Add notes to `specs/006-year-income-rules/spec.md` (FR-001 text extended by 007 FR-012; FR-005 narrowed to one month for Small Business by 007 FR-010) and to `specs/006-year-income-rules/contracts/calculation-output.md` (warning gains the 007 line)
- [x] T022 Run `make format-project` (or `gofmt -w` and `goimports -w` on changed files)
- [x] T023 Run `go vet ./...`, `go test -race ./...`, and `make lint-pipeline`; fix funlen, gocyclo, lll, mnd (use named constants for 100, 2, 15–22), wsl_v5, and exhaustive findings on changed lines
- [x] T024 Run `make build` (or `go build ./...` without Docker)
- [x] T025 Run the quickstart scenarios in `specs/007-income-type-declaration-fields/quickstart.md` (CLI 1–6, bot 1–5 where a test bot is available); record what was run, then set **Status** in `specs/007-income-type-declaration-fields/spec.md` to Implemented with the date
- [x] T026 Amend `.specify/memory/constitution.md` through `/speckit-constitution` (PATCH 1.0.6 → 1.0.7, factual correction): principle VI no longer lists `pkg/moneyutils` as untested, since T001 adds `pkg/moneyutils/moneyutils_test.go`; update the Sync Impact Report (analysis finding C1). Depends on T001

---

## Phase F: Code review fixes (2026-11-15)

- [x] T027 Reject amounts that overflow float64 (e.g. `1e400`) in `parseMoney` in `internal/service/amount.go` with `ErrInvalidAmount` (`"<raw>" is too large`), so `Calculate` cannot panic in `moneyutils`; tests in `internal/service/amount_test.go`
- [x] T028 Round the reported year income to 2 decimals in `Calculate` in `internal/service/service.go`, so the Year Income line, the limit check, and field (15) agree (FR-005); tests in `internal/service/declaration_test.go`
- [x] T029 Accept bot callbacks only from the message with the current keyboard (`session.keyboardMsgID`, set in `sendStep`, checked in `handleCallback` in `cmd/bot/handlers.go`), so a type button of an earlier income is rejected (US2.4); the fake Telegram API returns distinct message IDs and `press` uses the latest keyboard; note in `specs/004-telegram-bot-platform/spec.md` FR-015 and `README.md`
- [x] T030 Replace `DeclarationField.Formatted` with `Formatted()` and add `Prefix()` in `internal/service/declaration.go`; the bot builds lines from them with a running UTF-16 offset in `cmd/bot/handlers.go`
- [x] T031 Shorten the Non-cash and Unpaid supply descriptions so each CLI option fits 80 columns (test in `cmd/ge-tax-calc/tax_menus_test.go`); add `taxes.IncomeTypeLabel` and `TaxRate.Percent` and use them in both interfaces and the service; `ParseIncomeType` without allocation; doc comments on the new `String()` methods; update `data-model.md`, the contract, and `README.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 2** (T001–T004) → all stories. T001, T002, T003 in parallel; T004 after T002.
- **US2** (Phase 3): T005 after T003/T004; T006 after T005; T007 after T002; T008 after T002/T003. T007 and T008 can start before T005 (interface files only).
- **US1** (Phase 4): T009 after T001/T005 (same `Calculate`); T010 after T009; T011 after T001 (only `year_income.go`); T012 after T003; T013 after T008 (same handlers) and T009.
- **US3** (Phase 5): T014 after T009 (same `Calculate`); T015 after T014; T016 after T007 (same model); T017 after T008 (same keyboards file).
- **US4** (Phase 6): T018 after T009/T014; T019 after T013.
- **Polish**: after the desired stories; T020 and T021 can start once the contract is final; T026 after T001 is merged.

### Same-file ordering

- `internal/service/service.go`: T004 → T005 → T009 → T014.
- `internal/service/year_income.go`: T011 and T014 (either order, not at once).
- `cmd/ge-tax-calc/tax_menus.go`: T007 → T012 → T016.
- `cmd/bot/handlers.go`: T008 → T013; `cmd/bot/keyboards.go`: T008 → T017.

Story completion order: US2 → US1 (MVP) → US3 → US4.

## Parallel Examples

- Phase 2: T001 (`pkg/moneyutils`), T002 (`internal/taxes/income_type.go`), T003 (`internal/taxes/calculator.go`) together.
- US2: after T002/T003, T007 (`cmd/ge-tax-calc`) and T008 (`cmd/bot`) together while T005 → T006 run in `internal/service`.
- US1: after T009, T010 (service tests), T011 (`year_income.go`), T012 (CLI), and T013 (bot) together.
- US3: after T014, T015 (service tests), T016 (CLI), T017 (bot) together.

## Implementation Strategy

- **MVP**: US2 + US1 — the user picks the type and gets the eight field values to copy into the declaration (spec SC-001).
- Then US3, which keeps fields (16)–(17) tied to one reporting month; until it lands, a Small Business calculation with several months produces a block summing them, so ship US3 in the same release as the MVP.
- US4 adds regression tests only.
- Run `go test -race ./...` after each phase.

## Notes

- Do not hand-edit generated files or `vendor/`; this feature changes neither.
- Every new exported identifier gets a doc comment (constitution, Naming and Documentation): `IncomeType` and its constants, `ErrInvalidIncomeType`, `AllIncomeTypes`, `ParseIncomeType`, `IncomeResponse`, `Declaration`, `DeclarationField`, `ErrIncomesSpanMonths`, `ToStringFixed`, `IsSmallBusiness`.
- Commit with Conventional Commits (`feat:`, `test:`, `docs:`).
- Mind `exhaustive` on the `calcStep` switches in `cmd/bot` when adding `calcStepIncomeType`.
