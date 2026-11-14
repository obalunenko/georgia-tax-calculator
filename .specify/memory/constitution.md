<!--
Sync Impact Report
- Version change: 1.0.2 → 1.0.3 (PATCH: factual correction after spec gap
  fixes)
- Modified: VI. Test Behavior Changes (`pkg/nbggovge/mock` now has its own
  tests; `cmd/bot` is partly tested)
- Templates: plan-template.md, tasks-template.md, spec-template.md — no
  change needed; AGENTS.md — no change needed
- Follow-up: none
- Previous: 1.0.1 → 1.0.2 corrected VII (currency codes are regenerated
  manually; `make codegen` runs only stringer)
-->

# Georgia Tax Calculator Constitution

## Project Identity

Georgia Tax Calculator helps prepare Georgian income tax declarations. It takes
income amounts in the received currency, converts them to GEL at the official
National Bank of Georgia (NBG) rate for the date of income, and calculates tax
for the selected tax category. The project is one Go module with two user
interfaces over a shared service layer: the interactive `ge-tax-calc` CLI and a
Telegram bot. Shell scripts, Docker Compose, and goreleaser support development
and release, not the runtime.

## Core Principles

### I. Keep the Service Layer the Single Domain Entry Point

- `cmd/ge-tax-calc/` (urfave/cli v3, Bubble Tea v2) and `cmd/bot/` (telego)
  are composition roots and user interfaces. They call `internal/service` for
  calculation and conversion and may use shared value types from
  `internal/taxes`, `pkg/nbggovge/currencies`, `pkg/moneyutils`, and
  `pkg/dateutils`.
- Interfaces MUST NOT import `internal/converter` or `pkg/nbggovge` directly;
  new domain behavior lives in `internal/service` or the packages below it.
- `internal/service` composes `internal/converter` (currency conversion) and
  `internal/taxes` (tax categories and rates); `internal/models` holds shared
  domain types.

Evidence: imports of `cmd/ge-tax-calc` and `cmd/bot`; `service.New()` in
`internal/service/service.go`.

### II. Keep CLI and Telegram Bot Capabilities in Parity

- A user-facing capability (tax calculation, currency conversion, supported tax
  categories, currencies, and input rules) MUST be available in both interfaces
  unless the feature spec explicitly scopes it to one and states why.
- Changes to user input flows MUST describe both the CLI menu/commands and the
  bot flow (`/start`, `/calculate`, `/convert`, `/cancel`, `/help`).

Evidence: README states that `ge-bot` provides the same functionality as the
CLI; both interfaces call the same `service.Service`.

### III. Compute Money Through Official Rates and `moneyutils`

- Income is converted to GEL using the NBG official rate for the income date.
- Money arithmetic, rounding, parsing, and formatting MUST go through
  `pkg/moneyutils`, which performs decimal (`shopspring/decimal`) arithmetic
  behind its `float64` API. Do not add ad-hoc floating-point money math.
- Tax categories are the `taxes.TaxType` enum with rates in
  `internal/taxes`. Adding or changing a category or rate MUST update the enum,
  the rate table, the generated `String()` code, and tests together.

Evidence: `internal/taxes/calculator.go`, `pkg/moneyutils/moneyutils.go`,
`internal/converter/converter.go`.

### IV. Access NBG Only Through the Client Interface

- Currency rates come from `pkg/nbggovge.Client`. Production code uses the
  cached client (`nbggovge.NewCached`); cache behavior is documented in
  `pkg/nbggovge/README_CACHE.md`.
- Tests MUST NOT call the real NBG API by default. Use `pkg/nbggovge/mock`,
  an injected `HTTPClient`, or JSON fixtures under `testdata/`.
- Tests that call the real API MUST be gated by `INTEGRATION_TESTS=true` and run
  through `make test-integration`.

Evidence: `pkg/nbggovge/client.go`, `pkg/nbggovge/mock/`,
`pkg/nbggovge/client_test.go`, `internal/converter/testdata/`.

### V. Keep `pkg/` Independent of `internal/`

- Packages under `pkg/` are reusable libraries and MUST NOT import `internal/`
  packages of this module.
- `pkg/nbggovge/internal` and `pkg/nbggovge/option` are implementation details
  of the NBG client.

Evidence: no `pkg/` package imports `internal/` today.

### VI. Test Behavior Changes

- Behavior changes in `internal/` and `pkg/` MUST include or update tests.
  Tests for `cmd/bot` (partly covered: callback and amount validation,
  per-user serialization; Telegram I/O is untested), `pkg/moneyutils`, and
  `internal/spinner` (both untested) are encouraged but not required. The
  `pkg/nbggovge` support packages `currencies`, `internal`, and `option` have
  no tests of their own and are exercised through `pkg/nbggovge` and
  `internal/converter` tests; `pkg/nbggovge/mock` has its own tests;
  `pkg/nbggovge/currencies/generator` is a development tool.
- Tests live next to the code in `*_test.go` files, use
  `github.com/stretchr/testify` (`assert`/`require`), and prefer table-driven
  cases with `t.Run`. Fixtures go in the package's `testdata/` directory.
- Tests MUST pass with the race detector (`go test -race ./...`).
- No numeric coverage threshold or test-first process is required.

Evidence: existing `*_test.go` files; `scripts/tests/run.sh`.

### VII. Do Not Hand-Edit Generated or Vendored Code

- `internal/taxes/taxtype_string.go` is generated by stringer; regenerate it
  with `make codegen` after changing `TaxType`.
- `pkg/nbggovge/currencies/codes.gen.go` is generated by the currencies
  generator, which `make codegen` does not run. Regenerate it manually from
  the repository root with
  `go run ./pkg/nbggovge/currencies/generator -client -output pkg/nbggovge/currencies/codes.gen.go`
  (live NBG API) or with `-input <rates.json>` instead of `-client`.
- `vendor/` is committed. After changing `go.mod`, run `make sync-vendor`.

Evidence: `//go:generate` in `internal/taxes/calculator.go`,
`pkg/nbggovge/currencies/generator/`, `scripts/sync-vendor.sh`.

## Naming and Documentation

- Go packages use lowercase names; Go files use lowercase snake_case names.
- Every package has a package comment; `main` packages use the
  `// <command> is ...` form. New exported identifiers MUST have doc comments;
  existing undocumented `String()` methods are tolerated (golangci-lint
  reports only new issues).
- Errors are sentinel values (`errors.New`) wrapped with
  `fmt.Errorf("...: %w", err)`. Logging uses `github.com/obalunenko/logger`
  carried in `context.Context`.
- Commits follow Conventional Commits (`feat`, `fix`, `chore`, `refactor`,
  `docs`, `ci`) with an optional scope; PRs are squash-merged into `master`.
- No feature-branch naming convention is established. Feature specs live in
  `specs/NNN-feature-name` through the feature-numbering hook.
- Update `README.md` when CLI commands, bot commands, flags, environment
  variables, or user-visible behavior change.

## Quality Gates and Development Tools

- Tooling runs in Docker through the Makefile and
  `deployments/docker-compose/go-tools-docker-compose.yml`
  (`ghcr.io/obalunenko/go-tools`).
- Before merge, these MUST pass: `make vet`, `make test`, `make lint-pipeline`,
  and `make build`. CI (`.github/workflows/go.yml`) runs the same targets.
  Without Docker, `go vet ./...` and `go test -race ./...` are the local
  equivalents.
- Code is formatted with gofmt and goimports
  (`make format-project`). New code MUST respect the limits in
  `.golangci.yml` (funlen 100 lines / 50 statements, gocyclo 15, lll 140, mnd,
  wsl_v5, gosec, exhaustive).
- Releases are tagged `vX.Y.Z` on `master` and built by goreleaser
  (`.goreleaser.yml`); the bot also ships as a Docker image
  (`build/docker/bot/Dockerfile`).
- Spec Kit is installed and checked locally through `make speckit-install`,
  `make speckit-update-check`, and `make speckit-update-test`. Keep Spec Kit
  checks out of CI.

## Known Deviations

These are current exceptions. They are not rules, and new features MUST NOT
extend them:

- The bot binary is named `ge-tax` in `.goreleaser.yml` but `ge-bot` in
  `build/docker/bot/Dockerfile` and README.
- `.golangci.yml` excludes `internal/input/` and `internal/puzzles/`, which do
  not exist in this repository.
- `depguard` rules are configured in `.golangci.yml`, but the linter is not
  enabled.
- `internal/service` starts a terminal spinner (`internal/spinner`), so
  presentation leaks into the service layer and also runs for bot requests.

## Governance

- This constitution supersedes conflicting templates and agent instructions.
  Plans MUST pass the Constitution Check in `plan-template.md`; any violation
  MUST be justified in its Complexity Tracking table.
- Amendments update this file and its Sync Impact Report, and propagate to the
  Spec Kit templates and `AGENTS.md`. Versioning follows semantic versioning:
  MAJOR for removed or redefined principles, MINOR for new principles or
  sections, PATCH for clarifications.
- Resolving a known deviation removes it from the list in a PATCH amendment.

**Version**: 1.0.3 | **Ratified**: 2026-11-14 | **Last Amended**: 2026-11-15
