package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/mymmrac/telego/telegoapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "users.json")

	store := newUserStore(context.Background(), path)
	assert.Empty(t, store.All())

	for _, id := range []int64{10, 20} {
		isNew, err := store.Track(id)
		require.NoError(t, err)
		assert.True(t, isNew)
	}

	isNew, err := store.Track(10)
	require.NoError(t, err)
	assert.False(t, isNew, "a known chat is not new")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	reloaded := newUserStore(context.Background(), path)

	got := reloaded.All()
	slices.Sort(got)
	assert.Equal(t, []int64{10, 20}, got)
}

func TestUserStoreLoad(t *testing.T) {
	tests := []struct {
		name    string
		content *string
		want    []int64
		wantErr bool
	}{
		{name: "missing file", content: nil, want: nil, wantErr: false},
		{name: "valid file", content: new("[1,2]"), want: []int64{1, 2}, wantErr: false},
		{name: "invalid JSON", content: new("{"), want: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "users.json")

			if tt.content != nil {
				require.NoError(t, os.WriteFile(path, []byte(*tt.content), 0o600))
			}

			store := &userStore{path: path, chatIDs: make(map[int64]struct{})}

			err := store.load()
			assert.Equal(t, tt.wantErr, err != nil, err)

			got := store.All()
			slices.Sort(got)
			assert.Equal(t, len(tt.want), len(got))

			for _, id := range tt.want {
				assert.Contains(t, got, id)
			}
		})
	}
}

func TestUserStoreTrackReportsSaveError(t *testing.T) {
	// The parent directory does not exist, so the file cannot be written.
	path := filepath.Join(t.TempDir(), "missing-dir", "users.json")

	store := newUserStore(context.Background(), path)

	isNew, err := store.Track(42)
	require.Error(t, err)
	assert.True(t, isNew)
	assert.Equal(t, []int64{42}, store.All(), "the chat stays known in memory")
}

func newTestStore(t *testing.T, ids ...int64) *userStore {
	t.Helper()

	store := newUserStore(context.Background(), filepath.Join(t.TempDir(), "users.json"))

	for _, id := range ids {
		_, err := store.Track(id)
		require.NoError(t, err)
	}

	return store
}

func TestUserStoreRemove(t *testing.T) {
	store := newTestStore(t, 1, 2)

	removed, err := store.Remove(1)
	require.NoError(t, err)
	assert.True(t, removed)

	removed, err = store.Remove(1)
	require.NoError(t, err)
	assert.False(t, removed, "already removed")

	reloaded := newUserStore(context.Background(), store.path)
	assert.Equal(t, []int64{2}, reloaded.All(), "removal is persisted")
}

func TestIsChatUnreachable(t *testing.T) {
	blocked := &telegoapi.Error{ErrorCode: http.StatusForbidden, Description: "Forbidden: bot was blocked by the user"}

	assert.True(t, isChatUnreachable(fmt.Errorf("telego: sendMessage: api: %w", blocked)))
	assert.False(t, isChatUnreachable(&telegoapi.Error{ErrorCode: http.StatusTooManyRequests}))
	assert.False(t, isChatUnreachable(errors.New("network down")))
}

func TestBroadcast(t *testing.T) {
	const (
		ok      int64 = 1
		blocked int64 = 2
		failing int64 = 3
	)

	store := newTestStore(t, ok, blocked, failing)

	var (
		mu   sync.Mutex
		sent []int64
	)

	send := func(_ context.Context, chatID int64, text string) error {
		assert.Equal(t, msgMaintenance, text)

		mu.Lock()

		sent = append(sent, chatID)

		mu.Unlock()

		switch chatID {
		case blocked:
			return fmt.Errorf("api: %w", &telegoapi.Error{ErrorCode: http.StatusForbidden})
		case failing:
			return errors.New("network down")
		default:
			return nil
		}
	}

	start := time.Now()

	broadcast(context.Background(), send, store, msgMaintenance, 5*time.Millisecond)

	assert.GreaterOrEqual(t, time.Since(start), 10*time.Millisecond, "two pauses between three messages")
	assert.ElementsMatch(t, []int64{ok, blocked, failing}, sent)

	got := store.All()
	slices.Sort(got)
	assert.Equal(t, []int64{ok, failing}, got, "only the chat that blocked the bot is removed")
}

func TestBroadcastStopsWhenContextIsDone(t *testing.T) {
	store := newTestStore(t, 1, 2, 3)

	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	send := func(context.Context, int64, string) error {
		calls++

		cancel()

		return nil
	}

	broadcast(ctx, send, store, msgWelcomeBack, time.Hour)

	assert.Equal(t, 1, calls)
}
