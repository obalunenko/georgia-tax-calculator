package main

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/mymmrac/telego"
	log "github.com/obalunenko/logger"
)

const (
	msgMaintenance = "🔧 Bot is going offline for maintenance. We'll be back soon!"
	msgWelcomeBack = "👋 Bot is back online! Use /start to continue."
)

// userStore persists known user chat IDs across restarts.
type userStore struct {
	mu      sync.Mutex
	path    string
	chatIDs map[int64]struct{}
}

func newUserStore(path string) *userStore {
	s := &userStore{
		path:    path,
		chatIDs: make(map[int64]struct{}),
	}

	s.load()

	return s
}

// load reads chat IDs from the JSON file. Missing file is silently ignored.
func (s *userStore) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}

	var ids []int64
	if err = json.Unmarshal(data, &ids); err != nil {
		return
	}

	for _, id := range ids {
		s.chatIDs[id] = struct{}{}
	}
}

// save writes chat IDs to the JSON file. Must be called with lock held.
func (s *userStore) save() {
	ids := make([]int64, 0, len(s.chatIDs))
	for id := range s.chatIDs {
		ids = append(ids, id)
	}

	data, err := json.Marshal(ids)
	if err != nil {
		return
	}

	_ = os.WriteFile(s.path, data, 0o600)
}

// Track records a chat ID and persists the store when a new ID is seen.
func (s *userStore) Track(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.chatIDs[chatID]; ok {
		return
	}

	s.chatIDs[chatID] = struct{}{}
	s.save()
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
		_, err := bot.SendMessage(ctx, &telego.SendMessageParams{
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
