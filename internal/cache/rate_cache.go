package cache

import (
	"sync"
	"time"
)

type RateCache struct {
	mu         sync.RWMutex
	rates      map[string]float32 // ключ: "USD_RUB"
	lastUpdate time.Time
	ttl        time.Duration
}

func NewRateCache(ttl time.Duration) *RateCache {
	return &RateCache{
		rates: make(map[string]float32),
		ttl:   ttl,
	}
}

// GetRate возвращает курс, если он есть и не устарел
func (c *RateCache) GetRate(from, to string) (float32, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if time.Since(c.lastUpdate) > c.ttl {
		return 0, false // устарело
	}

	key := from + "_" + to
	if rate, ok := c.rates[key]; ok {
		return rate, true
	}
	return 0, false
}

// SetAllRates сохраняет все курсы из ответа exchanger а
func (c *RateCache) SetAllRates(rates map[string]float32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rates = rates
	c.lastUpdate = time.Now()
}
