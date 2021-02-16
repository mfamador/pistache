// Package handlers for the REST handlers
package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/mfamador/pistache/internal/services"
	"github.com/rs/zerolog/log"
)

const (
	defaultContentType = "application/octet-stream"
	pistacheHeader     = "X-Pistache"
	statusSkipped      = "skipped"
	statusCacheHit     = "hit"
	statusCacheMiss    = "miss"
)

// Cache exposes the interface
type Cache interface {
	Handle(echo.Context) error
}

type cacheHandler struct {
	cache services.Cache
	proxy services.Proxy
}

// NewCache creates a new Cache handler
func NewCache(cache services.Cache, proxy services.Proxy) Cache {
	return &cacheHandler{
		cache: cache,
		proxy: proxy,
	}
}

func (h *cacheHandler) Handle(ctx echo.Context) error {
	// Set the value so we can log this later with the request
	skip := h.cache.Skip(ctx.Request())
	ctx.Set("pistached", !skip)
	if skip {
		// Set the header, so our clients can know they've been pistached
		ctx.Response().Header().Set(pistacheHeader, statusSkipped)
		if _, err := h.proxy.Request(ctx); err != nil {
			log.Debug().Err(err).Msg("Failed to process request")
		}
		// The proxy service request will handle the HTTP flow for us
		// There is no need to return the error
		return nil
	}

	key, cachedResponse, err := h.cache.GetCachedResponse(ctx.Request())
	if err != nil {
		log.Debug().Err(err).Msg("Failed to cache request")
	}

	if cachedResponse != nil {
		// Set the cached headers
		for k, v := range cachedResponse.Header {
			ctx.Response().Header().Set(k, v[0])
		}

		// Set the header, so our clients can know they've been pistached
		ctx.Response().Header().Set(pistacheHeader, statusCacheHit)

		// return cached response
		return ctx.Blob(
			cachedResponse.StatusCode,
			extractContentType(cachedResponse.Header),
			cachedResponse.Body,
		)
	}

	// Set the header, so our clients can know they've been pistached
	ctx.Response().Header().Set(pistacheHeader, statusCacheMiss)
	response, err := h.proxy.Request(ctx)

	if err == nil && key != "" {
		go h.cache.Store(key, response)
	}

	return nil
}

func extractContentType(headers map[string][]string) string {
	if cts, ok := headers["Content-Type"]; ok {
		return cts[0]
	}

	return defaultContentType
}
