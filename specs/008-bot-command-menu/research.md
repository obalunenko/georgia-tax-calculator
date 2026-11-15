# Research: Bot Command Menu

**Feature**: [spec.md](spec.md) | **Date**: 2026-11-15

No open unknowns remained after `/speckit-clarify`. The decisions below map
the spec onto the existing bot code and the vendored telego v1 API.

## R1. Which Bot API call sets the menu

- **Decision**: One `bot.SetMyCommands` call with the four commands and no
  `Scope` or `LanguageCode`. Telegram then stores them as the default list
  (`BotCommandScopeDefault`) for users of every language without a dedicated
  list. A new call replaces the earlier default list completely.
- **Rationale**: This is FR-001 and FR-005 in a single call. The call is
  idempotent, so restarts do not duplicate entries (SC-004).
- **Alternatives considered**:
  - `getMyCommands` first, then set only on a difference: one more call, and
    nothing is gained because setting the same list again is harmless.
  - `deleteMyCommands` before setting: not needed, because setting replaces
    the list.
  - Lists per scope or per language: out of scope (spec Assumptions).

## R2. One definition for the menu and the handlers

- **Decision**: A new file `cmd/bot/commands.go` holds an ordered table
  `menuCommands`. Each entry has a command name (the existing `cmd*`
  constants), a menu description, and a handler constructor
  (`func(*sessionStore) telegohandler.MessageHandler`; all four existing
  constructors already have this signature). `registerHandlers` registers
  `/start` explicitly and then one `CommandEqual` handler per table entry.
  `botCommands()` turns the table into `[]telego.BotCommand` for R1.
- **Rationale**: FR-003 says that command handlers are registered only from
  the menu and for `/start`. A single table plus one registration loop makes
  that true by construction: a command added to the table appears in both
  places, and no other command is registered. A test checks that every menu
  command gets its own answer (R6). A handler added outside the loop is not
  caught by a test; the single registration place makes such a change
  visible in review.
- **Alternatives considered**:
  - A separate menu list next to the existing `HandleMessage` calls, plus a
    test comparing the two: telegohandler predicates cannot be inspected, so
    the test could only send commands and would miss a handler that was
    added without a menu entry.
  - `/start` in the table with a "hidden" flag: an extra field for one
    exception. `/start` stays registered on its own instead, with a comment
    that says why.

## R3. When the menu is published and what a failure does

- **Decision**: `run` calls `setCommandMenu(ctx, bot)` synchronously after
  `GetMe` and before `UpdatesViaLongPolling`. The call runs under a
  `context.WithTimeout` of `setCommandsTimeout = 10 * time.Second`.
  `setCommandMenu` returns nothing: it logs success or failure, and `run`
  continues in both cases (FR-006). There is no retry; telego's default
  caller does not retry either.
- **Rationale**: The bot is ready to answer only once polling starts, and the
  menu is set just before that. Tests stay deterministic: by the time
  `getUpdates` is called, `setMyCommands` is already recorded. The timeout
  keeps a hung Bot API from blocking start. telego's default
  `FastHTTPCaller` uses `DoDeadline` when the context has a deadline
  (`vendor/github.com/mymmrac/telego/telegoapi/caller.go`). 10 s is far
  above a normal Bot API round trip and short enough for a restart loop.
- **Alternatives considered**:
  - A goroutine, as for the welcome-back broadcast: it does not speed up
    answering users in any way that matters for a single call, and tests
    would have to poll for it.
  - Failing `run`, and retries in the background: both rejected in the
    clarification session (spec Clarifications, Q3).

## R4. Log lines

- **Decision**:
  - Success: `log.FromContext(ctx).WithField("commands", 4).Info("bot: command menu set")`.
  - Failure: `log.WithError(ctx, err).Error("bot: failed to set command menu")`.
    The error includes telego's `setMyCommands` prefix and Telegram's
    description, or the context error on timeout.
- **Rationale**: These follow the existing `bot: ...` messages in
  `cmd/bot/bot.go` and meet FR-006 (reason, error level) and FR-007 (info
  level, number of commands).

## R5. Descriptions and Telegram limits

- **Decision**: The descriptions are fixed in the table, as listed in spec
  [Command Menu](spec.md#command-menu). A table-driven test checks every entry
  against the Bot API limits: the name has 1–32 characters from `a-z`, `0-9`
  and `_`; the description has 1–256 characters and no line break (FR-004).
- **Rationale**: Telegram rejects the whole call when one entry breaks a
  limit, and a rejection is only logged (R3). The test catches it before
  release.

## R6. How the behavior is tested

- **Decision** (all in `cmd/bot`, against the existing fake Telegram API):
  - `run` test: exactly one `setMyCommands` call, before `getUpdates`, whose
    `commands` parameter equals the four entries in order, with no `scope` and
    no `language_code` (FR-001, FR-002, FR-005).
  - Failure test: `fakeTelegram` gets a `methodErrors map[string]int` field
    that makes a method return a Bot API error. With `setMyCommands` failing,
    `run` still polls for updates and stops without an error (FR-006).
  - Logs: a recording fake `log.Logger` (mutex-guarded, test-only), placed in
    the context with `log.ContextWithLogger`. The tests check the error line
    and its error on failure, and the info line with `commands=4` on success
    (FR-006, FR-007, SC-003). `log.Init` is not used because it replaces the
    package-global logger.
  - Harness test: every menu command and `/start` is answered by its own
    handler, not by the "Please use /calculate or /convert to start." text
    fallback (FR-003, FR-008, SC-002).
  - Harness test: the command names in the `/start` and `/help` replies, in
    order of first appearance, equal the menu names (FR-010).
  - Table test for R5.
- **Rationale**: The constitution (VI) encourages `cmd/bot` tests, and the
  fake Telegram API already records every call and answers unknown methods
  with `true`, so the success path needs no change to it.

## R7. README

- **Decision**: In "Bot Commands", add a paragraph after the list. It says
  that the bot sets the Telegram command menu at every start with all
  commands except `/start`, that BotFather `/setcommands` is not needed, and
  that a list set there is replaced on the next start (FR-009). The list
  itself stays as it is.
