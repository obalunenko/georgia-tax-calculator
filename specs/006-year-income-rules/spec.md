# Feature Specification: Year Income Rules for Tax Calculation

**Feature Directory**: `specs/006-year-income-rules`

**Feature Branch**: `fix/spec-gaps-bot-validation` (directory numbering does not create a branch)

**Created**: 2026-11-15

**Status**: Implemented (2026-11-15); FR-004 limit value pending 003 FR-015

**Input**: User description: "Year-income rules for tax calculation (follow-up
to specs/003-tax-calculation clarifications of 2026-11-15): Small Business
turnover-limit warning (003 FR-013), one calendar year per calculation (003
FR-014), tax on the total (003 FR-004), and renaming the Individual
Entrepreneur (3%) category to reflect the Small Business rate above the
limit. The legal source of rates and the limit (003 FR-015) is pending the
article text from the user."

## Project Scope and Compatibility

- **Affected interfaces**: Both. CLI `ge-tax-calc run` and bot `/calculate`
  show the same warning, offer the same years, and produce the same tax
  (constitution II). Currency conversion (`convert`, `/convert`) is not
  affected.
- **Affected boundaries**:
  - `internal/taxes/`: the renamed category and its generated name; tax on a
    total amount.
  - `internal/service/`: one-year check, tax on the total, the limit warning
    in the calculation result.
  - `cmd/ge-tax-calc/`: year options after the first income; warning in the
    output; category label.
  - `cmd/bot/`: the same for the bot flow and result message.
  - `internal/converter/`, `pkg/` packages: not changed.
- **Tax categories and rates**: Small Business (1%) gains the limit warning.
  Individual Entrepreneur (3%) is renamed to Small Business (over limit)
  (3%); its rate is unchanged. Employment (20%) is unchanged.
- **Currencies and rate dates**: Unchanged; each income still uses the NBG
  rate for its own date. All income dates of one calculation fall in one
  calendar year.
- **User-facing contract**: Category list and labels in both interfaces;
  year choices for the second and later incomes; a warning line in the
  calculation result; a new error for incomes from several years (service
  only, as interfaces prevent it). No new commands, flags, or environment
  variables.
- **Release impact**: Both binaries and the bot image; `TaxType` names are
  regenerated with `make codegen` (constitution VII).

## Clarifications

Decisions taken in `specs/003-tax-calculation/spec.md` (Session 2026-11-15)
and in this conversation:

- The limit only triggers a warning; the tax is still calculated at the
  category rate (003 FR-013).
- One calculation covers one calendar year, fixed by the first income (003
  FR-014).
- Tax is calculated once on the total converted income (003 FR-004).
- The Individual Entrepreneur (3%) category is kept and renamed to reflect
  the Small Business rate above the limit.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See a warning when the Small Business limit is exceeded (Priority: P1)

A taxpayer with Small Business status enters income since the start of the
year and new incomes. If the total for the year goes over the annual
turnover limit, the result says so, so the taxpayer knows that part of the
income is taxed differently and the status may be at risk.

**Why this priority**: Without the warning the tool silently reports 1% tax
for turnover the law taxes at a higher rate; this is the main reason the
year income is collected.

**Independent Test**: Calculate with Small Business, year income
`499000`, and one income of `2000 GEL` in either interface and check the
warning; repeat with a total below the limit and check that no warning is
shown.

**Acceptance Scenarios**:

1. **Given** category Small Business, year income 499 000 GEL, and an income
   of 2 000 GEL, **When** the tax is calculated, **Then** the result shows a
   warning that the year income of 501 000 GEL exceeds the limit of
   500 000 GEL by 1 000 GEL, and the tax is 20 GEL (1% of 2 000 GEL).
2. **Given** category Small Business and a year income of exactly
   500 000 GEL after the new incomes, **When** the tax is calculated,
   **Then** no warning is shown.
3. **Given** category Employment or Small Business (over limit) and a year
   income above 500 000 GEL, **When** the tax is calculated, **Then** no
   limit warning is shown.
4. **Given** the same input, **When** it is calculated in the CLI and in the
   bot, **Then** both show the same warning text and the same numbers.

---

### User Story 2 - Keep one calculation within one calendar year (Priority: P1)

A taxpayer adds several incomes. After the first income, only its year can
be chosen, so the year income and the limit refer to the same calendar
year.

**Why this priority**: Mixing years makes the year income, the limit
warning, and the declaration period meaningless.

**Independent Test**: In either interface add an income dated in 2025, then
start a second income and check that only 2025 is offered; call the
service with incomes from 2024 and 2025 and check the error.

**Acceptance Scenarios**:

1. **Given** a first income dated in 2025, **When** the user adds another
   income, **Then** the year choice offers only 2025 (preselected in the
   CLI, the only button in the bot).
2. **Given** the user answers "No" at the final review and the flow
   restarts income entry, **When** the first income is entered again,
   **Then** any year from 2016 to the current year can be chosen again.
3. **Given** a calculation request with incomes dated in 2024 and 2025,
   **When** the service calculates, **Then** it fails with an error that the
   incomes span more than one calendar year, and no result is returned.

---

### User Story 3 - Tax on the total converted income (Priority: P2)

A taxpayer with several incomes gets the tax computed once on their total,
as the declaration does, instead of a sum of separately rounded taxes.

**Why this priority**: The difference is a few tetri, but the result must
match the declaration.

**Independent Test**: Calculate three incomes of 0.50 GEL at 1% and check
that the tax is 0.02 GEL (rounding the total once), not 0.03 GEL (rounding
each income).

**Acceptance Scenarios**:

1. **Given** category Small Business and three incomes of 0.50 GEL, **When**
   the tax is calculated, **Then** the tax is 0.02 GEL.
2. **Given** category Employment, incomes converting to 1000 GEL and
   200 GEL, and year income 67.99 GEL, **When** the tax is calculated,
   **Then** the tax is 240 GEL and the year income is 1267.99 GEL (existing
   003 US1.3 still holds).

---

### User Story 4 - Choose the renamed over-limit category (Priority: P3)

A taxpayer whose turnover is already above the limit chooses the 3%
category under a name that says what it is.

**Why this priority**: The current name "Individual Entrepreneur" suggests
the general regime (20% on profit), which the tool does not calculate.

**Independent Test**: Open the category choice in both interfaces and check
the labels and rates.

**Acceptance Scenarios**:

1. **Given** the category choice, **When** it is shown, **Then** it lists
   Employment (20%), Small Business (1%), and Small Business (over limit)
   (3%) in this order, and no longer lists Individual Entrepreneur.
2. **Given** the category Small Business (over limit), **When** the tax is
   calculated, **Then** the result shows the rate as `Small Business (over
   limit) 3 %`.

---

### Edge Cases

- Year income of 0 with incomes above the limit: the warning is shown.
- Year income alone already above the limit before the new incomes: the
  warning is shown.
- The limit is compared with the reported year income (entered year income
  plus total converted income), strictly greater than the limit.
- Incomes in different currencies but the same year: allowed.
- A first income in the current year: later incomes are limited to the
  current year, with months and days up to today as before.
- Bot: a year button from an older keyboard (another year) is rejected as
  unexpected input (004 FR-015), because the year step now offers one year.
- `/cancel` or restarting the flow clears the fixed year.
- Rounding: converted amounts stay rounded to 2 places per income (002
  FR-003); only the tax is rounded once on their total.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: For the Small Business category, when the reported year income
  exceeds the Small Business annual turnover limit, the calculation result
  MUST include a warning that states the limit, the reported year income,
  and the amount above the limit, in GEL.
- **FR-002**: The warning MUST NOT change the tax: the tax is still
  calculated at the selected category's rate.
- **FR-003**: No limit warning MUST be shown for other categories or when the
  reported year income is equal to or below the limit.
- **FR-004**: The limit MUST be 500 000 GEL until confirmed or corrected
  against the Tax Code of Georgia (003 FR-015).
- **FR-005**: After the first income of a calculation, both interfaces MUST
  offer only the year of that income for later incomes; restarting income
  entry or the flow MUST clear this restriction.
- **FR-006**: The service MUST reject a calculation whose incomes are dated
  in more than one calendar year, with an error that says so.
- **FR-007**: The tax MUST be calculated once as the total converted income
  multiplied by the category rate and rounded to 2 decimal places.
- **FR-008**: The category with the 3% rate MUST be shown and accepted under
  the name "Small Business (over limit)"; the name "Individual
  Entrepreneur" MUST no longer be offered or accepted.
- **FR-009**: The CLI and the bot MUST show the same warning text, the same
  year choices, and the same tax for the same input.

### Key Entities

- **Limit warning**: Limit (GEL), reported year income (GEL), amount above
  the limit (GEL); present only when FR-001 applies.
- **Calculation year**: The calendar year of the first income; all incomes of
  the calculation share it.
- **Tax category**: Name and rate; "Small Business (over limit)" replaces
  "Individual Entrepreneur" with the same 3% rate.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For the inputs in US1.1–US1.3, both interfaces show the
  warning exactly when the reported year income is above 500 000 GEL, with
  the excess shown to the tetri.
- **SC-002**: For three incomes of 0.50 GEL at 1%, the tax is 0.02 GEL in
  both interfaces.
- **SC-003**: A calculation with incomes from two different years cannot be
  entered in either interface and is rejected by the service.
- **SC-004**: The category list in both interfaces contains no "Individual
  Entrepreneur" entry.
- **SC-005**: The CLI and the bot produce the same warning, tax, and totals
  for the same input.

## Assumptions

- The Small Business annual turnover limit is 500 000 GEL and the over-limit
  rate is 3%. Secondary sources agree on both
  (e.g., expathub.ge, Legalese Georgia, PwC Tax Summaries, checked on
  2026-11-15); the primary text (Tax Code of Georgia, articles on small
  business status) is pending from the user because matsne.gov.ge blocks
  automated access. Secondary sources disagree on how the 3% applies over
  the limit and when the status is lost, so the warning states only the
  facts above and does not compute the 3%.
- Whether the general Individual Entrepreneur regime should be supported is
  out of scope; it taxes profit, which the tool does not collect.
- No compatibility alias for the old category name is needed: categories are
  not stored, and both interfaces build their choices from the category
  list.
- The year income entered by the user refers to the calculation year.
