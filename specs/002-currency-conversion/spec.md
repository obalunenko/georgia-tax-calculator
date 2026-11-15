# Feature Specification: Currency Conversion

**Feature Directory**: `specs/002-currency-conversion`

**Feature Branch**: N/A (migrated from `master`; implemented before Spec Kit adoption)

**Created**: 2026-11-14

**Status**: migrated

**Input**: Reverse-engineered from `internal/converter/`, `internal/service/`
(`Convert`), `cmd/ge-tax-calc/convert_menu.go`, and the `/convert` flow in
`cmd/bot/`. Related history: `2d3a270` (rates in convert response, #81),
`005b6fa` (USD default in menu, #195), `79b82c2` (Bubble Tea menu, #196),
`0a76d1d` (Telegram bot, #221).

## Project Scope and Compatibility

- **Affected interfaces**: Both. CLI command `ge-tax-calc convert` and bot
  command `/convert`.
- **Affected boundaries**: `internal/converter/` (cross-rate calculation),
  `internal/service/` (`Convert`, `ConvertRequest`, `ConvertResponse`),
  `internal/models/` (`Money`), `pkg/moneyutils/` (decimal arithmetic),
  `pkg/dateutils/` (date parsing and date options), `cmd/ge-tax-calc/`
  (`convert_menu.go`, `prompts.go`, `menu.go`), `cmd/bot/` (`handlers.go`,
  `session.go`, `keyboards.go`). Depends on
  [001-nbg-rates-client](../001-nbg-rates-client/spec.md).
- **Tax categories and rates**: N/A.
- **Currencies and rate dates**: Any pair of codes from `currencies.All()`,
  including GEL on either side. The NBG rate for the selected date is used.
- **User-facing contract**: CLI `convert` command (interactive form); bot
  `/convert` (inline keyboards plus one text input). Output fields: date,
  amount, converted amount, rate.
- **Release impact**: None beyond both binaries.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Convert an amount at the official rate for a date (Priority: P1)

A user converts an amount from one currency to another at the NBG official
rate for a chosen date and sees the converted amount and the rate used.

**Why this priority**: It is the core calculation reused by tax calculation
(conversion of each income to GEL).

**Independent Test**: `TestConverter_Convert` in
`internal/converter/converter_test.go` with the
`internal/converter/testdata/2023-09-09-all.json` fixture, and
`Test_service_Convert` in `internal/service/service_test.go`.

**Acceptance Scenarios**:

1. **Given** NBG rates from the fixture and 2678.27 EUR, **When** it is
   converted to GEL, **Then** the result is 7521.39 GEL at rate 2.8083.
2. **Given** 2678.27 BYN, **When** it is converted to GEL, **Then** the result
   is 2792.1 GEL at rate 1.0425.
3. **Given** two non-GEL currencies (e.g., EUR → GBP, RUB → TMT), **When**
   converted, **Then** the cross rate is computed through GEL using each
   currency's rate per unit (`Rate / Quantity`).
4. **Given** the same currency on both sides (EUR → EUR, GEL → GEL), **When**
   converted, **Then** the amount is unchanged and the rate is 1.
5. **Given** an empty source or target currency, **When** converting, **Then**
   `ErrCurrencyNotSet` is returned with `from:` or `to:` context.
6. **Given** year "2022", month "December", day "08", amount "568",
   **When** the service converts, **Then** the response date is 2022-12-08 UTC
   and the amount, converted amount, and rate are returned.

---

### User Story 2 - Convert in the CLI (Priority: P2)

A user runs `ge-tax-calc convert` and fills an interactive form.

**Why this priority**: The CLI is the primary interface (README "Usage").

**Independent Test**: `cmd/ge-tax-calc/convert_menu_test.go`.

**Acceptance Scenarios**:

1. **Given** the form, **When** the user selects year, month, day, enters an
   amount, selects source currency (default USD) and target currency
   (default GEL), and confirms, **Then** the request is sent to the service
   and the result is printed.
2. **Given** the review step, **When** the user answers "No", **Then** the
   form restarts from the year with empty input.
3. **Given** a partially filled form, **When** a later step is shown, **Then**
   the current input summary is displayed above the prompt.
4. **Given** Ctrl+C at any step, **When** pressed, **Then** the command exits
   with "failed to collect converter input: input aborted".
5. **Given** the amount step, **When** the user enters `12abc`, `1,500`,
   `NaN`, `0`, or `-5`, **Then** the prompt shows a validation error and
   does not advance (FR-014).

---

### User Story 3 - Convert in the Telegram bot (Priority: P2)

A user sends `/convert` and follows inline keyboards.

**Why this priority**: Parity with the CLI (constitution II).

**Independent Test**: Manual check in a test bot; no automated tests exist.

**Acceptance Scenarios**:

1. **Given** `/convert`, **When** the user picks year, month, and day, types an
   amount, picks source and target currencies, and confirms, **Then** the bot
   replies "⏳ Converting..." and then the conversion result.
2. **Given** the review step, **When** the user picks "No", **Then** the flow
   restarts from the year.
3. **Given** the amount step, **When** the user sends text that cannot be
   parsed as a number, **Then** the bot asks again with "❌ Invalid amount".
4. **Given** a conversion error, **When** the service fails, **Then** the bot
   replies "❌ Conversion error: ..." and suggests `/convert`.
5. **Given** text at a step that expects a button, **When** it is sent,
   **Then** the bot replies "⚠️ Unexpected input".
6. **Given** the amount step, **When** the user sends `12abc`, `1,500`,
   `NaN`, `0`, or `-5`, **Then** the bot asks again with "❌ Invalid amount"
   and stays on the amount step (FR-014).
7. **Given** a button from an earlier step or keyboard, **When** it is
   pressed, **Then** the bot replies "⚠️ Unexpected input" and the step does
   not change (004 FR-015).

---

### Edge Cases

- Selectable years are 2016 through the current year; months and days are
  limited to dates up to today in both interfaces.
- ~~"Today" is the local date of the process (`time.Now()`). In the Docker
  image this is UTC, so from 00:00 to 04:00 Tbilisi time (UTC+4) the current
  Georgian date cannot be selected.~~ Since 2026-11-15 "today" is the date in
  Georgia (UTC+4) in both interfaces and the service (FR-016).
- ~~The service itself does not reject future dates, day 0, or day 31 in a
  30-day month; `time.Date` normalizes such dates.~~ Since 2026-11-15 the
  service rejects them, and dates before 2016 (FR-015).
- ~~Negative and zero amounts are accepted by both interfaces.~~ Superseded
  by BUG-001: amounts ≤ 0 MUST be rejected by the service and both
  interfaces (FR-014).
- ~~The bot accepts amounts like `12abc`, `1,500`, `NaN`, or `Inf` at input
  time (`fmt.Sscanf`) and fails only at conversion; the CLI rejects them
  immediately.~~ Superseded by BUG-001: both interfaces reject such input at
  the amount step (FR-014).
- A currency missing from the NBG response for the date returns
  `ErrCodeNotFound`.
- NBG errors and timeouts surface as conversion errors; no retry.
- ~~Stale inline buttons in the bot are applied to the current step; at the
  confirmation step any button other than "No" starts the conversion.~~
  Superseded by 004 BUG-001: buttons from another step or with an option the
  current step does not offer are rejected (004 FR-015).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST convert an amount between any two codes from
  `currencies.All()` using NBG rates for the selected date.
- **FR-002**: The rate MUST be `(from.Rate / from.Quantity) / (to.Rate /
  to.Quantity)`, computed with `pkg/moneyutils`.
- **FR-003**: The converted amount MUST be rounded to 2 decimal places and the
  rate to 4.
- **FR-004**: Empty source or target currency MUST return `ErrCurrencyNotSet`.
- **FR-005**: The service MUST accept the date as year, month name, and day
  strings and build a UTC date, validated by FR-015.
- **FR-006**: The service MUST parse the amount with `moneyutils.Parse` and
  return an error for invalid input, using the rule in FR-014.
- **FR-007**: The response MUST contain the date, the source amount, the
  converted amount, and the rate; its text form prints `Date`, `Amount`,
  `Converted`, `Rate`.
- **FR-008**: The CLI MUST offer years from 2016 to the current year, months
  and days up to today, currencies from `currencies.All()` (sorted), defaults
  USD → GEL, and a confirmation step that restarts the form on "No".
- **FR-009**: The bot MUST offer the same date and currency choices through
  inline keyboards, request the amount as text, and confirm before converting.
- **FR-010**: The CLI MUST validate the amount with ~~`moneyutils.Parse`~~
  the rule in FR-014 before accepting it.
- **FR-011**: The bot MUST validate the amount with the same rule as the CLI
  (FR-014) before accepting it. ~~[Known defect: the bot uses
  `fmt.Sscanf("%f")`, which accepts `12abc`, `1,500`, `NaN`, and `Inf`; see
  tasks.md gap 1 and BUG-001.]~~ Fixed on 2026-11-15 (BUG-001, T020).
- **FR-012**: Conversion errors MUST be shown to the user (CLI returns the
  error; the bot replies with the error and suggests `/convert`).
- **FR-013**: `/cancel` MUST abort an in-progress bot conversion (see
  [004-telegram-bot-platform](../004-telegram-bot-platform/spec.md)).
- **FR-014**: One amount rule in `internal/service` MUST be used by the
  service, the CLI, and the bot: trim spaces, parse with `moneyutils.Parse`,
  and require a value greater than 0.
- **FR-015**: The service MUST reject, with `ErrInvalidDate`, a date that does
  not exist (day 0, 31 April, 29 February in a common year), is before
  `service.FirstYear` (2016), or is after today in Georgia. (Added
  2026-11-15, T022.)
- **FR-016**: "Today" for year, month, and day options in both interfaces and
  for FR-015 MUST be the date in Georgia (`service.Now()`, UTC+4), and both
  interfaces MUST take the first selectable year from `service.FirstYear`.
  (Added 2026-11-15, T023.)

**Bugfix**: 2026-11-15 — BUG-001 Added FR-014 (shared amount rule, amount
greater than 0), updated FR-006, FR-010, FR-011, Edge Cases, and acceptance
scenarios US2.5, US3.6, US3.7.

### Key Entities

- **Money**: Amount (`float64`) and currency code; text form
  `<amount> <code>`.
- **ConvertRequest**: Year, month name, day, source and target currency,
  amount (all strings, as collected by the interfaces).
- **ConvertResponse**: Date, source amount, converted amount, rate (as
  `Money` without currency).
- **converter.Response**: Converted `Money` and rate.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For the fixture rates, converted amounts and rates match the
  expected values in `internal/converter/converter_test.go` for 13 currency
  pairs.
- **SC-002**: The CLI and the bot produce the same result for the same date,
  amount, and currencies.
- **SC-003**: A user completes a conversion in either interface in at most
  seven inputs (year, month, day, amount, two currencies, confirmation).

## Assumptions

- NBG publishes a rate for the selected date for both currencies.
- `float64` amounts with decimal arithmetic inside `moneyutils` give
  sufficient precision for declaration amounts.
- Rates before 2016 are not needed.
