package main

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"

	log "github.com/obalunenko/logger"
	"github.com/obalunenko/georgia-tax-calculator/internal/service"
)

// run starts the Telegram bot.
func run(ctx context.Context, token string) error {
	bot, err := telego.NewBot(token, telego.WithDefaultLogger(false, true))
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}

	me, err := bot.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("get me: %w", err)
	}

	log.FromContext(ctx).WithField("username", me.Username).Info("Bot started")

	updates, err := bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		return fmt.Errorf("start long polling: %w", err)
	}

	bh, err := telegohandler.NewBotHandler(bot, updates)
	if err != nil {
		return fmt.Errorf("create bot handler: %w", err)
	}

	store := newSessionStore()
	svc := service.New()

	registerHandlers(bh, store, svc)

	if err := bh.Start(); err != nil {
		return fmt.Errorf("bot handler: %w", err)
	}

	return nil
}

// registerHandlers registers all message and callback handlers.
func registerHandlers(bh *telegohandler.BotHandler, store *sessionStore, svc service.Service) {
	// Command handlers.
	bh.HandleMessage(handleStart(store), telegohandler.CommandEqual(cmdStart))
	bh.HandleMessage(handleHelp(store), telegohandler.CommandEqual(cmdHelp))
	bh.HandleMessage(handleCancel(store), telegohandler.CommandEqual(cmdCancel))
	bh.HandleMessage(handleCalculate(store), telegohandler.CommandEqual(cmdCalculate))
	bh.HandleMessage(handleConvert(store), telegohandler.CommandEqual(cmdConvert))

	// Text input handler (for amount fields).
	bh.HandleMessage(handleTextInput(store), telegohandler.AnyMessageWithText())

	// Callback query handler (for inline keyboard selections).
	bh.HandleCallbackQuery(handleCallback(store, svc), telegohandler.AnyCallbackQuery())
}
