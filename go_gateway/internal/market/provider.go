package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// MarketDataProvider defines the interface for fetching market data.
type MarketDataProvider interface {
	GetSpotPrice(ctx context.Context, ticker string) (float64, error)
	GetRiskFreeRate(ctx context.Context) (float64, error)
	GetVolatility(ctx context.Context, ticker string) (float64, error)
}

type cacheItem struct {
	value     float64
	expiresAt time.Time
}

// YahooFinanceProvider implements MarketDataProvider with a simple in-memory cache.
type YahooFinanceProvider struct {
	cache map[string]cacheItem
	mu    sync.RWMutex
	ttl   time.Duration
	client *http.Client
}

func NewYahooFinanceProvider(ttl time.Duration) *YahooFinanceProvider {
	return &YahooFinanceProvider{
		cache: make(map[string]cacheItem),
		ttl:   ttl,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (y *YahooFinanceProvider) getFromCache(key string) (float64, bool) {
	y.mu.RLock()
	defer y.mu.RUnlock()
	item, ok := y.cache[key]
	if !ok || time.Now().After(item.expiresAt) {
		return 0, false
	}
	return item.value, true
}

func (y *YahooFinanceProvider) setToCache(key string, value float64) {
	y.mu.Lock()
	defer y.mu.Unlock()
	y.cache[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(y.ttl),
	}
}

type yahooResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				RegularMarketPrice float64 `json:"regularMarketPrice"`
			} `json:"meta"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"chart"`
}

func (y *YahooFinanceProvider) GetSpotPrice(ctx context.Context, ticker string) (float64, error) {
	cacheKey := fmt.Sprintf("spot:%s", ticker)
	if val, ok := y.getFromCache(cacheKey); ok {
		return val, nil
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=1d", ticker)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	// Use a realistic User-Agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := y.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("yahoo finance returned status: %s", resp.Status)
	}

	var data yahooResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	if len(data.Chart.Result) == 0 {
		return 0, fmt.Errorf("no data found for ticker %s", ticker)
	}

	price := data.Chart.Result[0].Meta.RegularMarketPrice
	y.setToCache(cacheKey, price)
	return price, nil
}

func (y *YahooFinanceProvider) GetRiskFreeRate(ctx context.Context) (float64, error) {
	cacheKey := "risk_free_rate"
	if val, ok := y.getFromCache(cacheKey); ok {
		return val, nil
	}
	rate := 0.0425 
	y.setToCache(cacheKey, rate)
	return rate, nil
}

func (y *YahooFinanceProvider) GetVolatility(ctx context.Context, ticker string) (float64, error) {
	cacheKey := fmt.Sprintf("vol:%s", ticker)
	if val, ok := y.getFromCache(cacheKey); ok {
		return val, nil
	}

	vol := 0.30
	y.setToCache(cacheKey, vol)
	return vol, nil
}
