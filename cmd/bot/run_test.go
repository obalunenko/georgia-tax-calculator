package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/mymmrac/telego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// publishedMenu returns the commands of the only setMyCommands call made to api.
func publishedMenu(t *testing.T, api *fakeTelegram) []telego.BotCommand {
	t.Helper()

	calls := api.callsOf("setMyCommands")
	require.Len(t, calls, 1, "the command menu is published once per start")

	var commands []telego.BotCommand

	for _, c := range asSlice(calls[0].params["commands"]) {
		command := asMap(c)
		commands = append(commands, telego.BotCommand{
			Command:     asString(command["command"]),
			Description: asString(command["description"]),
		})
	}

	return commands
}

func TestRunStartsAndShutsDown(t *testing.T) {
	const (
		knownChat   = int64(42)
		blockedChat = int64(43)
	)

	path := filepath.Join(t.TempDir(), "users.json")
	require.NoError(t, os.WriteFile(path, []byte("[42,43]"), 0o600))
	t.Setenv("USER_STORE_PATH", path)

	api := &fakeTelegram{sendErrors: map[int64]int{blockedChat: http.StatusForbidden}}

	logCtx, logs := newLogContext()

	ctx, cancel := context.WithCancel(logCtx)

	stopped := make(chan error, 1)

	go func() { stopped <- run(ctx, testToken, telego.WithAPICaller(api), telego.WithDiscardLogger()) }()

	// Wait until the welcome-back broadcast has reached both chats (in any order) and dropped the one
	// that answered 403; canceling earlier would stop the broadcast half-way.
	require.Eventually(t, func() bool {
		welcomes := 0

		for _, text := range api.texts() {
			if text == msgWelcomeBack {
				welcomes++
			}
		}

		data, err := os.ReadFile(path)

		return welcomes == 2 && err == nil && string(data) == "[42]" && len(api.callsOf("getUpdates")) > 0
	}, 5*time.Second, 10*time.Millisecond, "the bot polls for updates and says it is back online")

	cancel()

	select {
	case err := <-stopped:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("run did not stop")
	}

	assert.Len(t, api.callsOf("getMe"), 1)

	assert.Equal(t, wantMenu, publishedMenu(t, api))

	methods := api.methods()
	assert.Less(t, slices.Index(methods, "setMyCommands"), slices.Index(methods, "getUpdates"),
		"the command menu is published before polling starts")
	assert.Empty(t, api.callsOf("deleteMyCommands"), "the bot never deletes the command menu, also not at shutdown")

	menuSet := logs.entries("bot: command menu set")
	require.Len(t, menuSet, 1)
	assert.Equal(t, "info", menuSet[0].level)
	assert.Equal(t, len(wantMenu), menuSet[0].fields["commands"])
	assert.Empty(t, logs.entries("bot: failed to set command menu"))

	var welcome, maintenance []int64

	for _, c := range api.callsOf("sendMessage") {
		chatID := asInt64(c.params["chat_id"])

		switch c.params["text"] {
		case msgWelcomeBack:
			welcome = append(welcome, chatID)
		case msgMaintenance:
			maintenance = append(maintenance, chatID)
		}
	}

	assert.ElementsMatch(t, []int64{knownChat, blockedChat}, welcome, "both known chats get the welcome-back message")
	assert.Equal(t, []int64{knownChat}, maintenance, "the chat that blocked the bot is dropped after the first 403")

	reloaded := newUserStore(context.Background(), path)
	assert.Equal(t, []int64{knownChat}, reloaded.All())
}

// runUntilPolling runs the bot against api until it polls for updates, then stops it and requires a clean exit.
func runUntilPolling(ctx context.Context, t *testing.T, api *fakeTelegram) {
	t.Helper()

	ctx, cancel := context.WithCancel(ctx)

	stopped := make(chan error, 1)

	go func() { stopped <- run(ctx, testToken, telego.WithAPICaller(api), telego.WithDiscardLogger()) }()

	require.Eventually(t, func() bool {
		return len(api.callsOf("getUpdates")) > 0
	}, 5*time.Second, 10*time.Millisecond, "the bot polls for updates")

	cancel()

	select {
	case err := <-stopped:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("run did not stop")
	}
}

func TestRunSetsDefaultCommandList(t *testing.T) {
	t.Setenv("USER_STORE_PATH", filepath.Join(t.TempDir(), "users.json"))

	// Each start gets its own fake, so a getUpdates call of the first start cannot end the second one early.
	starts := []*fakeTelegram{{}, {}}

	for _, api := range starts {
		runUntilPolling(context.Background(), t, api)
	}

	assert.Equal(t, wantMenu, publishedMenu(t, starts[0]))
	assert.Equal(t, publishedMenu(t, starts[0]), publishedMenu(t, starts[1]), "every start publishes the same list")

	for _, api := range starts {
		params := api.callsOf("setMyCommands")[0].params
		assert.NotContains(t, params, "scope", "the default scope covers all chats")
		assert.NotContains(t, params, "language_code", "no language code covers users of every language")
	}
}

func TestRunKeepsRunningWhenCommandMenuFails(t *testing.T) {
	t.Setenv("USER_STORE_PATH", filepath.Join(t.TempDir(), "users.json"))

	ctx, logs := newLogContext()
	api := &fakeTelegram{methodErrors: map[string]int{"setMyCommands": http.StatusBadRequest}}

	// Polling starts with the handlers registered; TestBotAnswersEveryCommand covers their replies.
	runUntilPolling(ctx, t, api)

	failed := logs.entries("bot: failed to set command menu")
	require.Len(t, failed, 1)
	assert.Equal(t, "error", failed[0].level)
	assert.ErrorContains(t, failed[0].err, "setMyCommands")
	assert.Empty(t, logs.entries("bot: command menu set"))
	assert.Len(t, api.callsOf("setMyCommands"), 1, "no retry before the next start")
}

func TestRunFailsWithInvalidToken(t *testing.T) {
	err := run(context.Background(), "not-a-token", telego.WithDiscardLogger())

	assert.ErrorContains(t, err, "create bot")
}

func TestUserStorePath(t *testing.T) {
	t.Setenv("USER_STORE_PATH", "")
	assert.Equal(t, "users.json", userStorePath())

	t.Setenv("USER_STORE_PATH", "/data/chats.json")
	assert.Equal(t, "/data/chats.json", userStorePath())
}
