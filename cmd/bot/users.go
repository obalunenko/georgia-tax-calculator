package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
	log "github.com/obalunenko/logger"
)

const (
	msgMaintenance = "🔧 Bot is going offline for maintenance. We'll be back soon!"
	msgWelcomeBack = "👋 Bot is back online! Use /start to continue."
)

// userStoreFileMode keeps the chat IDs readable only by the bot user.
const userStoreFileMode os.FileMode = 0o600

// userStore persists known user chat IDs across restarts.
type userStore struct {
	mu      sync.Mutex
	path    string
	chatIDs map[int64]struct{}
}

// newUserStore loads the store from path. A missing file starts an empty store;
// other load errors are logged and the store starts empty too.
func newUserStore(ctx context.Context, path string) *userStore {
	s := &userStore{
		path:    path,
		chatIDs: make(map[int64]struct{}),
	}

	if err := s.load(); err != nil {
		log.WithError(ctx, err).WithField("path", path).Warn("bot: failed to load user store")
	}

	return s
}

// load reads chat IDs from the JSON file. A missing file is not an error.
func (s *userStore) load() error {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	var ids []int64
	if err = json.Unmarshal(data, &ids); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	for _, id := range ids {
		s.chatIDs[id] = struct{}{}
	}

	return nil
}

// save writes chat IDs to the JSON file. Must be called with lock held.
func (s *userStore) save() error {
	ids := make([]int64, 0, len(s.chatIDs))
	for id := range s.chatIDs {
		ids = append(ids, id)
	}

	data, err := json.Marshal(ids)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	if err = os.WriteFile(s.path, data, userStoreFileMode); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

// Track records a chat ID and persists the store when a new ID is seen.
// It reports whether the ID is new; a save error keeps the ID in memory.
func (s *userStore) Track(chatID int64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.chatIDs[chatID]; ok {
		return false, nil
	}

	s.chatIDs[chatID] = struct{}{}

	if err := s.save(); err != nil {
		return true, fmt.Errorf("save user store %s: %w", s.path, err)
	}

	return true, nil
}

// Remove forgets a chat ID and persists the store. It reports whether the ID was known.
func (s *userStore) Remove(chatID int64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.chatIDs[chatID]; !ok {
		return false, nil
	}

	delete(s.chatIDs, chatID)

	if err := s.save(); err != nil {
		return true, fmt.Errorf("save user store %s: %w", s.path, err)
	}

	return true, nil
}

// All returns a snapshot of all known chat IDs.
func (s *userStore) All() []int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	ids := make([]int64, 0, len(s.chatIDs))
	for id := range s.chatIDs {
		ids = append(ids, id)
	}

	return ids
}

// broadcastInterval keeps broadcasts below Telegram's limit of about 30 messages per second.
const broadcastInterval = 50 * time.Millisecond

// sendFunc sends text to a chat.
type sendFunc func(ctx context.Context, chatID int64, text string) error

// botSender sends messages through bot.
func botSender(bot *telego.Bot) sendFunc {
	return func(ctx context.Context, chatID int64, text string) error {
		_, err := sendMessage(ctx, bot, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: chatID},
			Text:   text,
		})

		return err
	}
}

// broadcast sends text to every known chat, waiting interval between messages, until ctx is done.
// Chats that blocked the bot are removed from the store; other per-chat errors are logged and do not stop it.
func broadcast(ctx context.Context, send sendFunc, users *userStore, text string, interval time.Duration) {
	for i, chatID := range users.All() {
		if i > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
			}
		}

		err := send(ctx, chatID, text)
		if err == nil {
			continue
		}

		ctxLog := log.WithError(ctx, err).WithField("chat_id", chatID)

		if !isChatUnreachable(err) {
			ctxLog.Warn("broadcast: failed to send message")

			continue
		}

		if _, rmErr := users.Remove(chatID); rmErr != nil {
			ctxLog.WithField("remove_error", rmErr.Error()).Error("broadcast: failed to remove unreachable chat")

			continue
		}

		ctxLog.Info("broadcast: removed chat that blocked the bot")
	}
}

// isChatUnreachable reports a Telegram 403 error: the user blocked the bot or the bot was removed from the chat.
func isChatUnreachable(err error) bool {
	var apiErr *telegoapi.Error

	return errors.As(err, &apiErr) && apiErr.ErrorCode == http.StatusForbidden
}
