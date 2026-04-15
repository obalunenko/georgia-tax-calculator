package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	log "github.com/obalunenko/logger"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
)

const shutdownBroadcastTimeout = 30 * time.Second

// userStorePath returns the path for the persisted user store file.
// It can be overridden with the USER_STORE_PATH environment variable.
func userStorePath() string {
	if p := os.Getenv("USER_STORE_PATH"); p != "" {
		return p
	}

	return "users.json"
}

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

	users := newUserStore(userStorePath())

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

	registerHandlers(bh, store, svc, users)

	// Broadcast welcome-back message to all previously known users.
	if knownUsers := users.All(); len(knownUsers) > 0 {
		go broadcast(ctx, bot, knownUsers, msgWelcomeBack)
	}

	// Graceful shutdown: broadcast maintenance message before the process exits.
	shutdownDone := make(chan struct{})

	go func() {
		defer close(shutdownDone)

		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownBroadcastTimeout)
		defer cancel()

		broadcast(shutdownCtx, bot, users.All(), msgMaintenance)
	}()

	if err = bh.Start(); err != nil {
		return fmt.Errorf("bot handler: %w", err)
	}

	// Wait for shutdown broadcast to finish before returning.
	<-shutdownDone

	return nil
}

// trackUserMsg wraps a MessageHandler to record the chat ID in the user store.
func trackUserMsg(users *userStore, h telegohandler.MessageHandler) telegohandler.MessageHandler {
	return func(ctx *telegohandler.Context, msg telego.Message) error {
		users.Track(msg.Chat.ID)

		return h(ctx, msg)
	}
}

// trackUserCallback wraps a CallbackQueryHandler to record the chat ID in the user store.
func trackUserCallback(users *userStore, h telegohandler.CallbackQueryHandler) telegohandler.CallbackQueryHandler {
	return func(ctx *telegohandler.Context, query telego.CallbackQuery) error {
		if query.Message != nil {
			users.Track(query.Message.GetChat().ID)
		}

		return h(ctx, query)
	}
}

// registerHandlers registers all message and callback handlers.
func registerHandlers(bh *telegohandler.BotHandler, store *sessionStore, svc service.Service, users *userStore) {
	// Command handlers.
	bh.HandleMessage(trackUserMsg(users, handleStart(store)), telegohandler.CommandEqual(cmdStart))
	bh.HandleMessage(trackUserMsg(users, handleHelp(store)), telegohandler.CommandEqual(cmdHelp))
	bh.HandleMessage(trackUserMsg(users, handleCancel(store)), telegohandler.CommandEqual(cmdCancel))
	bh.HandleMessage(trackUserMsg(users, handleCalculate(store)), telegohandler.CommandEqual(cmdCalculate))
	bh.HandleMessage(trackUserMsg(users, handleConvert(store)), telegohandler.CommandEqual(cmdConvert))

	// Text input handler (for amount fields).
	bh.HandleMessage(trackUserMsg(users, handleTextInput(store)), telegohandler.AnyMessageWithText())

	// Callback query handler (for inline keyboard selections).
	bh.HandleCallbackQuery(trackUserCallback(users, handleCallback(store, svc)), telegohandler.AnyCallbackQuery())
}
