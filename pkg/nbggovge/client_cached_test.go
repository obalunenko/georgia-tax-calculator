package nbggovge

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCached(t *testing.T) {
	client := NewCached()

	assert.NotNil(t, client)
	assert.Equal(t, time.Hour, client.ttl)
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.cache)
}

func TestNewCachedWithTTL(t *testing.T) {
	ttl := time.Minute * 30
	client := NewCachedWithTTL(ttl)

	assert.NotNil(t, client)
	assert.Equal(t, ttl, client.ttl)
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.cache)
}

func TestNewCachedWithHTTPClient(t *testing.T) {
	httpClient := &http.Client{Timeout: time.Second * 10}
	client := NewCachedWithHTTPClient(httpClient)

	assert.NotNil(t, client)
	assert.Equal(t, time.Hour, client.ttl)
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.cache)
}

func TestNewCachedWithHTTPClientAndTTL(t *testing.T) {
	httpClient := &http.Client{Timeout: time.Second * 10}
	ttl := time.Minute * 15
	client := NewCachedWithHTTPClientAndTTL(httpClient, ttl)

	assert.NotNil(t, client)
	assert.Equal(t, ttl, client.ttl)
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.cache)
}
