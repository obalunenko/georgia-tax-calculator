package nbggovge

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

// mockClient is a simple mock implementation of the Client interface
type mockClient struct {
	callCount   int
	returnRates Rates
	returnError error
}

func (m *mockClient) Rates(ctx context.Context, opts ...option.RatesOption) (Rates, error) {
	m.callCount++
	return m.returnRates, m.returnError
}

func TestNewCachedClient(t *testing.T) {
	mockCli := &mockClient{}
	ttl := time.Hour

	cachedClient := NewCachedClient(mockCli, ttl)

	assert.NotNil(t, cachedClient)
	assert.Equal(t, mockCli, cachedClient.client)
	assert.Equal(t, ttl, cachedClient.ttl)
	assert.NotNil(t, cachedClient.cache)
}

func TestNewCachedClientWithDefaultTTL(t *testing.T) {
	mockCli := &mockClient{}

	cachedClient := NewCachedClientWithDefaultTTL(mockCli)

	assert.NotNil(t, cachedClient)
	assert.Equal(t, mockCli, cachedClient.client)
	assert.Equal(t, time.Hour, cachedClient.ttl)
	assert.NotNil(t, cachedClient.cache)
}

func TestCachedClient_Rates_CacheMiss(t *testing.T) {
	expectedRates := Rates{
		Date: "2023-12-01",
		Currencies: []Currency{
			{Code: "USD", Rate: 2.7, Name: "US Dollar"},
		},
	}

	mockCli := &mockClient{
		returnRates: expectedRates,
		returnError: nil,
	}
	cachedClient := NewCachedClient(mockCli, time.Hour)

	ctx := context.Background()
	opts := []option.RatesOption{option.WithDate(time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC))}

	// First call should hit the API
	rates, err := cachedClient.Rates(ctx, opts...)
	require.NoError(t, err)
	assert.Equal(t, expectedRates, rates)
	assert.Equal(t, 1, mockCli.callCount)
}

func TestCachedClient_Rates_CacheHit(t *testing.T) {
	expectedRates := Rates{
		Date: "2023-12-01",
		Currencies: []Currency{
			{Code: "USD", Rate: 2.7, Name: "US Dollar"},
		},
	}

	mockCli := &mockClient{
		returnRates: expectedRates,
		returnError: nil,
	}
	cachedClient := NewCachedClient(mockCli, time.Hour)

	ctx := context.Background()
	opts := []option.RatesOption{option.WithDate(time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC))}

	// First call should hit the API
	rates1, err := cachedClient.Rates(ctx, opts...)
	require.NoError(t, err)
	assert.Equal(t, expectedRates, rates1)
	assert.Equal(t, 1, mockCli.callCount)

	// Second call should use cache (mockCli.Rates should not be called again)
	rates2, err := cachedClient.Rates(ctx, opts...)
	require.NoError(t, err)
	assert.Equal(t, expectedRates, rates2)
	assert.Equal(t, 1, mockCli.callCount) // Still 1, no additional API call
}

func TestCachedClient_Rates_CacheExpired(t *testing.T) {
	expectedRates := Rates{
		Date: "2023-12-01",
		Currencies: []Currency{
			{Code: "USD", Rate: 2.7, Name: "US Dollar"},
		},
	}

	mockCli := &mockClient{
		returnRates: expectedRates,
		returnError: nil,
	}
	// Very short TTL for testing
	cachedClient := NewCachedClient(mockCli, time.Millisecond)

	ctx := context.Background()
	opts := []option.RatesOption{option.WithDate(time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC))}

	// First call should hit the API
	rates1, err := cachedClient.Rates(ctx, opts...)
	require.NoError(t, err)
	assert.Equal(t, expectedRates, rates1)
	assert.Equal(t, 1, mockCli.callCount)

	// Wait for cache to expire
	time.Sleep(time.Millisecond * 10)

	// Second call should hit the API again due to expiration
	rates2, err := cachedClient.Rates(ctx, opts...)
	require.NoError(t, err)
	assert.Equal(t, expectedRates, rates2)
	assert.Equal(t, 2, mockCli.callCount) // Should be 2 now
}

func TestCachedClient_GenerateCacheKey(t *testing.T) {
	mockCli := &mockClient{}
	cachedClient := NewCachedClient(mockCli, time.Hour)

	date := time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		currencyCodes  []string
		expectedLength int
	}{
		{
			name:           "empty currencies",
			currencyCodes:  []string{},
			expectedLength: 32, // MD5 hash length
		},
		{
			name:           "single currency",
			currencyCodes:  []string{"USD"},
			expectedLength: 32,
		},
		{
			name:           "multiple currencies",
			currencyCodes:  []string{"USD", "EUR", "GBP"},
			expectedLength: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := cachedClient.generateCacheKey(date, tt.currencyCodes)
			assert.Len(t, key, tt.expectedLength)
		})
	}

	// Test that same parameters produce same key
	key1 := cachedClient.generateCacheKey(date, []string{"USD", "EUR"})
	key2 := cachedClient.generateCacheKey(date, []string{"EUR", "USD"}) // Different order
	assert.Equal(t, key1, key2, "Cache keys should be same regardless of currency order")

	// Test that different dates produce different keys
	date2 := time.Date(2023, 12, 2, 0, 0, 0, 0, time.UTC)
	key3 := cachedClient.generateCacheKey(date2, []string{"USD"})
	key4 := cachedClient.generateCacheKey(date, []string{"USD"})
	assert.NotEqual(t, key3, key4, "Different dates should produce different cache keys")
}

func TestCachedClient_ClearCache(t *testing.T) {
	expectedRates := Rates{
		Date: "2023-12-01",
		Currencies: []Currency{
			{Code: "USD", Rate: 2.7, Name: "US Dollar"},
		},
	}

	mockCli := &mockClient{
		returnRates: expectedRates,
		returnError: nil,
	}
	cachedClient := NewCachedClient(mockCli, time.Hour)

	ctx := context.Background()
	opts := []option.RatesOption{option.WithDate(time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC))}

	// Populate cache
	_, err := cachedClient.Rates(ctx, opts...)
	require.NoError(t, err)
	assert.Equal(t, 1, mockCli.callCount)

	// Verify cache has entry
	stats := cachedClient.GetCacheStats()
	assert.Equal(t, 1, stats.TotalEntries)

	// Clear cache
	cachedClient.ClearCache()

	// Verify cache is empty
	stats = cachedClient.GetCacheStats()
	assert.Equal(t, 0, stats.TotalEntries)

	// Next call should hit API again
	_, err = cachedClient.Rates(ctx, opts...)
	require.NoError(t, err)
	assert.Equal(t, 2, mockCli.callCount) // Should be 2 now
}

func TestCachedClient_ClearExpired(t *testing.T) {
	expectedRates := Rates{
		Date: "2023-12-01",
		Currencies: []Currency{
			{Code: "USD", Rate: 2.7, Name: "US Dollar"},
		},
	}

	mockCli := &mockClient{
		returnRates: expectedRates,
		returnError: nil,
	}
	cachedClient := NewCachedClient(mockCli, time.Millisecond)

	ctx := context.Background()
	opts := []option.RatesOption{option.WithDate(time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC))}

	// Populate cache
	_, err := cachedClient.Rates(ctx, opts...)
	require.NoError(t, err)

	// Verify cache has entry
	stats := cachedClient.GetCacheStats()
	assert.Equal(t, 1, stats.TotalEntries)

	// Wait for expiration
	time.Sleep(time.Millisecond * 10)

	// Clear expired entries
	cachedClient.ClearExpired()

	// Verify cache is empty after clearing expired entries
	stats = cachedClient.GetCacheStats()
	assert.Equal(t, 0, stats.TotalEntries)
}

func TestCachedClient_GetCacheStats(t *testing.T) {
	mockCli := &mockClient{}
	cachedClient := NewCachedClient(mockCli, time.Millisecond)

	// Initially empty cache
	stats := cachedClient.GetCacheStats()
	assert.Equal(t, 0, stats.TotalEntries)
	assert.Equal(t, 0, stats.ExpiredEntries)

	expectedRates := Rates{
		Date: "2023-12-01",
		Currencies: []Currency{
			{Code: "USD", Rate: 2.7, Name: "US Dollar"},
		},
	}

	mockCli.returnRates = expectedRates
	mockCli.returnError = nil

	ctx := context.Background()
	opts := []option.RatesOption{option.WithDate(time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC))}

	// Populate cache
	_, err := cachedClient.Rates(ctx, opts...)
	require.NoError(t, err)

	// Check stats with fresh entry
	stats = cachedClient.GetCacheStats()
	assert.Equal(t, 1, stats.TotalEntries)
	assert.Equal(t, 0, stats.ExpiredEntries)

	// Wait for expiration
	time.Sleep(time.Millisecond * 10)

	// Check stats with expired entry
	stats = cachedClient.GetCacheStats()
	assert.Equal(t, 1, stats.TotalEntries)
	assert.Equal(t, 1, stats.ExpiredEntries)
}

func TestCachedClient_NoTTL(t *testing.T) {
	expectedRates := Rates{
		Date: "2023-12-01",
		Currencies: []Currency{
			{Code: "USD", Rate: 2.7, Name: "US Dollar"},
		},
	}

	mockCli := &mockClient{
		returnRates: expectedRates,
		returnError: nil,
	}
	// TTL = 0 means no expiration
	cachedClient := NewCachedClient(mockCli, 0)

	ctx := context.Background()
	opts := []option.RatesOption{option.WithDate(time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC))}

	// First call should hit the API
	rates1, err := cachedClient.Rates(ctx, opts...)
	require.NoError(t, err)
	assert.Equal(t, expectedRates, rates1)
	assert.Equal(t, 1, mockCli.callCount)

	// Wait some time
	time.Sleep(time.Millisecond * 10)

	// Second call should still use cache (no expiration)
	rates2, err := cachedClient.Rates(ctx, opts...)
	require.NoError(t, err)
	assert.Equal(t, expectedRates, rates2)
	assert.Equal(t, 1, mockCli.callCount) // Still 1

	// ClearExpired should not remove anything when TTL is 0
	cachedClient.ClearExpired()
	stats := cachedClient.GetCacheStats()
	assert.Equal(t, 1, stats.TotalEntries)
	assert.Equal(t, 0, stats.ExpiredEntries)
}
