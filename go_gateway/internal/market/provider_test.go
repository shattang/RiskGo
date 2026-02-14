package market

import (
	"context"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	provider := NewYahooFinanceProvider(500 * time.Millisecond)
	ctx := context.Background()

	// First call - should populate cache
	price1, err := provider.GetSpotPrice(ctx, "AAPL")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Second call - should come from cache (fast)
	price2, err := provider.GetSpotPrice(ctx, "AAPL")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if price1 != price2 {
		t.Errorf("Expected prices to match, got %f and %f", price1, price2)
	}

	// Wait for TTL to expire
	time.Sleep(600 * time.Millisecond)

	// Third call - should be "refreshed"
	_, err = provider.GetSpotPrice(ctx, "AAPL")
	if err != nil {
		t.Fatalf("Expected no error after expiry, got %v", err)
	}
}

func TestNotFound(t *testing.T) {
	provider := NewYahooFinanceProvider(1 * time.Minute)
	_, err := provider.GetSpotPrice(context.Background(), "INVALID")
	if err == nil {
		t.Error("Expected error for invalid ticker, got nil")
	}
}
