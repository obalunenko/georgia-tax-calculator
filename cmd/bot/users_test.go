package main

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

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
