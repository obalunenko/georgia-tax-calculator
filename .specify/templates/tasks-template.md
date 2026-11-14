---

description: "Task list template for Georgia Tax Calculator feature implementation"
---

# Tasks: [FEATURE NAME]

**Input**: Design documents from `/specs/[NNN-feature-name]/`

**Prerequisites**: plan.md (required), spec.md (required for user stories); research.md, data-model.md, quickstart.md, and contracts/ when relevant

**Tests**: Behavior changes in `internal/` and `pkg/` MUST include or update tests (constitution VI). Use Go `testing` with testify `assert`/`require`, table-driven `t.Run` cases, and fixtures in the package's `testdata/`. Tests for `cmd/bot`, `pkg/moneyutils`, and `internal/spinner` are encouraged but not required. Tests MUST NOT call the real NBG API unless gated by `INTEGRATION_TESTS=true`.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story. Within a story, follow package dependencies: `pkg/` → `internal/taxes`, `internal/converter` → `internal/service` → `cmd/ge-tax-calc` and `cmd/bot`.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies); it does not itself request spawning additional agents
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Interfaces: `cmd/ge-tax-calc/` (CLI) and `cmd/bot/` (Telegram bot)
- Domain: `internal/service/`, `internal/converter/`, `internal/taxes/`, `internal/models/`
- Libraries: `pkg/nbggovge/` (with `option/`, `mock/`, `currencies/`), `pkg/moneyutils/`, `pkg/dateutils/`
- Tests: `*_test.go` next to the code; fixtures in `<package>/testdata/`
- Generated: `internal/taxes/taxtype_string.go` (`make codegen`), `pkg/nbggovge/currencies/codes.gen.go` (manual: `go run ./pkg/nbggovge/currencies/generator -client -output pkg/nbggovge/currencies/codes.gen.go`)
- Docs: `README.md`, `pkg/nbggovge/README_CACHE.md`

<!--
  ============================================================================
  IMPORTANT: The tasks below are SAMPLE TASKS for illustration purposes only.

  The /speckit-tasks command MUST replace these with actual tasks based on:
  - User stories from spec.md (with their priorities P1, P2, P3...)
  - Feature requirements and affected packages from plan.md
  - Entities from data-model.md
  - Interface contracts from contracts/ (CLI commands, bot flows), if any

  Tasks MUST be organized by user story so each story can be:
  - Implemented independently
  - Tested independently
  - Delivered as an MVP increment

  Remove phases and examples unrelated to the feature. Never add databases,
  authentication, or web frontends without a feature requirement.

  DO NOT keep these sample tasks in the generated tasks.md file.
  ============================================================================
-->

## Phase 1: Feature Preparation

**Purpose**: Changes every story depends on; skip what the feature does not need

- [ ] T001 Add or update dependencies in `go.mod` and run `make sync-vendor`
- [ ] T002 [P] Add test fixtures in `[package]/testdata/` (e.g., NBG responses for the needed dates and currencies)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Domain changes that MUST be complete before any interface work

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

Examples of foundational tasks (adjust based on the plan):

- [ ] T003 Extend shared types in `internal/models/models.go`
- [ ] T004 [P] Add the tax category or rate in `internal/taxes/calculator.go`, then run `make codegen` to update `internal/taxes/taxtype_string.go`
- [ ] T005 [P] Extend `pkg/nbggovge` client or cache behavior in `pkg/nbggovge/[file].go`

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - [Title] (Priority: P1) 🎯 MVP

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own, e.g., a `go test` case plus a manual CLI or bot run]

### Tests for User Story 1

- [ ] T006 [P] [US1] Table-driven tests for [behavior] in `internal/service/service_test.go`
- [ ] T007 [P] [US1] Table-driven tests for [behavior] in `internal/[package]/[file]_test.go`

### Implementation for User Story 1

- [ ] T008 [US1] Implement [domain behavior] in `internal/[package]/[file].go`
- [ ] T009 [US1] Expose it through `internal/service/service.go` (depends on T008)
- [ ] T010 [P] [US1] Add the CLI flow in `cmd/ge-tax-calc/[file].go` (depends on T009)
- [ ] T011 [P] [US1] Add the bot flow in `cmd/bot/handlers.go` and `cmd/bot/session.go` (depends on T009)
- [ ] T012 [US1] Add validation and error wrapping (`fmt.Errorf("...: %w", err)`) for [inputs]

**Checkpoint**: At this point, User Story 1 should work in both the CLI and the bot and be testable independently

---

## Phase 4: User Story 2 - [Title] (Priority: P2)

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 2

- [ ] T013 [P] [US2] Table-driven tests for [behavior] in `[package]/[file]_test.go`

### Implementation for User Story 2

- [ ] T014 [US2] Implement [behavior] in `internal/[package]/[file].go`
- [ ] T015 [P] [US2] Update the CLI in `cmd/ge-tax-calc/[file].go`
- [ ] T016 [P] [US2] Update the bot in `cmd/bot/[file].go`

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

[Add more user story phases as needed, following the same pattern]

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Documentation and quality gates for the whole feature

- [ ] TXXX [P] Update `README.md` (CLI commands, bot commands, environment variables, behavior)
- [ ] TXXX Run `make codegen` if `TaxType` changed; run the currencies generator if currency codes changed
- [ ] TXXX Run `make format-project`
- [ ] TXXX Run `make vet` and `make lint-pipeline`
- [ ] TXXX Run `make test` (or `go test -race ./...` without Docker)
- [ ] TXXX Run `make test-integration` if NBG client behavior changed
- [ ] TXXX Run `make build`
- [ ] TXXX Run quickstart.md validation, if present

---

## Dependencies & Execution Order

### Phase Dependencies

- **Feature Preparation (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - Proceed sequentially in priority order (P1 → P2 → P3), or in parallel when stories touch different files
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May integrate with US1 but should be independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - May integrate with US1/US2 but should be independently testable

### Within Each User Story

- `pkg/` and `internal/taxes`, `internal/converter` before `internal/service`
- `internal/service` before `cmd/ge-tax-calc` and `cmd/bot`
- Tests alongside or before the code they cover; they MUST pass with `-race`
- CLI and bot changes for the same story can run in parallel
- Story complete before moving to next priority

### Parallel Opportunities

- Fixture and test tasks marked [P] in different packages can run in parallel
- CLI (`cmd/ge-tax-calc/`) and bot (`cmd/bot/`) tasks for the same story can run in parallel once the service change is done
- Different user stories can be worked on in parallel when they touch different files

---

## Parallel Example: User Story 1

```bash
# After the service change (T009), update both interfaces together:
Task: "Add the CLI flow in cmd/ge-tax-calc/[file].go"
Task: "Add the bot flow in cmd/bot/handlers.go and cmd/bot/session.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Feature Preparation
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Run the story's tests and try it in the CLI and the bot
5. Open a PR if ready

### Incremental Delivery

1. Complete Preparation + Foundational → Foundation ready
2. Add User Story 1 → Test independently → PR (MVP!)
3. Add User Story 2 → Test independently → PR
4. Each story adds value without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Do not hand-edit generated files or `vendor/`
- Commit with Conventional Commits (`feat:`, `fix:`, `refactor:`, ...)
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
