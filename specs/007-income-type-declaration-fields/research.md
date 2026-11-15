# Research: Income Type and Declaration Field Numbers

**Feature**: [spec.md](spec.md) | **Date**: 2026-11-15

No open unknowns remained after `/speckit-clarify`; the decisions below fix
how the spec maps onto the existing code.

## R1. Where the income type lives

- **Decision**: A new `IncomeType` enum in `internal/taxes` (`income_type.go`)
  with a fixed descriptor table: name, declaration field number, and short
  description. `String()`, `Field()`, `Label()` ("Cash (18)"),
  `Description()`, `ParseIncomeType`, and `AllIncomeTypes` (in field order)
  read the table. No stringer.
- **Rationale**: The interfaces need the list of types for their choices, and
  constitution I allows them to use shared value types from `internal/taxes`
  but not a new internal package. The type carries more than a name (field
  number, description), so a descriptor table is simpler than stringer plus
  side maps, and it adds no generated file that `AGENTS.md` and constitution
  VII would have to list.
- **Alternatives considered**: a new `internal/declaration` package (cleaner
  name, but the interfaces would import a package the constitution does not
  list); stringer with `--linecomment` like `TaxType` (still needs maps for
  the field number and description, and a new generated file).

## R2. Which categories use the declaration

- **Decision**: `TaxType.IsSmallBusiness()` returns true for Small Business and
  Small Business (over limit). The service and both interfaces use it to
  decide on the income type step, the one-month rule, the year-income prompt,
  and the declaration block.
- **Rationale**: One rule in the domain package; the interfaces already parse
  the tax type name with `taxes.ParseTaxType`.

## R3. Income type in the request

- **Decision**: `service.Income` gains `Type string`, the income type name
  ("Cash", "POS terminal", "Non-cash", "Other", "Unpaid supply"). The service
  parses it case-insensitively with `taxes.ParseIncomeType`. Bot callback
  values use the name; the longest, `cb:calc_inc_type:Unpaid supply`, is 30
  bytes, within Telegram's 64-byte limit.
- **Rationale**: Matches how `TaxType` and `Currency` already travel as
  strings in requests and are parsed by the service.

## R4. Validation order

- **Decision**: `Calculate` validates in this order, all before any
  conversion: tax type, year income, one calendar year (006), and for the
  Small Business categories one calendar month and a valid type on every
  income. For Employment the type is not parsed, so a stray value is ignored.
- **Rationale**: An invalid request makes no NBG calls (as 006 R4); the month
  check needs a valid single year to report clearly.
- **Errors**: `ErrIncomesSpanMonths` ("incomes span more than one calendar
  month") in `internal/service`; `taxes.ErrInvalidIncomeType` ("invalid income
  type") wrapped as `income <n>: "<raw>": invalid income type`.

## R5. One-month rule in the interfaces

- **Decision**: As with the year in 006, the month is derived from
  `incomes[0]`: once an income is captured in a Small Business calculation,
  the CLI month prompt and the bot month keyboard offer only that month
  (preselected in the CLI). Clearing the list (review "No", `/cancel`, a new
  `/calculate`) clears the restriction. The CLI income model now receives the
  tax type (`newIncomeModel(taxType)`), because tax details and incomes are
  collected by two separate Bubble Tea programs.
- **Rationale**: One source of truth; old month buttons in the bot fail
  `callbackValue` because the step offers only one month (004 FR-015).
- **Alternatives considered**: skipping the year and month steps after the
  first income (fewer inputs, but the user would not see the date being
  used, and 006 kept the year step for the same reason).

## R6. Declaration block from the service

- **Decision**: `service.Calculate` returns `Declaration *Declaration` in
  `CalculateResponse` (nil for Employment). It holds the eight fields in form
  order, each with number, label, numeric value, and the value formatted with
  two decimals. `DeclarationField.String()` gives the line
  `(<n>) <label>: <value>` and `Declaration.String()` the header plus lines,
  so the CLI prints the service text and the bot renders the same lines.
- **Rationale**: Constitution I and II, and the pattern of `LimitWarning`
  (006 R2): one text source guarantees FR-009; the bot needs the value
  separately to mark it monospace (R9).
- **Values**: (15) reported year income, (16) category rate × 100, (17)
  total converted income, (18)–(22) per-type sums of the converted amounts
  with `moneyutils.Add`. Converted amounts are already rounded to 2 places by
  the converter, so (17) equals the sum of (18)–(22) exactly (FR-005).

## R7. Two-decimal formatting

- **Decision**: Add `moneyutils.ToStringFixed(v float64, places int32) string`
  (decimal `StringFixed`), used for declaration values only. Other output
  keeps `models.Money` formatting (`35942 GEL`).
- **Rationale**: Constitution III requires money formatting through
  `pkg/moneyutils`; the form shows `4150.00`, and FR-006 requires exactly two
  decimals. Changing `models.Money` would change every existing line
  (SC-005).

## R8. Per-income type in the result

- **Decision**: `CalculateResponse.Incomes` becomes `[]IncomeResponse`, where
  `IncomeResponse` embeds `ConvertResponse` and adds `Type taxes.IncomeType`
  (zero value for Employment). Its `String()` adds a `Type: Cash (18)` line
  when the type is set.
- **Rationale**: Embedding keeps `inc.Amount`, `inc.Converted`, `inc.Rate`
  working in both interfaces; `ConvertResponse` stays unchanged for
  `Convert`.

## R9. Monospace values in the bot

- **Decision**: Send the result message as plain text with `Entities`: one
  `code` entity (`telego.EntityTypeCode`) per declaration value, with offset
  and length in UTF-16 code units, computed by a small helper in `cmd/bot`
  (`unicode/utf16`). `formatCalcResult` returns the text and the entities.
- **Rationale**: Entities leave the rest of the text untouched, so the emoji,
  the warning, and any error text cannot be lost or parsed as markup (spec
  edge case). Telegram copies a `code` span with a tap (SC-007).
- **Alternatives considered**: `ParseMode: HTML` with `<code>` (every dynamic
  part must be escaped, and a missed character breaks the message);
  `MarkdownV2` (more characters to escape); `telegoutil` entity helpers (not
  vendored today; the helper needed is a few lines).

## R10. Limit warning recalculation line

- **Decision**: `LimitWarning.String()` gets a second line:
  `For this month's declaration, recalculate with Small Business (over limit): field (16) is 3.00.`
  The category name and the rate come from the rate table and
  `moneyutils.ToStringFixed`.
- **Rationale**: FR-012; the warning exists only for Small Business (1%), so
  the line is always relevant when the warning is shown; the text has one
  source for both interfaces.

## R11. Year-income prompt for Small Business

- **Decision**: After a Small Business category is chosen, the CLI input
  title is `Field (15) of your previous month's declaration: income since the start of the year (GEL)`
  with the note `Enter 0 for your first declaration of the year.`; the bot
  text says the same. Employment keeps the current text in both interfaces.
- **Rationale**: FR-013. "First declaration of the year" covers January and a
  status obtained mid-year.

## R12. Layout and defaults

- **Decision**: The declaration block follows the tax line, separated by an
  empty line, and the limit warning stays the last line (006 contract). In
  the CLI the type of the previous income is preselected; for the first
  income the cursor starts on the first option (Cash). In the bot the type
  keyboard has one button per row, and the message lists each type with its
  description.
- **Rationale**: The warning's new line refers to field (16), so it reads
  after the block; one column fits the longest label; the bot cannot show
  option descriptions on buttons, so the message carries them (parity with
  the CLI descriptions).
- **Field labels** (deferred by `/speckit-clarify`): (15) `Income since the
  start of the year`, (16) `Tax rate, %`, (17) `Taxable income for the
  month`, (18)–(22) the income type names.
