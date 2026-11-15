# Contract: Calculation Input and Output (CLI `run`, bot `/calculate`)

**Feature**: [../spec.md](../spec.md)

Everything below applies to the Small Business categories (Small Business,
Small Business (over limit)) only. Employment input and output stay exactly
as today (FR-002, SC-005).

## Year-income prompt (FR-013)

| Interface | Employment (unchanged) | Small Business categories |
|-----------|------------------------|---------------------------|
| CLI input title | `Income from the beginning of a calendar year (GEL)` | `Field (15) of your previous month's declaration: income since the start of the year (GEL)` |
| CLI note | none | `Enter 0 for your first declaration of the year.` |
| Bot message after the tax type | `✅ Tax type: <name>` / `💰 Enter your income from the beginning of the calendar year in GEL:` / `(e.g. 0.00 if this is your first income)` | `✅ Tax type: <name>` / `💰 Enter field (15) of your previous month's declaration: income since the start of the year in GEL:` / `(0.00 for your first declaration of the year)` |

Input rules are unchanged (`≥ 0`, dot as decimal separator).

## Month choice after the first income (FR-010)

| Interface | Behavior |
|-----------|----------|
| CLI | "Select month of income" lists only the first income's month, preselected (the year prompt already lists only its year, 006) |
| Bot | "📅 Select the month of income:" shows one button with that month; any other month button is rejected with "⚠️ Unexpected input …" |

Days follow the existing rules for that month (up to today in the current
month).

## Income type choice (FR-001)

Shown right after the currency of each income.

Options, in this order (label, then description):

| Label | Value sent | Description |
|-------|------------|-------------|
| `Cash (18)` | `Cash` | cash payments (cash register turnover) |
| `POS terminal (19)` | `POS terminal` | card payments through a POS terminal |
| `Non-cash (20)` | `Non-cash` | bank transfers and other non-cash payments (not POS) |
| `Other (21)` | `Other` | services not taxed at source, barter, and other income |
| `Unpaid supply (22)` | `Unpaid supply` | goods or services supplied, not paid for this month |

| Interface | Presentation |
|-----------|--------------|
| CLI | Select prompt `Select income type (declaration field)`; each option shows `<Label> — <Description>`; the previous income's type is preselected, the first income starts on `Cash (18)` |
| Bot | Message `✅ Currency: <currency>` / empty line / `🧾 Select the income type (declaration field):` / one line `• <Label> — <Description>` per type; keyboard of five buttons, one per row, text `<Label>`, callback `cb:calc_inc_type:<Value>` |

After the choice the flow continues as today with "Add another income
entry?"; the bot confirmation starts with `✅ Income type: <Label>`.

Typing text at this step in the bot replies "⚠️ Unexpected input …", as at
other button steps. The bot accepts buttons only from the message with the
current keyboard, so a type button of an earlier income's message is
rejected the same way (spec US2.4).

## Income lists and review (FR-007)

| Place | Line format (Small Business categories) |
|-------|------------------------------------------|
| CLI "Captured incomes" | `  1) 2026-October-05 — 3950 GEL — Cash (18)` |
| CLI review details | extra line `   Type: Cash (18)` after `   Amount: …` |
| Bot "Captured incomes" and review | `  1) 2026-October-05 — 3950 GEL — Cash (18)` |

## Result (FR-004, FR-006, FR-007)

### CLI (`CalculateResponse.String()`)

Example for spec US1.1 (year income 31792, 3950 GEL Cash and 200 GEL
Non-cash on 2026-10-05 and 2026-10-20):

```text
Tax Rate: Small Business 1 %
Year Income: 35942 GEL
Incomes:
	- 1:
		Date: 2026-10-05
		Amount: 3950 GEL
		Converted: 3950 GEL
		Rate: 1
		Type: Cash (18)
	- 2:
		Date: 2026-10-20
		Amount: 200 GEL
		Converted: 200 GEL
		Rate: 1
		Type: Non-cash (20)
Total Income Converted: 4150 GEL
Taxes: 41.5 GEL

Declaration fields (Small Business monthly declaration):
(15) Income since the start of the year: 35942.00
(16) Tax rate, %: 1.00
(17) Taxable income for the month: 4150.00
(18) Cash: 3950.00
(19) POS terminal: 0.00
(20) Non-cash: 200.00
(21) Other: 0.00
(22) Unpaid supply: 0.00
```

### Bot (`formatCalcResult`)

```text
🧾 Tax Calculation Result

Tax Rate: Small Business 1 %
Year Income: 35942 GEL

Incomes converted to GEL:
  1) 3950 GEL → 3950 GEL (rate: 1) — Cash (18)
  2) 200 GEL → 200 GEL (rate: 1) — Non-cash (20)

Total Income (GEL): 4150 GEL
Taxes to Pay: 41.5 GEL

Declaration fields (Small Business monthly declaration):
(15) Income since the start of the year: 35942.00
…
(22) Unpaid supply: 0.00
```

The block lines are the service's `Declaration.String()` lines, so both
interfaces print the same labels and values (FR-009).

### Bot formatting (FR-014)

- The message is sent without a parse mode. Each value of the block (e.g.
  `35942.00`, `1.00`) gets one `code` entity, offset and length in UTF-16
  code units; nothing else is formatted.
- Eight entities when the block is present, none for Employment.

### Field labels

| Field | Label | Value |
|-------|-------|-------|
| 15 | `Income since the start of the year` | GEL, 2 decimals |
| 16 | `Tax rate, %` | percent, 2 decimals |
| 17 | `Taxable income for the month` | GEL, 2 decimals |
| 18 | `Cash` | GEL, 2 decimals |
| 19 | `POS terminal` | GEL, 2 decimals |
| 20 | `Non-cash` | GEL, 2 decimals |
| 21 | `Other` | GEL, 2 decimals |
| 22 | `Unpaid supply` | GEL, 2 decimals |

Values have no thousands separator, matching the form input (`35942.00`).

## Limit warning

The 006 warning line is unchanged and gets a second line (FR-012); both are
the last lines of the result, after the declaration block:

```text
⚠️ Year income 501000 GEL exceeds the Small Business limit of 500000 GEL by 1000 GEL. Under Article 90(2) of the Tax Code, income from the start of the month the limit was exceeded until the end of the year is taxed at 3%: calculate it with Small Business (over limit). The tax above is calculated at 1%.
For this month's declaration, recalculate with Small Business (over limit): field (16) is 3.00.
```

The category name and `3.00` come from the rate table.

## Errors

Service errors, wrapped by `failed to check incomes: …` as the 006 year
check; both interfaces prevent them, so they surface only through the
service API.

| Condition | Error text |
|-----------|------------|
| Small Business incomes in two months of one year | `September and October: incomes span more than one calendar month` |
| Small Business income without a type | `income 2: "": invalid income type` |
| Small Business income with an unknown type | `income 2: "Card": invalid income type` |
