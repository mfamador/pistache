// Package caches for the caches implementation
package caches

import (
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/mfamador/pistache/internal/models"
	"github.com/mfamador/pistache/internal/repos"
)

type inMemory struct {
	cache *ristretto.Cache
}

func (i inMemory) Fetch(s string) (*models.Response, error) {
	value, found := i.cache.Get(s)
	if !found {
		return nil, nil
	}

	resp := value.(*models.Response)

	return resp, nil
}

func (i inMemory) Store(s string, response *models.Response, ttl time.Duration) bool {
	return i.cache.SetWithTTL(s, response, 1, ttl)
}

// NewInMemory handles in memory cache
func NewInMemory() (repos.Cache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})

	if err != nil {
		return nil, err
	}

	return &inMemory{
		cache: cache,
	}, nil
}
