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

- [x] T005 [P] [US1] Table-driven conversion tests for 13 currency pairs and empty source/target currency in `internal/converter/converter_test.go`
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
- [x] T016 [US3] Implement `/convert`, amount text input with `validateMoney`, step callbacks, confirmation, and result formatting in `cmd/bot/handlers.go`

**Checkpoint**: `/convert` works end to end in the bot.

---

## Phase N: Polish & Cross-Cutting Concerns

- [x] T017 [P] Document the `convert` command and bot `/convert` in `README.md`

---

## Gaps Found During Migration

Not tasks of this feature. Track them with `/speckit-bugfix-report` or a new
`/speckit-specify` feature.

1. **Parity defect in amount validation.** The bot's `parseMoneyAmount`
   (`cmd/bot/handlers.go:697`) uses `fmt.Sscanf("%f")`, which accepts `12abc`
   (as 12) and `1,500` (as 1). The raw text is stored and fails later in
   `service.Convert` with `can't convert … to decimal`, after the user has
   confirmed. The CLI rejects such input immediately with `moneyutils.Parse`.
2. **Negative and zero amounts** are accepted by both interfaces and the
   service.
3. **Date validation lives only in the interfaces.** `dateutils.ParseDay`
   accepts 0 (its message says "should be >0") and does not check the number
   of days in the month; `service.Convert` passes such values to `time.Date`,
   which normalizes them (e.g., February 31 becomes March 2 or 3). Future
   dates are not rejected by the service.
4. **Stale inline buttons in the bot.** `handleConvertCallback` applies the
   callback data to the current step without checking which keyboard it came
   from, so pressing an old button can store a year as a currency.
5. **No tests for the bot `/convert` flow or for `pkg/moneyutils`.**
6. **Converter tests do not check the requested date.** `mock.NewClient`
   ignores the date (see 001 gaps), so tests pass `time.Now()` and would not
   detect a wrong date being sent. `internal/converter/testdata/2022-11-25-all.json`
   is not used by any test.
7. **Spinner in the service.** `service.Convert` starts a terminal spinner,
   also for bot requests (constitution known deviation).
8. **Unused `survey:"…"` struct tags** on `DateRequest`, `ConvertRequest`, and
   `Income` remain from the survey library used before Bubble Tea.
9. **Unwrapped errors.** Date and amount parsing errors in `service.Convert`
   are returned without context; only converter errors are wrapped.
