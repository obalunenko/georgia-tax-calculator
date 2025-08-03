package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/option"
)

func main() {
	// Create cached client with 1-hour TTL
	client := nbggovge.NewCached()

	ctx := context.Background()

	// Test with today's date
	today := time.Now()
	fmt.Printf("Fetching rates for %s...\n", today.Format("2006-01-02"))

	// First call - hits API
	fmt.Println("🌐 First call (API request)")
	start := time.Now()
	rates, err := client.Rates(ctx,
		option.WithDate(today),
		option.WithCurrency("USD"),
		option.WithCurrency("EUR"),
	)
	apiDuration := time.Since(start)
	if err != nil {
		log.Fatal("Error fetching rates:", err)
	}

	fmt.Printf("✅ Got %d currencies in %v\n", len(rates.Currencies), apiDuration)

	// Second call - uses cache
	fmt.Println("⚡ Second call (cached)")
	start = time.Now()
	rates2, err := client.Rates(ctx,
		option.WithDate(today),
		option.WithCurrency("USD"),
		option.WithCurrency("EUR"),
	)
	cacheDuration := time.Since(start)
	if err != nil {
		log.Fatal("Error fetching cached rates:", err)
	}

	fmt.Printf("✅ Got %d currencies in %v\n", len(rates2.Currencies), cacheDuration)

	// Show performance improvement
	speedup := float64(apiDuration) / float64(cacheDuration)
	fmt.Printf("🚀 Cache is %.0fx faster!\n", speedup)

	// Display cache stats
	stats := client.GetCacheStats()
	fmt.Printf("📊 Cache stats: %d entries, %d expired\n", stats.TotalEntries, stats.ExpiredEntries)

	// Show some rates
	if len(rates.Currencies) > 0 {
		fmt.Println("\n💱 Sample rates:")
		for _, currency := range rates.Currencies {
			fmt.Printf("   %s: %s GEL\n", currency.Code, currency.RateFormated)
		}
	}
}
