package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
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

	users := newUserStore(ctx, userStorePath())

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

// trackChat records chatID in the user store and logs new chats and save errors.
func trackChat(ctxLog log.Logger, users *userStore, chatID int64) {
	isNew, err := users.Track(chatID)
	if err != nil {
		ctxLog.WithError(err).Error("bot: failed to save user store")
	}

	if isNew {
		ctxLog.Info("bot: new user registered")
	}
}

// trackUserMsg wraps a MessageHandler to record the chat ID in the user store.
func trackUserMsg(users *userStore, h telegohandler.MessageHandler) telegohandler.MessageHandler {
	return func(ctx *telegohandler.Context, msg telego.Message) error {
		ctxLog := log.WithField(ctx.Context(), "chat_id", msg.Chat.ID).
			WithField("text", msg.Text)

		if msg.From != nil {
			ctxLog = ctxLog.WithField("user_id", msg.From.ID).
				WithField("username", msg.From.Username)
		}

		ctxLog.Debug("bot: user input")

		trackChat(ctxLog, users, msg.Chat.ID)

		return h(ctx, msg)
	}
}

// trackUserCallback wraps a CallbackQueryHandler to record the chat ID in the user store.
func trackUserCallback(users *userStore, h telegohandler.CallbackQueryHandler) telegohandler.CallbackQueryHandler {
	return func(ctx *telegohandler.Context, query telego.CallbackQuery) error {
		ctxLog := log.WithField(ctx.Context(), "user_id", query.From.ID).
			WithField("username", query.From.Username).
			WithField("callback_data", query.Data)

		ctxLog.Debug("bot: user input")

		if query.Message != nil {
			chatID := query.Message.GetChat().ID
			ctxLog = ctxLog.WithField("chat_id", chatID)

			trackChat(ctxLog, users, chatID)
		}

		return h(ctx, query)
	}
}

const msgInternalError = "❌ Something went wrong. Please use /cancel to start over."

// recoverPanics is a middleware that keeps the bot running when a handler panics:
// it logs the panic and tells the user that the request failed.
func recoverPanics(ctx *telegohandler.Context, update telego.Update) (err error) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}

		log.WithField(ctx.Context(), "panic", fmt.Sprint(recovered)).
			WithField("stack", string(debug.Stack())).
			Error("bot: handler panicked")

		chatID, ok := updateChatID(update)
		if !ok {
			return
		}

		_, err = sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: chatID},
			Text:   msgInternalError,
		})
	}()

	return ctx.Next(update)
}

// updateChatID returns the chat of a message or callback update.
func updateChatID(update telego.Update) (int64, bool) {
	switch {
	case update.Message != nil:
		return update.Message.Chat.ID, true
	case update.CallbackQuery != nil && update.CallbackQuery.Message != nil:
		return update.CallbackQuery.Message.GetChat().ID, true
	default:
		return 0, false
	}
}

// serializeMsg wraps a MessageHandler so that updates of one user are handled one at a time.
// Messages without a sender are ignored: handlers need the user ID for the session,
// and trackUserMsg has already logged the input.
func serializeMsg(store *sessionStore, h telegohandler.MessageHandler) telegohandler.MessageHandler {
	return func(ctx *telegohandler.Context, msg telego.Message) error {
		if msg.From == nil {
			return nil
		}

		unlock := store.lock(msg.From.ID)
		defer unlock()

		return h(ctx, msg)
	}
}

// serializeCallback wraps a CallbackQueryHandler so that updates of one user are handled one at a time.
func serializeCallback(store *sessionStore, h telegohandler.CallbackQueryHandler) telegohandler.CallbackQueryHandler {
	return func(ctx *telegohandler.Context, query telego.CallbackQuery) error {
		unlock := store.lock(query.From.ID)
		defer unlock()

		return h(ctx, query)
	}
}

// registerHandlers registers all message and callback handlers.
func registerHandlers(bh *telegohandler.BotHandler, store *sessionStore, svc service.Service, users *userStore) {
	// Middlewares apply only to handlers registered after them.
	bh.Use(recoverPanics)

	msgHandler := func(h telegohandler.MessageHandler) telegohandler.MessageHandler {
		return trackUserMsg(users, serializeMsg(store, h))
	}

	// Command handlers.
	bh.HandleMessage(msgHandler(handleStart(store)), telegohandler.CommandEqual(cmdStart))
	bh.HandleMessage(msgHandler(handleHelp(store)), telegohandler.CommandEqual(cmdHelp))
	bh.HandleMessage(msgHandler(handleCancel(store)), telegohandler.CommandEqual(cmdCancel))
	bh.HandleMessage(msgHandler(handleCalculate(store)), telegohandler.CommandEqual(cmdCalculate))
	bh.HandleMessage(msgHandler(handleConvert(store)), telegohandler.CommandEqual(cmdConvert))

	// Text input handler (for amount fields).
	bh.HandleMessage(msgHandler(handleTextInput(store)), telegohandler.AnyMessageWithText())

	// Callback query handler (for inline keyboard selections).
	bh.HandleCallbackQuery(
		trackUserCallback(users, answerCallback(serializeCallback(store, handleCallback(store, svc)))),
		telegohandler.AnyCallbackQuery(),
	)
}
