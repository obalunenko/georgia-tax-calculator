# Contract: Bot Command Menu

**Feature**: [spec.md](../spec.md) | **Data model**: [data-model.md](../data-model.md)

## Bot API request at start

Sent once per start, after `getMe` succeeds and before the first
`getUpdates` (research R3). Timeout: 10 s.

```http
POST https://api.telegram.org/bot<token>/setMyCommands
Content-Type: application/json

{
  "commands": [
    {"command": "calculate", "description": "Start the tax calculation flow"},
    {"command": "convert",   "description": "Start the currency conversion flow"},
    {"command": "cancel",    "description": "Cancel the current operation"},
    {"command": "help",      "description": "Show available commands"}
  ]
}
```

- No `scope` field: the list is the default list for all chats (FR-001).
- No `language_code` field: the list applies to users of every language that
  has no dedicated list (FR-001, spec Assumptions).
- `/start` is not in the list (FR-002).
- The expected answer is `{"ok": true, "result": true}`. Any other answer,
  a transport error, or the timeout counts as a failure.

## Log lines

| Case | Level | Message | Fields |
|------|-------|---------|--------|
| Accepted | info | `bot: command menu set` | `commands`: `4` |
| Failed | error | `bot: failed to set command menu` | `error`: telego's error, for example `telego: setMyCommands: api: 400 "Bad Request: ..."`, or on timeout an error that contains `context deadline exceeded` |

The bot keeps starting in both cases (FR-006), and it does not retry before
the next start.

## What users see

The Telegram app shows the list when the user types `/` or opens the Menu
button:

```text
/calculate   Start the tax calculation flow
/convert     Start the currency conversion flow
/cancel      Cancel the current operation
/help        Show available commands
```

Unchanged: the reply to each command, `/start` and the Start button, the
handling of unknown commands, and the texts of `/start` and `/help`. The
command names in those two texts are exactly the four above, in this order
(FR-010).

## README (`README.md`, "Bot Commands")

The command list stays as it is. A paragraph follows it with this content
(FR-009):

> The bot sets this list, except `/start`, as its Telegram command menu at
> every start, so Telegram suggests the commands when you type `/`. You do
> not need BotFather's `/setcommands`; a list set there is replaced on the
> next start.
