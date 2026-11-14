# Quickstart: Validate Year Income Rules

**Feature**: [spec.md](spec.md)

## Prerequisites

- Go 1.27.x; repository root as working directory.
- Network access to nbg.gov.ge for the manual CLI runs (automated tests do
  not need it).

## Automated checks

```sh
make codegen               # regenerates internal/taxes/taxtype_string.go
go vet ./...
go test -race ./internal/taxes/... ./internal/service/... ./cmd/...
make lint-pipeline
```

Expected: all pass. Covered cases: rename and parsing, tax on the total
(3 × 0.50 GEL at 1% = 0.02 GEL), mixed years rejected, warning above / at /
below the limit and only for Small Business, CLI and bot year options after
the first income. See [data-model.md](data-model.md) and
[contracts/calculation-output.md](contracts/calculation-output.md).

## Manual CLI scenarios (`go run ./cmd/ge-tax-calc run`)

| # | Input | Expected |
|---|-------|----------|
| 1 | Small Business, year income `499000`, one income 2000 GEL (any date this year) | Tax 20 GEL; warning "Year income 501000 GEL exceeds … by 1000 GEL" |
| 2 | Small Business, year income `498000`, one income 2000 GEL | Tax 20 GEL; no warning (500000 is not above the limit) |
| 3 | Employment, year income `600000`, one income 100 GEL | Tax 20 GEL; no warning |
| 4 | Add an income in 2025, answer "Yes" to add another | Year prompt lists only 2025 |
| 5 | Category list | Employment, Small Business, Small Business (over limit) |

## Manual bot scenarios (test bot, `/calculate`)

Repeat scenarios 1, 4, and 5; after scenario 4 press a year button from the
first income's year keyboard for another year: the bot replies
"⚠️ Unexpected input".
