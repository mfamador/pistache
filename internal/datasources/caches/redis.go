// Package caches for the caches implementation
package caches

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-redis/redis/v8"
	"github.com/mfamador/pistache/internal/models"
	"github.com/mfamador/pistache/internal/repos"
)

// RedisConfig contains the config options for a REDIS cluster server
type RedisConfig struct {
	Servers []struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"servers"`
}

type redisc struct {
	cache *redis.ClusterClient
}

func (i redisc) Fetch(s string) (*models.Response, error) {
	value, err := i.cache.Get(context.Background(), s).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	resp := models.Response{}
	if err := json.Unmarshal([]byte(value), &resp); err != nil {
		log.Warn().Err(err).Msg("Failed to get unmarshal response")
		return nil, err
	}

	return &resp, nil
}

func (i redisc) Store(s string, response *models.Response, ttl time.Duration) bool {
	res := i.cache.Set(context.Background(), s, response, ttl)
	if res.Err() != nil {
		log.Warn().
			Interface("response", response).
			Err(res.Err()).Msg("Failed to store response in REDIS")
		return false
	}

	return true
}

// NewRedis handles cache supported in Redis
func NewRedis(conf *RedisConfig) repos.Cache {
	redisURLs := make([]string, len(conf.Servers))
	for i, server := range conf.Servers {
		redisURLs[i] = fmt.Sprintf("%s:%d", server.Host, server.Port)
	}

	cache := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: redisURLs,
	})

	return &redisc{
		cache: cache,
	}
}
