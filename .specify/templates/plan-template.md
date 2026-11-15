# Implementation Plan: [FEATURE]

**Feature Directory**: `specs/[NNN-feature-name]` | **Branch**: [Current branch, if applicable]

**Date**: [DATE] | **Spec**: [link]

**Input**: Feature specification from `/specs/[NNN-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

[Extract from feature spec: primary requirement + technical approach from research]

## Technical Context

<!--
  Defaults below describe the current project. Keep what applies, and add
  feature-specific details in the brackets.
-->

**Language/Version**: Go 1.27.3, single module
`github.com/obalunenko/georgia-tax-calculator` with committed `vendor/`.

**Primary Dependencies**: urfave/cli v3 and Bubble Tea v2 (CLI), telego (bot),
shopspring/decimal (via `pkg/moneyutils`), obalunenko/logger, getenv, version.
[List feature-specific dependency changes; they require `make sync-vendor`.]

**Storage**: None for calculations. The bot keeps sessions in memory and
persists chat IDs to a JSON file (`USER_STORE_PATH`). NBG rates are cached in
memory. [Affected storage, or N/A.]

**Testing**: Go `testing` with testify `assert`/`require`, table-driven
`t.Run` cases, `testdata/` JSON fixtures, `pkg/nbggovge/mock`; race detector
on. Real-API tests only under `INTEGRATION_TESTS=true`. [Relevant test files.]

**Target Platform**: CLI on linux/darwin/windows; bot on linux/darwin and as a
Docker image. CGO disabled. [Feature-specific platform scope.]

**Project Type**: Interactive CLI and Telegram bot over a shared service layer,
plus reusable libraries in `pkg/`.

**Performance Goals**: [Feature-specific measurable goals or N/A; no general
latency or throughput goal is established.]

**Constraints**: Money math through `pkg/moneyutils`; NBG access through
`nbggovge.Client`; CLI/bot parity; no network in default tests.
[Explain applicable constraints.]

**Scale/Scope**: [Feature scope: affected commands, flows, categories,
currencies.]

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [ ] I. Interfaces (`cmd/ge-tax-calc`, `cmd/bot`) reach domain logic only
      through `internal/service`; no direct `internal/converter` or
      `pkg/nbggovge` imports.
- [ ] II. CLI and bot stay in parity, or the spec justifies a one-interface
      scope.
- [ ] III. Money arithmetic uses `pkg/moneyutils`; rates use the NBG rate for
      the income date; `TaxType` changes update enum, rates, generated code,
      and tests together.
- [ ] IV. NBG access goes through `nbggovge.Client`; default tests use mocks
      or fixtures; real-API tests are gated by `INTEGRATION_TESTS`.
- [ ] V. `pkg/` does not import `internal/`.
- [ ] VI. Behavior changes in `internal/` and `pkg/` include table-driven
      testify tests that pass with `-race`.
- [ ] VII. Generated code is regenerated, not hand-edited (`make codegen`
      for `TaxType`, the currencies generator for currency codes); `vendor/`
      is synced after `go.mod` changes.
- [ ] Known deviations from the constitution are not extended.

## Project Structure

### Documentation (this feature)

```text
specs/[NNN-feature-name]/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)
<!--
  Keep only the paths this feature touches and name the files that change.
  Tests live next to the code in *_test.go files.
-->

```text
cmd/
├── ge-tax-calc/        # CLI: commands, Bubble Tea menus, prompts
└── bot/                # Telegram bot: handlers, sessions, keyboards, user store
internal/
├── service/            # Domain facade: Calculate, Convert
├── converter/          # Conversion to GEL at NBG rates (+ testdata/)
├── taxes/              # TaxType enum, rates, calculation (stringer output)
├── models/             # Shared domain types (Money)
└── spinner/            # Terminal spinner
pkg/
├── nbggovge/           # NBG API client, cache, option/, mock/, currencies/
├── moneyutils/         # Decimal-backed money arithmetic
└── dateutils/          # Date parsing and formatting
build/docker/           # Bot image
deployments/docker-compose/
scripts/                # Build, test, lint, codegen, release scripts
tools/                  # Pinned tools, one go.mod per tool (go tool)
```

**Structure Decision**: [Name the affected packages and files, and explain any
new package or cross-package dependency.]

**Validation**: `make vet`, `make test`, `make lint-pipeline`, `make build`.
Add
`make codegen` when `TaxType` changes, the currencies generator (constitution
VII) when currency codes change, `make sync-vendor` when dependencies change,
and `make test-integration` when NBG client behavior changes.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., bot-only capability] | [current need] | [why CLI parity is not feasible now] |
| [e.g., direct `nbggovge` import in `cmd/`] | [specific problem] | [why going through `internal/service` is insufficient] |
