package main

import (
	"regexp"
	"slices"
	"testing"
	"unicode/utf8"

	"github.com/mymmrac/telego"
	"github.com/stretchr/testify/assert"
)

// Bot API limits of a command: https://core.telegram.org/bots/api#botcommand.
const maxCommandDescriptionLen = 256

var validCommandName = regexp.MustCompile(`^[a-z0-9_]{1,32}$`)

// wantMenu is the command menu of specs/008-bot-command-menu/contracts/command-menu.md.
var wantMenu = []telego.BotCommand{
	{Command: "calculate", Description: "Start the tax calculation flow"},
	{Command: "convert", Description: "Start the currency conversion flow"},
	{Command: "cancel", Description: "Cancel the current operation"},
	{Command: "help", Description: "Show available commands"},
}

func menuCommandNames() []string {
	names := make([]string, 0, len(menuCommands))

	for _, c := range menuCommands {
		names = append(names, c.name)
	}

	return names
}

func TestMenuCommands(t *testing.T) {
	assert.Equal(t, wantMenu, botCommands(), "the contract's menu, in order")
	assert.NotContains(t, menuCommandNames(), cmdStart, "Telegram's Start button sends /start")

	seen := make(map[string]bool, len(menuCommands))

	// Whether a description states the purpose of the command's /help entry is checked on review.
	for _, c := range menuCommands {
		t.Run(c.name, func(t *testing.T) {
			assert.Regexp(t, validCommandName, c.name)
			assert.False(t, seen[c.name], "duplicate command")

			seen[c.name] = true

			length := utf8.RuneCountInString(c.description)
			assert.GreaterOrEqual(t, length, 1)
			assert.LessOrEqual(t, length, maxCommandDescriptionLen)
			assert.NotContains(t, c.description, "\n", "a menu description is one line")
			assert.NotNil(t, c.handler)
		})
	}
}

func TestBotAnswersEveryCommand(t *testing.T) {
	h := newBotHarness(t, &fakeService{})

	commands := []string{"/" + cmdStart}

	for _, name := range menuCommandNames() {
		commands = append(commands, "/"+name)
	}

	for _, command := range commands {
		h.text(command)
		assert.NotEqual(t, "Please use /calculate or /convert to start.", h.api.lastText(t),
			"%s is answered by its own handler, not as text input", command)
	}
}

// TestBotCommandTexts checks that the /start and /help texts list exactly the menu commands, in menu order.
func TestBotCommandTexts(t *testing.T) {
	h := newBotHarness(t, &fakeService{})
	commandRef := regexp.MustCompile(`/([a-z0-9_]+)`)

	for _, command := range []string{"/" + cmdStart, "/" + cmdHelp} {
		h.text(command)

		var listed []string

		for _, match := range commandRef.FindAllStringSubmatch(h.api.lastText(t), -1) {
			if !slices.Contains(listed, match[1]) {
				listed = append(listed, match[1])
			}
		}

		assert.Equal(t, menuCommandNames(), listed, "%s lists the menu commands", command)
	}
}
