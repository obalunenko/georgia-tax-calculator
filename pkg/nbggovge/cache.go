package nbggovge

import (
	"context"
	"fmt"
	"hash/crc32"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/internal"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

// CachedClient wraps nbggovge client with caching functionality.
type CachedClient struct {
	client Client
	cache  map[string]cachedRates
	mutex  sync.RWMutex
	ttl    time.Duration
}

// cachedRates holds cached rates data with timestamp.
type cachedRates struct {
	rates     Rates
	timestamp time.Time
}

// NewCachedClient creates a new cached client with specified TTL.
// If ttl is 0, cache entries will not expire (cache forever until program restart).
func NewCachedClient(client Client, ttl time.Duration) *CachedClient {
	return &CachedClient{
		client: client,
		cache:  make(map[string]cachedRates),
		ttl:    ttl,
	}
}

// NewCachedClientWithDefaultTTL creates a new cached client with default TTL of 1 hour.
func NewCachedClientWithDefaultTTL(client Client) *CachedClient {
	return NewCachedClient(client, time.Hour)
}

// Rates returns rates with caching. If rates for the same date and currencies
// are already cached and not expired, returns cached data. Otherwise, fetches
// new data from the API and caches it.
func (c *CachedClient) Rates(ctx context.Context, opts ...option.RatesOption) (Rates, error) {
	// Parse options to get parameters
	var params internal.RatesParams
	for _, opt := range opts {
		opt.Apply(&params)
	}

	if params.Date.IsZero() {
		params.Date = time.Now()
	}

	// Generate cache key
	cacheKey := c.generateCacheKey(params.Date, params.CurrencyCodes)

	// Try to get from cache first
	c.mutex.RLock()
	if cached, exists := c.cache[cacheKey]; exists {
		// Check if cache entry is still valid
		if c.ttl == 0 || time.Since(cached.timestamp) < c.ttl {
			c.mutex.RUnlock()
			return cached.rates, nil
		}
	}
	c.mutex.RUnlock()

	// Cache miss or expired, fetch from API
	rates, err := c.client.Rates(ctx, opts...)
	if err != nil {
		return Rates{}, err
	}

	// Store in cache
	c.mutex.Lock()
	c.cache[cacheKey] = cachedRates{
		rates:     rates,
		timestamp: time.Now(),
	}
	c.mutex.Unlock()

	return rates, nil
}

// generateCacheKey creates a unique cache key based on date and currency codes.
func (c *CachedClient) generateCacheKey(date time.Time, currencyCodes []string) string {
	// Format date as YYYY-MM-DD
	dateStr := date.Format("2006-01-02")

	// Sort currency codes to ensure consistent cache keys
	sortedCodes := make([]string, len(currencyCodes))
	copy(sortedCodes, currencyCodes)
	sort.Strings(sortedCodes)

	// Create key: date + sorted currencies
	keyData := dateStr + ":" + strings.Join(sortedCodes, ",")

	// Use CRC32 hash for shorter, consistent keys
	hash := crc32.ChecksumIEEE([]byte(keyData))
	return fmt.Sprintf("%x", hash)
}

// ClearCache removes all cached entries.
func (c *CachedClient) ClearCache() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = make(map[string]cachedRates)
}

// ClearExpired removes expired cache entries.
func (c *CachedClient) ClearExpired() {
	if c.ttl == 0 {
		return // No expiration when TTL is 0
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key, cached := range c.cache {
		if time.Since(cached.timestamp) >= c.ttl {
			delete(c.cache, key)
		}
	}
}

// CacheStats returns statistics about the cache.
type CacheStats struct {
	TotalEntries   int
	ExpiredEntries int
}

// GetCacheStats returns current cache statistics.
func (c *CachedClient) GetCacheStats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := CacheStats{
		TotalEntries: len(c.cache),
	}

	if c.ttl > 0 {
		for _, cached := range c.cache {
			if time.Since(cached.timestamp) >= c.ttl {
				stats.ExpiredEntries++
			}
		}
	}

	return stats
}
