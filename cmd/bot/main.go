// bot is a Telegram bot for taxes calculations.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/obalunenko/logger"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ctx = log.ContextWithLogger(ctx, log.FromContext(ctx))

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.FromContext(ctx).Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	if err := run(ctx, token); err != nil {
		log.WithError(ctx, err).Fatal("Bot failed")
	}
}
