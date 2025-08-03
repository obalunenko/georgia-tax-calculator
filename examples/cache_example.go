package cacheexample

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

func main() {
	// Create a cached client with 30-minute TTL
	client := nbggovge.NewCachedWithTTL(time.Minute * 30)

	ctx := context.Background()
	date := time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)

	fmt.Println("=== Cache Example ===")

	// First call - hits the API
	fmt.Println("First call (API hit)...")
	start := time.Now()
	rates1, err := client.Rates(ctx,
		option.WithDate(date),
		option.WithCurrency("USD"),
		option.WithCurrency("EUR"),
	)
	duration1 := time.Since(start)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Duration: %v\n", duration1)
	fmt.Printf("USD Rate: %.4f\n", rates1.Currencies[1].Rate) // Assuming USD is second after EUR
	fmt.Printf("EUR Rate: %.4f\n", rates1.Currencies[0].Rate)

	// Check cache stats
	stats := client.GetCacheStats()
	fmt.Printf("Cache entries: %d\n", stats.TotalEntries)

	// Second call - uses cache
	fmt.Println("\nSecond call (cache hit)...")
	start = time.Now()
	rates2, err := client.Rates(ctx,
		option.WithDate(date),
		option.WithCurrency("USD"),
		option.WithCurrency("EUR"),
	)
	duration2 := time.Since(start)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Duration: %v\n", duration2)
	fmt.Printf("USD Rate: %.4f\n", rates2.Currencies[1].Rate)
	fmt.Printf("EUR Rate: %.4f\n", rates2.Currencies[0].Rate)

	fmt.Printf("\nPerformance improvement: %.2fx faster\n",
		float64(duration1)/float64(duration2))

	// Different date - new API call
	fmt.Println("\nDifferent date (new API hit)...")
	newDate := time.Date(2024, 2, 9, 0, 0, 0, 0, time.UTC)
	rates3, err := client.Rates(ctx,
		option.WithDate(newDate),
		option.WithCurrency("USD"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("USD Rate for %s: %.4f\n", newDate.Format("2006-01-02"), rates3.Currencies[0].Rate)

	// Final cache stats
	stats = client.GetCacheStats()
	fmt.Printf("\nFinal cache entries: %d\n", stats.TotalEntries)

	// Demonstrate cache clearing
	fmt.Println("\nClearing cache...")
	client.ClearCache()
	stats = client.GetCacheStats()
	fmt.Printf("Cache entries after clear: %d\n", stats.TotalEntries)
}
