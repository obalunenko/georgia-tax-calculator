---

description: "Migrated task list for currency conversion"
---

# Tasks: Currency Conversion

**Input**: Design documents from `/specs/002-currency-conversion/`

**Prerequisites**: plan.md, spec.md; depends on `specs/001-nbg-rates-client`

**Status**: migrated — all tasks reflect code that already exists on `master`.
Gaps found during migration are listed at the end without checkboxes; they are
candidates for follow-up features or bug reports, not pending work for this
feature.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)

## Phase 1: Feature Preparation

- [x] T001 Add NBG fixture `internal/converter/testdata/2023-09-09-all.json`

---

## Phase 2: Foundational (Blocking Prerequisites)

- [x] T002 [P] Define `Money` and `NewMoney` with text form in `internal/models/models.go` (tests in `internal/models/models_test.go`)
- [x] T003 [P] Implement decimal-backed `Parse`, `Multiply`, `Div`, `Add`, `Round`, `ToString` in `pkg/moneyutils/moneyutils.go`
- [x] T004 [P] Implement `ParseYear`, `ParseMonth`, `ParseDay`, and month/day lists up to a date in `pkg/dateutils/dateutils.go` (tests in `pkg/dateutils/dateutils_test.go`)

---

## Phase 3: User Story 1 - Convert an amount at the official rate for a date (Priority: P1) 🎯 MVP

### Tests for User Story 1

- [x] T005 [P] [US1] Table-driven conversion tests for 13 currency pairs and empty source/target currency in `internal/converter/converter_test.go`; since 2026-11-15 the tests use the fixture dates and also cover another date, a date without rates, and an unknown currency
- [x] T006 [P] [US1] Service tests (correct request, invalid request, converter error) and `ConvertResponse.String` test in `internal/service/service_test.go`

### Implementation for User Story 1

- [x] T007 [US1] Implement `Converter`, cross rate through per-unit rates, rounding, and `ErrCurrencyNotSet` in `internal/converter/converter.go`
- [x] T008 [US1] Define `DateRequest`, `ConvertRequest`, `ConvertResponse` and its `String` in `internal/service/service.go`
- [x] T009 [US1] Implement `service.Convert`: parse date and amount, call the converter, build the response (`internal/service/service.go`)

**Checkpoint**: Conversion works through the service.

---

## Phase 4: User Story 2 - Convert in the CLI (Priority: P2)

### Tests for User Story 2

- [x] T010 [P] [US2] Form tests (confirm, decline restarts, current input summary, prompt height) in `cmd/ge-tax-calc/convert_menu_test.go`

### Implementation for User Story 2

- [x] T011 [US2] Implement year, month, day, and currency options and `validateMoneyInput` in `cmd/ge-tax-calc/prompts.go`
- [x] T012 [US2] Implement the Bubble Tea conversion form with summary and confirmation in `cmd/ge-tax-calc/convert_menu.go`
- [x] T013 [US2] Register the `convert` command and print the result (`cmd/ge-tax-calc/commands.go`, `menuConvert` in `cmd/ge-tax-calc/menu.go`)

**Checkpoint**: `ge-tax-calc convert` works end to end.

---

## Phase 5: User Story 3 - Convert in the Telegram bot (Priority: P2)

### Implementation for User Story 3

- [x] T014 [P] [US3] Add `flowConvert`, `convertStep`, and `convertReq` to the session in `cmd/bot/session.go`
- [x] T015 [P] [US3] Build year, month, day, currency, and confirmation keyboards in `cmd/bot/keyboards.go`
- [x] T016 [US3] ⚠️ Reopened Implement `/convert`, amount text input with `validateMoney`, step callbacks, confirmation, and result formatting in `cmd/bot/handlers.go` (reopened — BUG-001: `validateMoney` did not meet FR-011; closed by T020 on 2026-11-15)

**Checkpoint**: `/convert` works end to end in the bot.

---

## Phase B: Bugfix BUG-001 - Shared amount rule

- [x] T018 [US1] Add `ErrInvalidAmount`, `ParseAmount` (> 0), and `ParseYearIncome` (≥ 0) in `internal/service/amount.go`; use `ParseAmount` in `Convert` (FR-006, FR-014)
- [x] T019 [P] [US1] Table-driven tests for `ParseAmount` and `ParseYearIncome` (`12abc`, `1,500`, `NaN`, `Inf`, `-5`, `0`, `0.00`, ` 12 `, `1500.50`) and service-level rejection in `internal/service/amount_test.go`
- [x] T020 [US3] Make the bot `validateMoney` call `service.ParseAmount` and remove `parseMoneyAmount` in `cmd/bot/handlers.go`; tests in `cmd/bot/handlers_test.go` (FR-011)
- [x] T021 [US2] Make the CLI `validateMoneyInput` call `service.ParseAmount` in `cmd/ge-tax-calc/prompts.go`; update `cmd/ge-tax-calc/prompts_test.go` (FR-010)

**Checkpoint**: The service, the CLI, and the bot reject the same amounts.

**Bugfix**: 2026-11-15 — BUG-001 Updated from bugfix patch

---

## Phase F: Follow-up for migration gaps 3 and 10 (2026-11-15)

- [x] T022 [US1] Add `FirstYear`, `ErrInvalidDate`, and `newDate` (rejects dates that do not exist, are before 2016, or are after today in Georgia) in `internal/service/date.go` and use it in `Convert` (`internal/service/service.go`); tests in `internal/service/date_test.go` (gap 3, FR-015)
- [x] T023 [US2] Add `service.Now()` (fixed UTC+4 zone, no tzdata needed in the Alpine image) and use it with `service.FirstYear` for the date options in `cmd/ge-tax-calc/prompts.go` and `cmd/bot/keyboards.go` (gap 10, FR-016)

---

## Phase N: Polish & Cross-Cutting Concerns

- [x] T017 [P] Document the `convert` command and bot `/convert` in `README.md`

---

## Gaps Found During Migration

Not tasks of this feature. Track them with `/speckit-bugfix-report` or a new
`/speckit-specify` feature.

1. **Fixed by T018–T021 (BUG-001).** **Parity defect in amount validation.** The bot's `parseMoneyAmount`
   (`cmd/bot/handlers.go:697`) uses `fmt.Sscanf("%f")`, which accepts `12abc`
   (as 12), `1,500` (as 1), `NaN`, and `Inf`. The raw text is stored and fails
   later in `service.Convert` with `can't convert … to decimal`, after the
   user has confirmed. The CLI rejects such input immediately with
   `moneyutils.Parse`. Canonical entry; 003 gap 1 points here.
2. **Negative and zero amounts** are accepted by both interfaces and the
   service. Decision (2026-11-15): income and conversion amounts MUST be
   greater than 0; year income in 003 MUST be 0 or greater (0 is valid for
   the first income of the year). Implemented by T018–T021 and 003 T017.
3. **Fixed by T022.** **Date validation lives only in the interfaces.** `dateutils.ParseDay`
   accepts 0 (its message says "should be >0") and does not check the number
   of days in the month; `service.Convert` passes such values to `time.Date`,
   which normalizes them (e.g., February 31 becomes March 2 or 3). Future
   dates are not rejected by the service.
4. **Stale inline buttons in the bot.** `handleConvertCallback` applies the
   callback data to the current step without checking which keyboard it came
   from, so pressing an old button can store a year as a currency. At
   `convertStepConfirm` (`cmd/bot/handlers.go:533`) any data other than `no`
   starts the conversion, so an old button runs it. The same applies to
   `/calculate` (003 gap 9). Decision (2026-11-15): callback data will carry
   the step (`cb:<step>:<value>`), and data from another step will be
   rejected. Implemented by 004 T016–T018.
5. **Partly fixed by T020 and 004 T018, T022.** **No tests for the bot `/convert` flow or for `pkg/moneyutils`.** As a
   result, SC-002 (same result in the CLI and the bot) is checked only by
   hand.
6. **Fixed (2026-11-15, 001 T020).** Converter tests did not check the
   requested date: `mock.NewClient` ignored it, tests passed `time.Now()`, and
   `internal/converter/testdata/2022-11-25-all.json` was unused. Now the mock
   selects the set by date, tests pass the fixture dates, and a case on
   2022-11-25 plus "no rates for the date" and "unknown currency" cases were
   added. A mutation check (sending the next day's date) fails 14 cases.
7. **Fixed (2026-11-15, 005 T018).** **Spinner in the service.** `service.Convert` starts a terminal spinner,
   also for bot requests. Tracked as a known deviation in
   `.specify/memory/constitution.md`; 003 refers here.
8. **Unused `survey:"…"` struct tags** on `DateRequest`, `ConvertRequest`, and
   `Income` remain from the survey library used before Bubble Tea.
9. **Unwrapped errors.** Date and amount parsing errors in `service.Convert`
   are returned without context; only converter errors are wrapped.
10. **Fixed by T023.** **No time zone for "today".** Year, month, and day options in both
    interfaces are limited by `time.Now()` in the process time zone
    (`cmd/ge-tax-calc/prompts.go`, `cmd/bot/keyboards.go`). The bot image
    runs in UTC, so from 00:00 to 04:00 Tbilisi time the current Georgian
    date is not offered. The spec does not say which time zone applies.
