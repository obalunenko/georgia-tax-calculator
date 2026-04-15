// bot is a Telegram bot for taxes calculations.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/obalunenko/logger"
)

const envBotLogLevel = "TELEGRAM_BOT_LOG_LEVEL"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger := log.Init(ctx, log.Params{
		Level: botLogLevel(),
	})

	ctx = log.ContextWithLogger(ctx, logger)

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.FromContext(ctx).Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	if err := run(ctx, token); err != nil {
		log.WithError(ctx, err).Fatal("Bot failed")
	}
}

func botLogLevel() string {
	lvl := os.Getenv(envBotLogLevel)
	if lvl == "" {
		return "info"
	}

	return lvl
}
