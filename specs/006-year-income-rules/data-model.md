# Data Model: Year Income Rules for Tax Calculation

**Feature**: [spec.md](spec.md) | **Date**: 2026-11-15

## TaxType (changed)

| Value | Constant | Display name | Rate |
|-------|----------|--------------|------|
| 1 | `TaxTypeSmallBusinessOverLimit` (was `TaxTypeIndividualEntrepreneur`) | Small Business (over limit) | 3% |
| 2 | `TaxTypeSmallBusiness` | Small Business | 1% |
| 3 | `TaxTypeEmployment` | Employment | 20% |

- Parsing (`ParseTaxType`) accepts display names case-insensitively;
  "Individual Entrepreneur" is no longer accepted (`ErrInvalidTaxType`).
- New constant: `SmallBusinessTurnoverLimit` = 500 000 (GEL per calendar
  year); see [research.md](research.md) R1.

## LimitWarning (new, `internal/service`)

| Field | Type | Meaning |
|-------|------|---------|
| `Limit` | `models.Money` (GEL) | Small Business annual turnover limit |
| `YearIncome` | `models.Money` (GEL) | Reported year income (entered + converted) |
| `Excess` | `models.Money` (GEL) | `YearIncome − Limit`, > 0 |

- Present only when the category is Small Business and `YearIncome > Limit`.
- `String()` returns the text in
  [contracts/calculation-output.md](contracts/calculation-output.md).

## CalculateResponse (changed)

Adds `LimitWarning *LimitWarning` (nil when no warning). Other fields are
unchanged; `Tax` is now `round(TotalIncomeConverted × rate, 2)`.

`String()` appends the warning line after `Taxes:` when present.

## Calculation year (derived, interfaces)

- CLI `incomeModel` and bot `session`: the year of `incomes[0]` once at least
  one income is captured; otherwise none.
- Not stored separately: clearing the income list (review "No", `/cancel`,
  new `/calculate`) clears it.

## Validation rules

| Rule | Where | Error |
|------|-------|-------|
| All incomes have the same parsed year | `service.Calculate`, before conversion | `ErrIncomesSpanYears`: "incomes span more than one calendar year" |
| Year income ≥ 0, amounts > 0 | unchanged (002 FR-014, 003 FR-016) | `ErrInvalidAmount` |
