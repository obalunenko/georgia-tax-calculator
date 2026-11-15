# Project Instructions

Read `.specify/memory/constitution.md` for the project's architecture and
behavioral contracts. This repository is one Go module with two user
interfaces, the `ge-tax-calc` CLI and a Telegram bot, over a shared service
layer.

## Work Boundaries

- Interface work: `cmd/ge-tax-calc/` (urfave/cli v3, Bubble Tea v2) and
  `cmd/bot/` (telego). Interfaces call `internal/service`; they do not import
  `internal/converter` or `pkg/nbggovge` directly. Keep CLI and bot
  capabilities in parity unless a spec says otherwise. `internal/spinner/` is
  terminal presentation and belongs to this layer.
- Domain work: `internal/service/`, `internal/converter/`, `internal/taxes/`,
  and `internal/models/`. Money arithmetic goes through `pkg/moneyutils`.
- Library work: `pkg/nbggovge/`, `pkg/moneyutils/`, `pkg/dateutils/`. `pkg/`
  never imports `internal/`.
- Tooling work: `Makefile`, `scripts/` (except `scripts/speckit/`),
  `build/docker/`, `deployments/docker-compose/`, `.goreleaser.yml`,
  `.golangci*.yml`, `sonar-project.properties`, and `.github/`.
- Spec Kit work: `.specify/`, `.speckit/feature-numbering/`,
  `scripts/speckit/`, and generated Claude Code skills in `.claude/skills/`. The local feature-numbering
  extension source is `.speckit/feature-numbering/`. Regenerate skills through
  `make speckit-install`; do not edit generated copies. Preserve installed hooks
  and project customizations when refreshing tooling.
- Generated code: `internal/taxes/taxtype_string.go` comes from
  `make codegen` (stringer). `pkg/nbggovge/currencies/codes.gen.go` is
  regenerated manually with
  `go run ./pkg/nbggovge/currencies/generator -client -output pkg/nbggovge/currencies/codes.gen.go`.
  `vendor/` comes from `make sync-vendor`. Do not edit them by hand.
- User documentation: update `README.md` when commands, flags, bot commands,
  environment variables, or user-visible behavior change.

## Validation

- Makefile targets run tools in Docker: `make vet`, `make test`,
  `make lint-pipeline`, `make build`. CI runs the same targets.
- Without Docker, use `go vet ./...` and `go test -race ./...`.
- Format with `make format-project` (gofmt and goimports).
- Tests that call the real NBG API require `INTEGRATION_TESTS=true`
  (`make test-integration`); default tests use `pkg/nbggovge/mock` or
  `testdata/` fixtures.
- Spec Kit checks (`make speckit-update-check`, `make speckit-update-test`) are
  local; keep them out of CI.

## Observed Conventions

Lowercase Go package names, snake_case Go filenames, colocated `*_test.go`
files with table-driven testify tests, doc comments on exported identifiers,
sentinel errors wrapped with `%w`, and logging through
`github.com/obalunenko/logger` in context. Commits follow Conventional Commits
and PRs are squash-merged into `master`. No feature-branch naming convention or
numeric coverage threshold is established. Feature directories use
`specs/NNN-feature-name` through the installed numbering hook.
