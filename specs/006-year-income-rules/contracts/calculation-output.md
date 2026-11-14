# Contract: Calculation Output (CLI `run`, bot `/calculate`)

**Feature**: [../spec.md](../spec.md)

## Category labels

| Interface | Labels (order) |
|-----------|----------------|
| CLI select | `Employment` (20 %), `Small Business` (1 %), `Small Business (over limit)` (3 %) |
| Bot buttons | `Employment (20%)`, `Small Business (1%)`, `Small Business (over limit) (3%)`; callback value is the name, e.g. `cb:calc_tax:Small Business (over limit)` |

Result line for the rate (both): `Tax Rate: Small Business (over limit) 3 %`.

## Limit warning line

Printed after the tax line, identical text in both interfaces:

```text
⚠️ Year income <YearIncome> GEL exceeds the Small Business limit of <Limit> GEL by <Excess> GEL. Under Article 90(2) of the Tax Code, income from the start of the month the limit was exceeded until the end of the year is taxed at <over-limit rate>: calculate it with Small Business (over limit). The tax above is calculated at <Small Business rate>.
```

The rates and the category name come from the rate table (`3%`, `1%` today),
so the text follows any rate change.

Numbers use the existing `models.Money` format (shortest decimal, no
thousands separator). Example for US1.1:

```text
⚠️ Year income 501000 GEL exceeds the Small Business limit of 500000 GEL by 1000 GEL. Under Article 90(2) of the Tax Code, income from the start of the month the limit was exceeded until the end of the year is taxed at 3%: calculate it with Small Business (over limit). The tax above is calculated at 1%.
```

- CLI: last line of `CalculateResponse.String()` output.
- Bot: last line of the "🧾 Tax Calculation Result" message, after "Taxes to
  Pay".

## Year choice after the first income

| Interface | Behavior |
|-----------|----------|
| CLI | "Select year of income" lists only the first income's year, preselected |
| Bot | "📅 Select the year of income:" shows one button with that year; any other year button is rejected as unexpected input |

## Errors

| Condition | Service error text (wrapped by `failed to …` as today) |
|-----------|--------------------------------------------------------|
| Incomes from several years | `incomes span more than one calendar year` |
| Old category name "Individual Entrepreneur" | `Individual Entrepreneur: invalid tax type` (via `failed to parse tax type`) |
