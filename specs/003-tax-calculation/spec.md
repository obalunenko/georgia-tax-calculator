# Feature Specification: Tax Calculation

**Feature Directory**: `specs/003-tax-calculation`

**Feature Branch**: N/A (migrated from `master`; implemented before Spec Kit adoption)

**Created**: 2026-11-14

**Status**: migrated

**Input**: Reverse-engineered from `internal/taxes/`, `internal/service/`
(`Calculate`), `cmd/ge-tax-calc/tax_menus.go`, and the `/calculate` flow in
`cmd/bot/`. Related history: `58e9221` (multiple income response, #166),
`79b82c2` (Bubble Tea menu, #196), `0a76d1d` (Telegram bot, #221).

## Project Scope and Compatibility

- **Affected interfaces**: Both. CLI command `ge-tax-calc run` and bot
  command `/calculate`.
- **Affected boundaries**: `internal/taxes/` (`TaxType`, rates, `Calc`;
  generated `taxtype_string.go`), `internal/service/` (`Calculate`,
  `CalculateRequest`, `Income`, `CalculateResponse`), `cmd/ge-tax-calc/`
  (`tax_menus.go`, `menu.go`, `prompts.go`), `cmd/bot/` (`handlers.go`,
  `session.go`, `keyboards.go`). Each income is converted through
  [002-currency-conversion](../002-currency-conversion/spec.md).
- **Tax categories and rates**: Individual Entrepreneur 3%, Small Business
  1%, Employment 20%.
- **Currencies and rate dates**: Each income uses the NBG rate for its own
  date and is converted to GEL.
- **User-facing contract**: CLI `run` (two interactive forms: tax details,
  then incomes); bot `/calculate`. Output: tax rate, year income, each income
  with converted amount and rate, total converted income, tax.
- **Release impact**: Adding a category requires `make codegen` (stringer).

## Clarifications

### Session 2026-11-15

- Q: How must year income affect the tax for Small Business? → A: Only a
  warning: the tax is still calculated at the category rate, and a warning is
  shown when the annual turnover limit is exceeded.
- Q: Can one calculation mix incomes from different calendar years? → A: No,
  only one year: the year of the first income fixes the year of the
  calculation.
- Q: How must tax be rounded when there are several incomes? → A: Tax on the
  total: each converted income is rounded to 2 places, then the tax is
  `round(total × rate, 2)` once.
- Q: What should happen to the Individual Entrepreneur 3% category? → A: Keep
  it, and check all rates against the Tax Code before the next release,
  recording the source.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Calculate tax for incomes in any currency (Priority: P1)

A taxpayer selects a tax category, enters income received since the start of
the calendar year (in GEL), and lists incomes with date, amount, and currency.
The system converts each income to GEL at the NBG rate for its date and
calculates the tax.

**Why this priority**: It is the main purpose of the tool (preparing a
Georgian tax declaration).

**Independent Test**: `TestCalc`, `TestTaxType_Rate`, `TestAllTaxTypes` in
`internal/taxes/calculator_test.go`; `Test_service_Calculate` and
`TestCalculateResponse_String` in `internal/service/service_test.go`.

**Acceptance Scenarios**:

1. **Given** an income of 100278.88 GEL and category Small Business, **When**
   the tax is calculated, **Then** it is 1002.79 GEL at rate 1%.
2. **Given** categories Individual Entrepreneur and Employment, **When** their
   rates are requested, **Then** they are 3% and 20%.
3. **Given** category Employment, year income "67.99", and incomes 1000 EUR
   and 200 USD that convert to 1000 GEL and 200 GEL, **When** the service
   calculates, **Then** total converted income is 1200 GEL, tax is 240 GEL,
   and year income is 1267.99 GEL.
4. **Given** an unknown category name, **When** the service calculates,
   **Then** it fails with `failed to parse tax type: ...: invalid tax type`.
5. **Given** an income that cannot be converted, **When** the service
   calculates, **Then** it fails with `failed to convert income: ...` and no
   partial result is returned.
6. **Given** incomes dated in two different calendar years, **When** the
   service calculates, **Then** it fails with an error that the incomes span
   more than one calendar year (FR-014, not implemented yet).

---

### User Story 2 - Calculate in the CLI (Priority: P2)

A user runs `ge-tax-calc run` and fills two forms.

**Why this priority**: The CLI is the primary interface.

**Independent Test**: `cmd/ge-tax-calc/tax_menus_test.go`.

**Acceptance Scenarios**:

1. **Given** the tax details form, **When** the user selects a category
   (default Small Business; options show the rate) and enters income since the
   start of the year in GEL, **Then** the income form starts.
2. **Given** the income form, **When** the user enters year, month, day,
   amount, and currency (default USD) and answers "Yes" to "Add another
   income entry?", **Then** another income is collected.
3. **Given** the final review, **When** the user answers "No", **Then** income
   entry restarts with an empty list; the category and year income are kept.
4. **Given** a terminal that supports hyperlinks, **When** `run` starts,
   **Then** a "Declarations" link to `https://decl.rs.ge/decls.aspx` is
   printed.

---

### User Story 3 - Calculate in the Telegram bot (Priority: P2)

A user sends `/calculate` and follows the bot.

**Why this priority**: Parity with the CLI (constitution II).

**Independent Test**: Manual check in a test bot; no automated tests exist.

**Acceptance Scenarios**:

1. **Given** `/calculate`, **When** the user picks a category (buttons show
   the rate), types year income, then for each income picks year, month, day,
   types the amount, and picks the currency, **Then** the bot shows captured
   incomes and asks "➕ Add another income entry?".
2. **Given** the review step, **When** the user confirms, **Then** the bot
   replies "⏳ Calculating taxes..." and then the result with each income
   converted to GEL, total income, and taxes to pay.
3. **Given** the review step, **When** the user picks "No", **Then** the whole
   flow restarts from the category.
4. **Given** a calculation error, **When** the service fails, **Then** the bot
   replies "❌ Calculation error: ..." and suggests `/calculate`.
5. **Given** the year income step, **When** the user sends `0` or `0.00`,
   **Then** it is accepted; **When** the user sends `-1` or `12abc`, **Then**
   the bot asks again with "❌ Invalid amount" (FR-016).
6. **Given** the income amount step, **When** the user sends `0`, `-5`, or
   `12abc`, **Then** the bot asks again with "❌ Invalid amount" (FR-016).

---

### Edge Cases

- Year income is only added to the reported year income; it does not change
  the tax. For Small Business, exceeding the annual turnover limit only adds
  a warning (FR-013).
- Incomes from different calendar years MUST NOT be mixed in one
  calculation (FR-014); the current code still allows it.
- Tax is calculated once on the total converted income and rounded to 2
  decimal places (FR-004); the current code still rounds per income.
- A category string is matched case-insensitively and trimmed.
- Amount and date edge cases are inherited from
  [002-currency-conversion](../002-currency-conversion/spec.md) (~~negative
  amounts, bot `fmt.Sscanf` parsing,~~ date validation only in interfaces).
  Amounts follow FR-016 since BUG-001.
- ~~Stale inline buttons in the bot are applied to the current step; at the
  final review any button other than "No" starts the calculation.~~
  Superseded by 004 BUG-001: such buttons are rejected (004 FR-015).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST support the tax categories Individual
  Entrepreneur (3%), Small Business (1%), and Employment (20%).
- **FR-002**: Categories MUST be parsed from their display names,
  case-insensitively; unknown names MUST return `ErrInvalidTaxType`.
- **FR-003**: Each income MUST be converted to GEL at the NBG rate for its
  date using the currency conversion feature.
- **FR-004**: The tax MUST be `total converted income × rate`, computed with
  `pkg/moneyutils` and rounded once to 2 decimal places. [Current code
  rounds the tax per income and sums the results; not changed yet.]
- **FR-005**: Each converted income MUST be rounded to 2 decimal places
  (002 FR-003); the total converted income MUST be the sum of the rounded
  converted incomes.
- **FR-006**: The reported year income MUST be the entered year income plus
  the total converted income, in GEL.
- **FR-007**: Any parse, conversion, or calculation error MUST abort the
  calculation with a wrapped error.
- **FR-008**: The response MUST include the tax rate, year income, per-income
  conversion details (date, amount, converted amount, rate), total converted
  income, and tax.
- **FR-009**: The CLI MUST collect the category and year income, then one or
  more incomes with an "add another" loop and a final confirmation.
- **FR-010**: The bot MUST collect the same data in the same order through
  inline keyboards and text input, with the same confirmation.
- **FR-011**: Both interfaces MUST show each category with its rate.
- **FR-012**: Adding or changing a category MUST update `TaxType`, the rate
  table, `taxtype_string.go` (via `make codegen`), and tests together
  (constitution III).
- **FR-013**: For Small Business, when the reported year income (FR-006)
  exceeds the Small Business annual turnover limit, both interfaces MUST show
  a warning with the result. The tax MUST still be calculated at the category
  rate (FR-004). Year income MUST NOT change the rate or the category.
  [Not implemented yet.]
- **FR-014**: All incomes in one calculation MUST be in the same calendar
  year, fixed by the year of the first income. After the first income, both
  interfaces MUST offer only that year; the service MUST reject a request
  whose incomes span more than one year. [Not implemented yet.]
- **FR-015**: Before the next release, every rate in the rate table and the
  Small Business turnover limit (FR-013) MUST be checked against the current
  Tax Code of Georgia, and the source (article) and check date MUST be
  recorded in Assumptions. [Not done yet.]
- **FR-016**: Year income MUST be validated with `service.ParseYearIncome`
  (0 or greater) and each income amount with `service.ParseAmount` (greater
  than 0), in the service and in both interfaces (002 FR-014).

**Bugfix**: 2026-11-15 — 002 BUG-001 Added FR-016 (year income ≥ 0, income
amount greater than 0), acceptance scenarios US3.5 and US3.6, and updated
Edge Cases.

### Key Entities

- **TaxType**: Enum of categories with display names from stringer
  (`--trimprefix --linecomment`).
- **TaxRate**: Category and rate (`float64`); text form `<name> <percent> %`.
- **Income**: Date (year, month name, day), currency, amount (strings).
- **CalculateRequest**: Category name, year income (GEL string), incomes.
- **CalculateResponse**: Tax rate, year income, converted incomes, total
  converted income, tax (all `Money` in GEL except source amounts), and a
  turnover-limit warning when FR-013 applies.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For the test inputs, per-category tax and totals match
  `internal/taxes/calculator_test.go` and `internal/service/service_test.go`.
- **SC-002**: The CLI and the bot produce the same tax for the same inputs.
- **SC-003**: A user can calculate tax for several incomes in different
  currencies in one session in either interface.

## Assumptions

- The three categories and their rates reflect the rules the tool targets;
  rates are fixed in code. Their legal source is not recorded yet, and the
  Individual Entrepreneur 3% rate in particular is unverified (FR-015).
- The user enters year income already in GEL.
- The Small Business annual turnover limit is assumed to be 500 000 GEL; the
  value MUST be checked against the current Tax Code of Georgia before FR-013
  is implemented.
- Each income's tax is calculated independently of the others.
