---

description: "Task list for the Bot Command Menu feature"
---

# Tasks: Bot Command Menu

**Input**: Design documents from `/specs/008-bot-command-menu/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/command-menu.md](contracts/command-menu.md), [quickstart.md](quickstart.md)

**Tests**: Included. The spec requires automated tests (FR-003, FR-010), and
research R6 lists them. All changes are in `cmd/bot` (package `main`), where
the constitution (VI) encourages tests. They use Go `testing` with testify,
the fake Telegram API in `cmd/bot/fake_telegram_test.go`, and the bot harness
(`newBotHarness`). They need no network and MUST pass with `-race`.

**Organization**: Tasks are grouped by user story. Only `cmd/bot/` and
`README.md` change; `internal/`, `pkg/`, `cmd/ge-tax-calc/`, `go.mod`, and
generated code do not.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies); it does not itself request spawning additional agents
- **[Story]**: Which user story this task belongs to (US1, US2, US3)

---

## Phase 1: Foundational (Blocking Prerequisites)

**Purpose**: The single command table that both the handlers and the menu
use (research R2). There is no setup phase: no dependency, fixture, or
generated-code change is needed.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T001 Create `cmd/bot/commands.go` (package `main`) with an unexported `menuCommand` struct and an ordered package-level `menuCommands []menuCommand` table, per [data-model.md](data-model.md). Fields: `name string` ("Command without the slash, one of the existing `cmd*` constants"; rule: "1–32 characters from `a-z`, `0-9`, `_`; unique in the table"), `description string` (rule: "1–256 characters, one line, English; same purpose as the `/help` entry (FR-004)"), and `handler func(*sessionStore) telegohandler.MessageHandler`. Entries, in this order: `{cmdCalculate, "Start the tax calculation flow", handleCalculate}`, `{cmdConvert, "Start the currency conversion flow", handleConvert}`, `{cmdCancel, "Cancel the current operation", handleCancel}`, `{cmdHelp, "Show available commands", handleHelp}`. Do not add `/start`. Doc comments: the table is the Telegram command menu in menu order, and the bot answers each entry plus `/start`, which Telegram's Start button sends. Keep the `cmd*` constants and handlers in `cmd/bot/handlers.go`.
- [x] T002 In `registerHandlers` in `cmd/bot/bot.go`, replace the five `bh.HandleMessage(..., telegohandler.CommandEqual(cmdX))` lines. Register `/start` first with `bh.HandleMessage(msgHandler(handleStart(store)), telegohandler.CommandEqual(cmdStart))` and a comment saying it is not in the menu because Telegram's Start button sends it. Then loop `for _, c := range menuCommands { bh.HandleMessage(msgHandler(c.handler(store)), telegohandler.CommandEqual(c.name)) }`. Keep the text input and callback handlers after the commands, unchanged. After this change, `registerHandlers` is the only place that registers command handlers, and it uses only `menuCommands` and `cmdStart` (FR-003) (depends on T001).
- [x] T003 Create `cmd/bot/commands_test.go` with `TestMenuCommands`. It asserts that the names in `menuCommands` equal `[]string{"calculate", "convert", "cancel", "help"}` in order (FR-002) and that `start` is not among them. It also runs a table-driven `t.Run` per entry (research R5): the name matches `^[a-z0-9_]{1,32}$`; names are unique; the description is 1–256 characters (`utf8.RuneCountInString`) and contains no `\n`; `handler` is not nil. Whether a description states the same purpose as the `/help` entry is checked on review, not by this test (FR-004) (depends on T001).
- [x] T004 Add `TestBotAnswersEveryCommand` to `cmd/bot/commands_test.go`, using `newBotHarness(t, &fakeService{})`. For `/start` and for `"/" + c.name` of every entry in `menuCommands`, send the command with `h.text(...)` and assert that `h.api.lastText(t)` is not `"Please use /calculate or /convert to start."` (the text fallback). The bot must answer each one with its own handler (FR-003, FR-008, SC-002) (depends on T002, T003).

**Checkpoint**: `go test -race ./cmd/bot/...` passes. The bot answers the same five commands as before, now registered from the table.

---

## Phase 2: User Story 1 - Pick a command from the suggestions (Priority: P1) 🎯 MVP

**Goal**: At start the bot publishes the four commands with their
descriptions, so Telegram suggests them when the user types `/`.

**Independent Test**: `go test -race ./cmd/bot/ -run 'TestRun|TestBotCommandTexts'`. Then start the bot with a test token, type `/` in its chat, see the four commands in [contract](contracts/command-menu.md) order without `/start`, and pick `/convert` (quickstart scenarios 1–4).

### Implementation for User Story 1

- [x] T005 [US1] Add these to `cmd/bot/commands.go` (research R1, R3, R4):
  - `const setCommandsTimeout = 10 * time.Second`.
  - `botCommands() []telego.BotCommand`: maps `menuCommands` in order to `telego.BotCommand{Command: c.name, Description: c.description}`.
  - `setCommandMenu(ctx context.Context, bot *telego.Bot)`, which returns nothing. It derives `context.WithTimeout(ctx, setCommandsTimeout)` and calls `bot.SetMyCommands(callCtx, &telego.SetMyCommandsParams{Commands: botCommands()})`, leaving `Scope` and `LanguageCode` unset (default list for all chats and languages, FR-001). On error it logs `log.WithError(ctx, err).Error("bot: failed to set command menu")` and returns, with no retry (FR-006). On success it logs `log.FromContext(ctx).WithField("commands", len(menuCommands)).Info("bot: command menu set")` (FR-007).
  - Doc comments on all three, matching [contracts/command-menu.md](contracts/command-menu.md) (depends on T001).
- [x] T006 [US1] In `run` in `cmd/bot/bot.go`, call `setCommandMenu(ctx, bot)` right after the `"Bot started"` log line, before `newUserStore` and `bot.UpdatesViaLongPolling`. Do not change how `run` handles errors: a failed publish never makes `run` return an error (depends on T005).

### Tests for User Story 1

- [x] T007 [P] [US1] In `TestRunStartsAndShutsDown` in `cmd/bot/run_test.go`, assert `require.Len(t, api.callsOf("setMyCommands"), 1)`. Map the call's `params["commands"]` (use `asSlice`/`asMap`/`asString` from `cmd/bot/fake_telegram_test.go`) to command/description pairs and assert they equal the four entries of [contracts/command-menu.md](contracts/command-menu.md) in order. Also assert that the `setMyCommands` call comes before the first `getUpdates` call: add a `methods() []string` helper to `fakeTelegram` in `cmd/bot/fake_telegram_test.go` that returns the recorded method names in call order. Finally, assert `assert.Empty(t, api.callsOf("deleteMyCommands"))`: the bot never deletes the menu, including at shutdown (data-model Lifecycle) (depends on T006).
- [x] T008 [P] [US1] Add `TestBotCommandTexts` to `cmd/bot/commands_test.go` (FR-010). Send `/start` and then `/help` through `newBotHarness`. For each reply, extract the command names with `regexp.MustCompile("/([a-z0-9_]+)")`, keep the order of first appearance, and drop duplicates. Assert the result equals the names of `menuCommands` in order. Do not change the texts in `cmd/bot/handlers.go` (FR-008) (depends on T004).

**Checkpoint**: US1 is complete. The menu is published at every start, and the start and help texts match it.

---

## Phase 3: User Story 2 - Deploy the bot without setting the menu by hand (Priority: P2)

**Goal**: The operator needs no BotFather step. Every start replaces the
default list, and restarts do not duplicate entries.

**Independent Test**: `go test -race ./cmd/bot/ -run TestRunSetsDefaultCommandList`. Then quickstart scenarios 5–6: set `/old` in BotFather, restart the bot, and see exactly the four commands.

### Tests for User Story 2

- [x] T009 [US2] Add `TestRunSetsDefaultCommandList` to `cmd/bot/run_test.go`. Run `run` twice in sequence, each run with its own `fakeTelegram` and the same `t.TempDir()` `USER_STORE_PATH`. For each run, wait with `require.Eventually` until that run's fake has a `getUpdates` call, then cancel and wait for `run` to return nil, as in `TestRunStartsAndShutsDown`. Assert that each fake has exactly one `setMyCommands` call, that the `commands` params of the two calls are equal (SC-004), and that neither call has a `scope` or `language_code` key in its params (FR-001, FR-005). Separate fakes keep the runs independent: a `getUpdates` call recorded by the first run cannot end the second run early (depends on T007).

### Implementation for User Story 2

- [x] T010 [P] [US2] In `README.md`, section "Bot Commands", keep the command list as it is and add, after it, the paragraph from [contracts/command-menu.md](contracts/command-menu.md#readme-readmemd-bot-commands). It says the bot sets the list, except `/start`, as its Telegram command menu at every start, so Telegram suggests the commands when you type `/`; BotFather `/setcommands` is not needed; and a list set there is replaced on the next start (FR-009).

**Checkpoint**: US2 is complete. The default-list contract is tested and documented for operators.

---

## Phase 4: User Story 3 - The bot works when the menu cannot be set (Priority: P3)

**Goal**: If the publish fails, the bot logs the reason and runs as usual.
After a successful publish the log says how many commands the menu has.

**Independent Test**: `go test -race ./cmd/bot/ -run 'TestRunKeepsRunningWhenCommandMenuFails|TestRunStartsAndShutsDown'` (spec US3, SC-003).

### Test support for User Story 3

- [x] T011 [US3] Add a `methodErrors map[string]int` field to `fakeTelegram` in `cmd/bot/fake_telegram_test.go`, documented like `sendErrors`. In `Call`, after recording the call and before the `switch`, a method listed in `methodErrors` gets `&telegoapi.Response{Ok: false, Error: &telegoapi.Error{ErrorCode: code, Description: "Bad Request"}}`.
- [x] T012 [US3] Add a test-only recording logger to `cmd/bot/fake_telegram_test.go`. `logRecorder` holds a mutex and a slice of `logEntry{level, msg string; err error; fields map[string]any}`. `recordingLogger` implements `log.Logger` (`github.com/obalunenko/logger`): `Debug/Info/Warn/Error/Fatal` append an entry with their level; `WithError`, `WithField`, and `WithFields` return a copy with the error or fields added that shares the same recorder; `Fatal` only records. Add `newLogContext(t) (context.Context, *logRecorder)`, which builds the context with `log.ContextWithLogger(context.Background(), ...)`, and an `entries(msg string) []logEntry` lookup. Do not call `log.Init`, which replaces the package-global logger (research R6) (depends on T011).

### Tests for User Story 3

- [x] T013 [US3] Add `TestRunKeepsRunningWhenCommandMenuFails` to `cmd/bot/run_test.go`. Use `fakeTelegram{methodErrors: map[string]int{"setMyCommands": http.StatusBadRequest}}`, a `t.TempDir()` `USER_STORE_PATH`, and the context from `newLogContext`. Start `run` with `telego.WithAPICaller(api), telego.WithDiscardLogger()`. Wait with `require.Eventually` until `getUpdates` has been called, then cancel; `run` must return nil (FR-006). Assert one `bot: failed to set command menu` entry at level `error` whose `err` contains `setMyCommands`, and no `bot: command menu set` entry. Reaching `getUpdates` shows that the handlers are registered and polling has started; T004 already covers the replies to typed commands (US3 scenario 2) (depends on T006, T012).
- [x] T014 [US3] In `TestRunStartsAndShutsDown` in `cmd/bot/run_test.go`, derive the run context from `newLogContext` and assert one `bot: command menu set` entry at level `info` with `fields["commands"] == 4`, and no `bot: failed to set command menu` entry (FR-007) (depends on T007, T012).

**Checkpoint**: All three stories work and are tested.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Quality gates for the whole feature

- [x] T015 Run `make format-project` (gofmt and goimports) over `cmd/bot/`
- [x] T016 Run `make vet` and `make lint-pipeline`. Fix findings in `cmd/bot/commands.go`, `cmd/bot/bot.go`, and the test files (funlen, lll 140, wsl_v5, mnd, revive doc comments)
- [x] T017 Run `make test` (or `go test -race ./...` without Docker)
- [x] T018 Run `make build` (both binaries; the bot image is unaffected except for the binary)
- [x] T019 Run the manual scenarios 1–8 in [quickstart.md](quickstart.md) with a test bot token, and record the result in the spec status

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies. BLOCKS all user stories.
- **US1 (Phase 2)**: Depends on Phase 1.
- **US2 (Phase 3)**: The test T009 depends on US1 (T007). The README task T010 depends only on the contract and can be done at any time.
- **US3 (Phase 4)**: Depends on US1 (T006 for the behavior, T007 for the shared run test).
- **Polish (Phase 5)**: Depends on all stories.

### Task Graph

```text
T001 ─┬─ T002 ── T004 ── T008
      ├─ T003 ── T004
      └─ T005 ── T006 ─┬─ T007 ─┬─ T009
                       │        └─ T014
                       └─ T013
T011 ── T012 ─┬─ T013
              └─ T014
T010 (independent)
T015 → T016 → T017 → T018 → T019 (after all)
```

### Within Each User Story

- Production code in `cmd/bot/commands.go` before `cmd/bot/bot.go`, and both before the tests that run them.
- `cmd/bot/run_test.go` tasks (T007, T009, T013, T014) touch one file: run them in order.
- `cmd/bot/commands_test.go` tasks (T003, T004, T008) touch one file: run them in order.
- `cmd/bot/fake_telegram_test.go` tasks (T007 for the `methods()` helper, T011, T012) touch one file: do not run them at the same time.

### Parallel Opportunities

- T007 (`run_test.go`) and T008 (`commands_test.go`) once T006 and T004 are done.
- T010 (`README.md`) with any other task.
- T011–T012 (`fake_telegram_test.go`) with T005–T006 (`commands.go`, `bot.go`). They must not overlap T007, which also edits `fake_telegram_test.go`.

---

## Parallel Example: User Story 1

```bash
# After T006 (publish wired into run) and T004 (harness test file exists):
Task: "T007 Assert the setMyCommands call in cmd/bot/run_test.go"
Task: "T008 Add TestBotCommandTexts in cmd/bot/commands_test.go"
# Any time:
Task: "T010 Add the menu paragraph to README.md"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 (T001–T004): table and registration, behavior unchanged.
2. Complete Phase 2 (T005–T008): the menu is published at start.
3. **STOP and VALIDATE**: `go test -race ./cmd/bot/...` and quickstart scenarios 1–4 with a test bot.

### Incremental Delivery

1. Foundational + US1 → the menu appears (MVP).
2. US2 → default-list contract tested, README updated.
3. US3 → failure path and log lines tested.
4. Polish → quality gates, manual quickstart, then one PR with a `feat(bot):` Conventional Commit.

---

## Notes

- [P] tasks = different files, no dependencies.
- `setCommandMenu`'s failure handling (FR-006) is written in T005 because it is part of the same function; US3 adds the tests that prove it.
- SC-003 is covered by two tests together: T013 shows that polling starts after a failed publish, and T004 shows that each command gets its own reply. A single end-to-end test is not possible because the fake `getUpdates` does not deliver updates to `run`.
- The 10 s timeout is not tested by waiting: telego's default caller honors the context deadline (research R3), and a waiting test would add 10 s to the suite.
- Do not hand-edit generated files or `vendor/`; no `make codegen` or `make sync-vendor` is needed.

---

## Phase 6: Convergence

- [x] T020 Replay quickstart scenarios 5–6 against the test bot through the Bot API: set the default list to `start - Start` and `old - Old` with `setMyCommands`, as BotFather `/setcommands` would; start `ge-bot` with an empty `USER_STORE_PATH`; check that `getMyCommands` returns exactly the four commands of `contracts/command-menu.md`; restart the bot three times and check again that each command appears once; record the result in the spec status per US2/AC2, FR-005, SC-004 (partial)
- [x] T021 Have the developer run quickstart scenarios 2–4 and 7 in a Telegram app with the test bot (type `/`, type `/c`, pick each command, pick `/cancel` during a calculation, send `/start` by hand) and record the result in the spec status; this needs a person with the Telegram app per SC-001, SC-002, US1/AC1–AC4 (partial)
