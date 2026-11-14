package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"

	"github.com/mymmrac/telego"
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

// broadcast sends text to every chatID. Per-user errors are logged but do not abort the loop.
func broadcast(ctx context.Context, bot *telego.Bot, chatIDs []int64, text string) {
	for _, chatID := range chatIDs {
		_, err := sendMessage(ctx, bot, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: chatID},
			Text:   text,
		})
		if err != nil {
			log.WithError(ctx, err).
				WithField("chat_id", chatID).
				Warn("broadcast: failed to send message")
		}
	}
}
