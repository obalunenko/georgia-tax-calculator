package main

import (
	"context"

	"github.com/mymmrac/telego"
	log "github.com/obalunenko/logger"
)

// sendMessage wraps Bot.SendMessage with a debug log for outgoing bot responses.
func sendMessage(
	ctx context.Context,
	bot *telego.Bot,
	params *telego.SendMessageParams,
) (*telego.Message, error) {
	log.WithField(ctx, "chat_id", params.ChatID.String()).
		WithField("text", params.Text).
		Debug("bot: response")

	return bot.SendMessage(ctx, params)
}
