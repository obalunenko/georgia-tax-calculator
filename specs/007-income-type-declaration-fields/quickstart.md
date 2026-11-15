# Quickstart: Validate Income Type and Declaration Fields

**Feature**: [spec.md](spec.md)

## Prerequisites

- Go 1.27.x; repository root as working directory.
- Network access to nbg.gov.ge for the manual runs with foreign currencies
  (automated tests do not need it; GEL incomes need no rate).
- For the bot: a test bot token in `TELEGRAM_BOT_TOKEN`.

## Automated checks

```sh
go vet ./...
go test -race ./pkg/moneyutils/... ./internal/taxes/... ./internal/service/... ./cmd/...
make lint-pipeline
```

Expected: all pass. Covered cases:

- `internal/taxes`: income type names, fields, order, parsing;
  `IsSmallBusiness`.
- `pkg/moneyutils`: `ToStringFixed` (`4150` → `4150.00`, `0.005` rounding,
  `0` → `0.00`).
- `internal/service`: declaration values for spec US1.1–US1.3 (with
  `mockConverter`, 1:1 conversion); (17) = sum of (18)–(22); no declaration
  for Employment and types ignored there; missing and unknown types rejected;
  two months rejected for Small Business, allowed for Employment; all before
  conversion; warning recalculation line.
- `cmd/ge-tax-calc`: type step only for Small Business, previous type
  preselected, month limited after the first income, year-income prompt per
  category, type in lists and review.
- `cmd/bot`: type keyboard and callback, old month and type buttons rejected,
  year-income text per category, eight `code` entities on the result with
  correct UTF-16 offsets, none for Employment.

See [data-model.md](data-model.md) and
[contracts/calculation-output.md](contracts/calculation-output.md).

## Manual CLI scenarios (`go run ./cmd/ge-tax-calc run`)

| # | Input | Expected |
|---|-------|----------|
| 1 | Small Business; year income `31792`; 3950 GEL Cash and 200 GEL Non-cash, both in one month | Block: (15) 35942.00, (16) 1.00, (17) 4150.00, (18) 3950.00, (19) 0.00, (20) 200.00, (21) 0.00, (22) 0.00; tax 41.5 GEL (spec SC-001) |
| 2 | Small Business (over limit); one income | (16) 3.00 |
| 3 | Small Business; first income in October, then "Yes" to add another | Year prompt lists only that year, month prompt only October; type prompt preselects the first income's type |
| 4 | Small Business; year income `499000`; 2000 GEL Non-cash | (16) 1.00, (20) 2000.00; tax 20 GEL; warning with the line "For this month's declaration, recalculate with Small Business (over limit): field (16) is 3.00." |
| 5 | Employment; any incomes | Old year-income title, no type step, no month limit, no block |
| 6 | Small Business; tax type chosen | Year-income title asks for field (15) of the previous month's declaration |

## Manual bot scenarios (test bot, `/calculate`)

1. Repeat scenario 1: tap each value in the block; Telegram copies exactly
   `35942.00`, `1.00`, … (spec SC-007).
2. Repeat scenario 3; after the second income starts, press an older month
   button from the first income's keyboard: "⚠️ Unexpected input".
3. At the type step, type text instead of pressing a button: "⚠️ Unexpected
   input"; `/cancel`, then `/calculate` again: all months are offered.
4. Repeat scenario 4 and check that the emoji and the warning text are shown
   as plain text.
5. Repeat scenario 5: the result is the same as before the feature.
