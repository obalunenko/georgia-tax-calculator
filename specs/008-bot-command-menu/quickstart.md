# Quickstart: Validate the Bot Command Menu

**Feature**: [spec.md](spec.md)

## Prerequisites

- Go 1.27.x; the repository root as working directory.
- For the manual runs: a test bot token in `TELEGRAM_BOT_TOKEN` and a Telegram
  app signed in to an account that can message the test bot. The automated
  tests need neither the token nor network access.

## Automated checks

```sh
go vet ./...
go test -race ./cmd/bot/...
make lint-pipeline
```

Expected: all pass. Covered cases (research R6):

- The menu table: four entries in the order of
  [contract](contracts/command-menu.md), names and descriptions within the
  Bot API limits, one line each.
- `run`: one `setMyCommands` call before `getUpdates`, with the four commands,
  no `scope`, and no `language_code`; info log `bot: command menu set` with
  `commands=4`.
- `run` with `setMyCommands` failing: it still polls, stops without an
  error, and logs `bot: failed to set command menu` with the error.
- Harness: `/start` and each of the four menu commands get their own reply,
  not the text fallback. The command names in the `/start` and `/help`
  replies equal the menu.

## Manual scenarios (`go run ./cmd/bot`)

| # | Setup | Action | Expected |
|---|-------|--------|----------|
| 1 | Test bot with no commands set | Start the bot; in the app, open the chat and type `/` | Four commands with descriptions, in contract order; no `/start` (SC-001) |
| 2 | Bot running | Type `/c` | Only `/calculate`, `/convert`, `/cancel` remain |
| 3 | Bot running | Pick each command from the list | The same reply as when you type it (SC-002) |
| 4 | Bot running; in a calculation | Pick `/cancel` | "✅ Operation cancelled. …" and the flow is reset |
| 5 | In BotFather, `/setcommands` for the test bot set to `start - Start` and `old - Old` | Restart the bot, then reopen the chat and type `/` | Exactly the four commands; no `/old` (US2, FR-005) |
| 6 | Bot restarted three times | Type `/` | Each command appears once (SC-004) |
| 7 | Bot running | Send `/start` by hand | The welcome message, as before |
| 8 | Bot log | Look at the start lines | `bot: command menu set` with `commands=4` after `Bot started` |

Telegram apps can cache the menu. If the list looks old, close and reopen the
chat.

The failure path (US3, SC-003) needs a Bot API that rejects the call, so
only the automated test covers it.
