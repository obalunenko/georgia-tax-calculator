package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mymmrac/telego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunStartsAndShutsDown(t *testing.T) {
	const (
		knownChat   = int64(42)
		blockedChat = int64(43)
	)

	path := filepath.Join(t.TempDir(), "users.json")
	require.NoError(t, os.WriteFile(path, []byte("[42,43]"), 0o600))
	t.Setenv("USER_STORE_PATH", path)

	api := &fakeTelegram{sendErrors: map[int64]int{blockedChat: http.StatusForbidden}}

	ctx, cancel := context.WithCancel(context.Background())

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
