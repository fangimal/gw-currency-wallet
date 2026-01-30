package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateCache_SetAndGet(t *testing.T) {
	cache := NewRateCache(1 * time.Second)
	cache.SetAllRates(map[string]float32{"USD_RUB": 90.5})

	rate, ok := cache.GetRate("USD", "RUB")
	assert.True(t, ok)
	assert.Equal(t, float32(90.5), rate)
}

func TestRateCache_Expired(t *testing.T) {
	cache := NewRateCache(100 * time.Millisecond)
	cache.SetAllRates(map[string]float32{"USD_RUB": 90.5})

	time.Sleep(150 * time.Millisecond)

	_, ok := cache.GetRate("USD", "RUB")
	assert.False(t, ok)
}
