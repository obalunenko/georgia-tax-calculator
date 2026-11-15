# Data Model: Income Type and Declaration Field Numbers

**Feature**: [spec.md](spec.md) | **Date**: 2026-11-15

## IncomeType (new, `internal/taxes`)

| Value | Constant | Name (`String()`) | Field | Label (`Label()`) | Description |
|-------|----------|-------------------|-------|-------------------|-------------|
| 0 | `incomeTypeUnknown` | — | — | — | not a valid type |
| 1 | `IncomeTypeCash` | Cash | 18 | Cash (18) | cash payments (cash register turnover) |
| 2 | `IncomeTypePOSTerminal` | POS terminal | 19 | POS terminal (19) | card payments through a POS terminal |
| 3 | `IncomeTypeNonCash` | Non-cash | 20 | Non-cash (20) | bank transfers and other non-cash payments (not POS) |
| 4 | `IncomeTypeOther` | Other | 21 | Other (21) | services not taxed at source, barter, and other income |
| 5 | `IncomeTypeUnpaidSupply` | Unpaid supply | 22 | Unpaid supply (22) | goods or services supplied, not paid for this month |

- `AllIncomeTypes()` returns the five valid types in field order.
- `ParseIncomeType(raw)` accepts the names case-insensitively, trimming
  spaces; anything else returns `ErrInvalidIncomeType`
  (`"<raw>": invalid income type`).
- `Valid()` is true for values 1–5.
- `IncomeTypeLabel(name)` returns the label for a type name, or `""`; both
  interfaces use it for income lines.
- For invalid values (0, ≥ 6): `String()` returns `IncomeType(<n>)`, `Field()`
  returns 0, `Label()` and `Description()` return `""`.

## TaxType (changed, `internal/taxes`)

Adds `IsSmallBusiness() bool`: true for `TaxTypeSmallBusiness` and
`TaxTypeSmallBusinessOverLimit`. `TaxRate` gains `Percent() float64` (rate ×
100), used by every place that shows a rate. No enum, rate, or generated-code
change.

## Income (changed, `internal/service` request)

| Field | Type | Change |
|-------|------|--------|
| `DateRequest` | year, month, day strings | unchanged |
| `Currency` | string | unchanged |
| `Amount` | string | unchanged |
| `Type` | string | new: income type name; required for the Small Business categories, ignored for Employment |

## IncomeResponse (new, `internal/service`)

| Field | Type | Meaning |
|-------|------|---------|
| embedded `ConvertResponse` | date, amount, converted, rate | as today |
| `Type` | `taxes.IncomeType` | parsed type; zero value for Employment |

`String()`: the `ConvertResponse` lines plus `Type: <Label>` when `Type` is
valid.

## DeclarationField (new, `internal/service`)

| Field | Type | Meaning |
|-------|------|---------|
| `Number` | int | declaration field number, 15–22 |
| `Label` | string | short English label ([contract](contracts/calculation-output.md)) |
| `Value` | float64 | GEL amount, or the rate in percent for field 16 |

Methods: `Formatted()` returns `Value` with exactly two decimals
(`moneyutils.ToStringFixed`), `Prefix()` returns `(<Number>) <Label>: `, and
`String()` returns `Prefix() + Formatted()`. The value is not stored twice, so
the text cannot drift from it (code review, 2026-11-15).

## Declaration (new, `internal/service`)

| Field | Type | Meaning |
|-------|------|---------|
| `Fields` | `[]DeclarationField` | fields 15–22, in this order, all present |

`String()` returns the header line followed by one line per field.

| Field | Value |
|-------|-------|
| 15 | reported year income (entered year income + total converted income) |
| 16 | category rate × 100 (1.00 or 3.00) |
| 17 | total converted income |
| 18–22 | sum of `Converted.Amount` of the incomes whose type has that field; 0 when none |

Invariant: field 17 = sum of fields 18–22 = `TotalIncomeConverted`; field
15 = `YearIncome` (FR-005).

## CalculateResponse (changed, `internal/service`)

| Field | Change |
|-------|--------|
| `Incomes` | `[]ConvertResponse` → `[]IncomeResponse` |
| `Declaration` | new `*Declaration`; nil for Employment |
| `LimitWarning` | unchanged type; `String()` gains the recalculation line |
| others | unchanged |

`String()` order: tax rate, year income, incomes, total, taxes, empty line
and the declaration block (when present), then the warning (when present).

## LimitWarning (changed text, `internal/service`)

Fields unchanged. `String()` appends a second line, see
[contract](contracts/calculation-output.md#limit-warning).

## Calculation month (derived, interfaces)

- CLI `incomeModel` and bot `session`: the month of `incomes[0]` once an
  income is captured and the category is a Small Business category;
  otherwise none.
- Not stored separately: clearing the income list clears it (as the
  calculation year in 006).

## Interface state (changed)

| Interface | Change |
|-----------|--------|
| CLI `incomeModel` | knows the tax type (`newIncomeModel(taxType)`); new step `incomeStepType` after `incomeStepCurrency` for the Small Business categories |
| Bot `session` | new step `calcStepIncomeType` after `calcStepCurrency`, step ID `calc_inc_type`, used only for the Small Business categories |

## Validation rules

| Rule | Where | Error |
|------|-------|-------|
| All incomes in one calendar year | `service.Calculate`, before conversion | `ErrIncomesSpanYears` (006, unchanged) |
| Small Business: all incomes in one calendar month | `service.Calculate`, before conversion | `ErrIncomesSpanMonths`: `<month> and <month>: incomes span more than one calendar month` |
| Small Business: every income has a valid type | `service.Calculate`, before conversion | `income <n>: "<raw>": invalid income type` (`taxes.ErrInvalidIncomeType`) |
| Year income ≥ 0, amounts > 0 | unchanged | `ErrInvalidAmount` |
| Year income and amounts within float64 (e.g. `1e400` is rejected) | `parseMoney`, used by the CLI, the bot, and the service | `ErrInvalidAmount`: `"<raw>" is too large` |

The reported year income is rounded to 2 decimals before use, so the Year
Income line, the limit check, and field 15 agree (FR-005).
