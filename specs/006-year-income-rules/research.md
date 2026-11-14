# Research: Year Income Rules for Tax Calculation

**Feature**: [spec.md](spec.md) | **Date**: 2026-11-15

## R1. Small Business limit and over-limit rate

- **Decision**: Use 500 000 GEL as the annual turnover limit and keep 3% as
  the rate of the renamed category; the warning does not compute the 3%.
- **Rationale**: Secondary sources agree on 1% up to 500 000 GEL per
  calendar year and 3% above it: [expathub.ge](https://expathub.ge/tax-freelancers-individuals-small-businesses-georgia/),
  [Legalese Georgia](https://legalese.ge/blog/small-business-status-georgia-1-percent/),
  [Georgiafy](https://georgiafy.ge/small-business-micro-business-tax-georgia/),
  [PwC Tax Summaries](https://taxsummaries.pwc.com/georgia/individual/taxes-on-personal-income).
  They disagree on how the 3% applies (excess only, from which month) and on
  when the status is lost (two consecutive years; one source mentions a
  1 000 000 GEL ceiling), so the tool states only the facts it can compute.
- **Primary source**: Tax Code of Georgia on matsne.gov.ge (document
  1043717, consolidated version with amendment no. 248 of 28.10.2026,
  published 30.10.2026). Automated access returned "Access Denied"; the user
  provided the Georgian PDF on 2026-11-15. Confirmed: 1% (Article 90(1));
  3% when the total income from economic activity exceeds 500 000 GEL,
  from the start of the month in which the limit is exceeded until the end
  of the calendar year (Article 90(2)); the limit is 700 000 GEL for wine
  tourism and agrotourism (Articles 88(3), 90(2)); the status is cancelled
  from the next year if the limit is exceeded in each of 2 calendar years
  (Article 89(2)(a), 89(3)); Employment 20% (Article 81(1)). The secondary
  sources' 1 000 000 GEL ceiling is not in the text.
- **Alternatives considered**: computing 3% on the excess (rejected by the
  003 clarification "warning only"); a configurable limit (no need: the
  value changes with the law, not per user).

## R2. Where the warning is produced

- **Decision**: `service.Calculate` returns an optional `LimitWarning` in
  `CalculateResponse`, with a `String()` used by both interfaces.
- **Rationale**: Constitution I and II: the rule is domain logic; one text
  source guarantees identical output (spec FR-009).
- **Alternatives considered**: computing the warning in each interface
  (duplicated rule, risk of drift).

## R3. Comparison with the limit

- **Decision**: Warn when the reported year income (entered year income
  plus total converted income) is strictly greater than the limit; the
  excess is `reported − limit`, computed with `moneyutils.Add(yi, -limit)`.
- **Rationale**: "Exceeds" means above the limit; 500 000.00 is still within
  it (spec US1.2).

## R4. One calendar year

- **Decision**: The service checks the parsed year of every income before
  converting and returns `ErrIncomesSpanYears` if they differ. The CLI and
  the bot offer only the first income's year for later incomes; restarting
  income entry clears it, because the restriction is derived from the list
  of captured incomes, not stored separately.
- **Rationale**: No NBG calls for an invalid request; deriving the year from
  the income list keeps one source of truth and makes "No" at the review
  and `/cancel` reset it automatically.
- **Alternatives considered**: skipping the year step after the first
  income (fewer inputs, but the user would not see which year is used);
  storing a separate "calculation year" field (can drift from the list).

## R5. Tax on the total

- **Decision**: Sum the rounded converted incomes with `moneyutils.Add`,
  then call `taxes.Calc` once with the total; `taxes.Calc` already rounds to
  2 places.
- **Rationale**: Matches 003 FR-004/FR-005; reuses the existing rounding.
  Example: three incomes of 0.50 GEL at 1% give 0.02 GEL (per-income
  rounding gave 0.03, since `decimal.Round` rounds 0.005 up).

## R6. Renaming the 3% category

- **Decision**: Rename the constant to `TaxTypeSmallBusinessOverLimit` with
  the line comment `Small Business (over limit)`, keep its enum value (1),
  and regenerate `String()` with `make codegen`. No alias for the old name.
- **Rationale**: Keeping the value keeps the display order (sorted by value,
  descending: Employment, Small Business, Small Business (over limit)) and
  the rate table; categories are not persisted, so no compatibility alias is
  needed. Bot callback values use the name; the longest callback
  (`cb:calc_tax:Small Business (over limit)`) is 39 bytes, within
  Telegram's 64-byte limit.
- **Alternatives considered**: adding a new value and deprecating the old
  one (no stored data to migrate).
