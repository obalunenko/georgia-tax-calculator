package main

import (
	"context"
	"time"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	log "github.com/obalunenko/logger"
)

// setCommandsTimeout bounds the call that publishes the command menu at start.
const setCommandsTimeout = 10 * time.Second

// menuCommand is an entry of the Telegram command menu and the handler that answers it.
type menuCommand struct {
	// name is the command without the slash.
	name string
	// description is shown next to the command in the menu.
	description string
	handler     func(store *sessionStore) telegohandler.MessageHandler
}

// menuCommands is the Telegram command menu, in menu order.
// The bot answers each entry and /start, which Telegram's Start button sends.
var menuCommands = []menuCommand{
	{name: cmdCalculate, description: "Start the tax calculation flow", handler: handleCalculate},
	{name: cmdConvert, description: "Start the currency conversion flow", handler: handleConvert},
	{name: cmdCancel, description: "Cancel the current operation", handler: handleCancel},
	{name: cmdHelp, description: "Show available commands", handler: handleHelp},
}

// botCommands returns the command menu as Bot API commands, in menu order.
func botCommands() []telego.BotCommand {
	commands := make([]telego.BotCommand, 0, len(menuCommands))

	for _, c := range menuCommands {
		commands = append(commands, telego.BotCommand{Command: c.name, Description: c.description})
	}

	return commands
}

// setCommandMenu publishes the command menu as the bot's default command list for all chats and languages,
// replacing the list set earlier. A failure is logged and does not stop the bot; the next start tries again.
func setCommandMenu(ctx context.Context, bot *telego.Bot) {
	callCtx, cancel := context.WithTimeout(ctx, setCommandsTimeout)
	defer cancel()

	if err := bot.SetMyCommands(callCtx, &telego.SetMyCommandsParams{Commands: botCommands()}); err != nil {
		log.WithError(ctx, err).Error("bot: failed to set command menu")

		return
	}

	log.FromContext(ctx).WithField("commands", len(menuCommands)).Info("bot: command menu set")
}
