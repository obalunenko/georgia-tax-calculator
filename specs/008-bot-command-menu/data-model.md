# Data Model: Bot Command Menu

**Feature**: [spec.md](spec.md) | **Research**: [research.md](research.md)

The feature stores no data. The only structure is a fixed, in-code table in
`cmd/bot/commands.go` (research R2). Telegram keeps the published copy.

## Menu command (`menuCommand`, unexported)

| Field | Meaning | Rules |
|-------|---------|-------|
| `name` | Command without the slash, one of the existing `cmd*` constants | 1–32 characters from `a-z`, `0-9`, `_`; unique in the table |
| `description` | Text shown next to the command in the menu | 1–256 characters, one line, English; same purpose as the `/help` entry (FR-004) |
| `handler` | Constructor of the command's message handler, `func(*sessionStore) telegohandler.MessageHandler` | One of the existing `handleCalculate`, `handleConvert`, `handleCancel`, `handleHelp` |

## Menu (`menuCommands`, unexported)

An ordered table, and its order is the menu order (FR-002):

| # | `name` | `description` | `handler` |
|---|--------|---------------|-----------|
| 1 | `calculate` | Start the tax calculation flow | `handleCalculate` |
| 2 | `convert` | Start the currency conversion flow | `handleConvert` |
| 3 | `cancel` | Cancel the current operation | `handleCancel` |
| 4 | `help` | Show available commands | `handleHelp` |

The table does not contain `/start`. `registerHandlers` registers it on its
own, so every command the bot answers is either `/start` or an entry of the
table (FR-003).

## Derived values

- `botCommands() []telego.BotCommand`: the table as Bot API values, in table
  order: `Command` = `name`, `Description` = `description`. It is the body of
  the `setMyCommands` call ([contract](contracts/command-menu.md)).

## Lifecycle

```text
bot start → getMe ok → setMyCommands(default list) ─┬─ ok    → info log → long polling
                                                     └─ error → error log → long polling
```

Telegram keeps the last list it accepted, also while the bot is offline. The
next start replaces it (FR-005). The bot never deletes it.
