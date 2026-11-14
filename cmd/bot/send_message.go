package main

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/mymmrac/telego"
	log "github.com/obalunenko/logger"
)

// sendMessage wraps Bot.SendMessage with a debug log for outgoing bot responses.
// The text is not logged: results contain the user's amounts.
func sendMessage(
	ctx context.Context,
	bot *telego.Bot,
	params *telego.SendMessageParams,
) (*telego.Message, error) {
	log.WithField(ctx, "chat_id", params.ChatID.String()).
		WithField("text_len", utf8.RuneCountInString(params.Text)).
		Debug("bot: response")

	return bot.SendMessage(ctx, params)
}

// inputLogField describes user input for logs without its content: amounts are personal data.
// Commands are logged as is, other text only by length.
func inputLogField(text string) (key string, value any) {
	if strings.HasPrefix(text, "/") {
		command, _, _ := strings.Cut(text, " ")

		return "command", command
	}

	return "text_len", utf8.RuneCountInString(text)
}
