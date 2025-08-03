# NBG.gov.ge Client with Caching

This package provides a caching layer for the National Bank of Georgia (NBG) currency rates API client to reduce HTTP requests and improve performance.

## Features

- **Automatic Caching**: Caches currency rates by date and currency codes
- **Configurable TTL**: Set custom time-to-live for cache entries
- **Thread-Safe**: Safe for concurrent use
- **Cache Management**: Methods to clear cache and get statistics
- **Drop-in Replacement**: Implements the same `Client` interface

## Quick Start

### Basic Usage with Default Caching

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
    "github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

func main() {
    // Create a cached client with default 1-hour TTL
    client := nbggovge.NewCached()
    
    ctx := context.Background()
    date := time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)
    
    // First call hits the API
    rates1, err := client.Rates(ctx, 
        option.WithDate(date),
        option.WithCurrency("USD"),
        option.WithCurrency("EUR"),
    )
    if err != nil {
        panic(err)
    }
    
    // Second call uses cache (no HTTP request)
    rates2, err := client.Rates(ctx, 
        option.WithDate(date),
        option.WithCurrency("USD"),
        option.WithCurrency("EUR"),
    )
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("First call: %+v\n", rates1)
    fmt.Printf("Second call: %+v\n", rates2)
    
    // Check cache statistics
    stats := client.GetCacheStats()
    fmt.Printf("Cache entries: %d, Expired: %d\n", stats.TotalEntries, stats.ExpiredEntries)
}
```

### Custom TTL

```go
// Cache for 30 minutes
client := nbggovge.NewCachedWithTTL(time.Minute * 30)

// No expiration (cache until program restart)
client := nbggovge.NewCachedWithTTL(0)
```

### Custom HTTP Client

```go
import "net/http"

httpClient := &http.Client{
    Timeout: time.Second * 30,
}

// With default 1-hour TTL
client := nbggovge.NewCachedWithHTTPClient(httpClient)

// With custom TTL
client := nbggovge.NewCachedWithHTTPClientAndTTL(httpClient, time.Minute * 15)
```

### Direct Cache Management

```go
client := nbggovge.NewCached()

// Clear all cache entries
client.ClearCache()

// Remove only expired entries
client.ClearExpired()

// Get cache statistics
stats := client.GetCacheStats()
fmt.Printf("Total entries: %d\n", stats.TotalEntries)
fmt.Printf("Expired entries: %d\n", stats.ExpiredEntries)
```

## Cache Key Generation

The cache uses a combination of:
- Request date (YYYY-MM-DD format)
- Sorted currency codes (comma-separated)

This ensures that:
- Same date + currencies = cache hit
- Different dates = separate cache entries
- Currency order doesn't matter (`["USD", "EUR"]` same as `["EUR", "USD"]`)

## Thread Safety

The cached client is thread-safe and can be used concurrently from multiple goroutines. All cache operations are protected by read-write mutexes.

## Performance Considerations

- **Memory Usage**: Cache grows with unique date/currency combinations
- **TTL Strategy**: Choose TTL based on how fresh you need the data
  - Real-time trading: short TTL (minutes)
  - Daily reports: longer TTL (hours)
  - Historical analysis: no expiration (TTL = 0)
- **Cache Cleanup**: Call `ClearExpired()` periodically in long-running applications

## Migration from Regular Client

Replace your existing client instantiation:

```go
// Before
client := nbggovge.New()

// After - with caching
client := nbggovge.NewCached()
```

All other code remains the same as the cached client implements the same `Client` interface.

## Implementation Details

- Uses MD5 hashing for cache keys to ensure consistent key length
- Cache entries include timestamp for TTL calculations
- Read-write mutex ensures thread safety without blocking reads
- Expired entries can be manually cleared or will be overwritten when accessed
