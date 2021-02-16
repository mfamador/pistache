// Package services for the services
package services

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/mfamador/pistache/internal/datasources/caches"
	"github.com/mfamador/pistache/internal/models"
	"github.com/mfamador/pistache/internal/repos"
	"github.com/rs/zerolog/log"
)

// CacheConfig contains the Cache service config options
type CacheConfig struct {
	Redis      *caches.RedisConfig `yaml:"redis"`
	Exceptions []string            `yaml:"exceptions"`
	Methods    []string            `yaml:"methods"`
	TTL        struct {
		Success int `yaml:"success"`
		Error   int `yaml:"error"`
	} `yaml:"ttl"`
	ForwardingHeaders []string `yaml:"forwardingHeaders"`
	Hash              struct {
		Prefix       string `yaml:"prefix" default:"app"`
		HashElements `yaml:",inline"`
		Overrides    []struct {
			OriginalPath string `yaml:"originalPath"`
			HashElements `yaml:",inline"`
		} `yaml:"overrides"`
	} `yaml:"hash"`
}

// HashElements details the elements used for computing a request hash key
type HashElements struct {
	UsePath     bool     `yaml:"usePath"`
	Headers     []string `yaml:"headers"`
	QueryParams []string `yaml:"queryParams"`
}

// Cache defines a cache service
type Cache interface {
	GetCachedResponse(*http.Request) (string, *models.Response, error)
	Store(string, *models.Response) bool
	Skip(*http.Request) bool
}

type cache struct {
	prefix            string
	inMemory          repos.Cache
	redis             repos.Cache
	hashElements      HashElements
	ttlSuccess        time.Duration
	ttlError          time.Duration
	exceptions        []string
	methods           []string
	overrides         map[string]HashElements
	forwardingHeaders []string
}

// NewCache creates a new Configs service
func NewCache(conf *CacheConfig) (Cache, error) {
	inMemory, err := caches.NewInMemory()
	if err != nil {
		return nil, err
	}

	c := &cache{
		prefix:            conf.Hash.Prefix,
		inMemory:          inMemory,
		hashElements:      conf.Hash.HashElements,
		ttlSuccess:        time.Duration(conf.TTL.Success) * time.Second,
		ttlError:          time.Duration(conf.TTL.Error) * time.Second,
		exceptions:        conf.Exceptions,
		methods:           conf.Methods,
		forwardingHeaders: conf.ForwardingHeaders,
	}

	for i, v := range c.hashElements.Headers {
		c.hashElements.Headers[i] = http.CanonicalHeaderKey(v)
	}

	c.overrides = make(map[string]HashElements)
	for _, v := range conf.Hash.Overrides {
		for i, h := range v.HashElements.Headers {
			v.HashElements.Headers[i] = http.CanonicalHeaderKey(h)
		}
		c.overrides[v.OriginalPath] = v.HashElements
	}

	log.Debug().
		Interface("conf", conf).
		Msg("NewCache")

	if conf.Redis != nil {
		c.redis = caches.NewRedis(conf.Redis)
	}

	return c, nil
}

func (c *cache) GetCachedResponse(r *http.Request) (string, *models.Response, error) {
	key, err := c.keyFromRequest(r)
	if err != nil {
		log.Warn().Err(err).Msg("Error creating key")
	}

	if err == nil {
		// We have the key to the cache, let's get it!
		cachedResponse, cerr := c.getFromCache(key)
		if cerr != nil {
			log.Warn().Err(cerr).Msg("Failed to get cached value")
		}
		return key, cachedResponse, cerr
	}

	return key, nil, err
}

func (c *cache) getFromCache(s string) (*models.Response, error) {
	resp, err := c.inMemory.Fetch(s)
	if resp != nil || err != nil {
		return resp, err
	}

	if c.redis != nil {
		redisResp, err := c.redis.Fetch(s)
		if err != nil {
			return nil, err
		}

		if redisResp != nil {
			// store locally
			c.inMemory.Store(s, redisResp, c.getTTL(redisResp.StatusCode))
			return redisResp, nil
		}
	}

	return nil, nil
}

// keyFromRequest returns the cache key for a request
// Key is the SHA-256 hash of the Method, Host, Path, Query fields and
// configured headers
// represented in lowercase hexadecimal string
func (c *cache) keyFromRequest(req *http.Request) (string, error) {
	return c.getKey(req, c.getHashElements(req))
}

func (c *cache) getURL(req *http.Request) (*url.URL, error) {
	for _, v := range c.forwardingHeaders {
		val := req.Header.Get(v)
		if val != "" {
			log.Debug().Str("val", val).Msg("Change to new URL")
			return url.Parse(fmt.Sprintf("http://%s%s", req.URL.Host, val))
		}
	}

	return req.URL, nil
}

func (c *cache) getKey(req *http.Request, elems HashElements) (string, error) {
	reqURL, err := c.getURL(req)
	if err != nil {
		return "", err
	}

	h := sha256.New()

	if _, err := h.Write([]byte(req.Method)); err != nil {
		return "", err
	}

	if _, err := h.Write([]byte(req.Host)); err != nil {
		return "", err
	}

	if elems.UsePath {
		if _, err := h.Write([]byte(reqURL.Path)); err != nil {
			return "", err
		}
	}

	queryMap := map[string][]string(reqURL.Query())
	if err := hashWriteMap(h, queryMap, elems.QueryParams); err != nil {
		return "", err
	}

	headerMap := map[string][]string(req.Header)
	if err := hashWriteMap(h, headerMap, elems.Headers); err != nil {
		return "", err
	}

	key := fmt.Sprintf("%x", h.Sum(nil))

	log.Debug().
		Str("key", key).
		Interface("elems", elems).
		Str("method", req.Method).
		Str("host", req.Host).
		Str("originalPath", req.URL.Path).
		Str("originalQuery", req.URL.RawQuery).
		Str("path", reqURL.Path).
		Str("query", reqURL.RawQuery).
		Interface("headers", req.Header).
		Msg("keyFromRequest")

	return fmt.Sprintf("{%s}-%s-", c.prefix, key), nil
}

func hashWriteMap(h hash.Hash, m map[string][]string, keys []string) error {
	if len(keys) == 0 {
		// If we want all the keys, we have to sort them first, so we get the same
		// hash all every time
		keys = make([]string, len(m))
		i := 0
		for k := range m {
			keys[i] = k
			i++
		}

		sort.Strings(keys)
	}

	for _, k := range keys {
		v, ok := m[k]
		if ok {
			val := fmt.Sprintf("%s=%s", k, v[0])
			log.Debug().Str("val", val).Msg("hashWriteMap")
			if _, err := h.Write([]byte(val)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *cache) getHashElements(req *http.Request) HashElements {
	el, ok := c.overrides[req.URL.Path]
	if ok {
		return el
	}

	return c.hashElements
}

// Store caches a response locally and in Redis
func (c *cache) Store(s string, response *models.Response) bool {
	ttl := c.getTTL(response.StatusCode)

	if c.redis != nil && !c.redis.Store(s, response, ttl) {
		return false
	}

	return c.inMemory.Store(s, response, ttl)
}

func (c *cache) getTTL(statusCode int) time.Duration {
	if statusCode >= http.StatusBadRequest {
		return c.ttlError
	}

	return c.ttlSuccess
}

// Skip determines if a request should skip the cache altogether
func (c *cache) Skip(req *http.Request) bool {
	pistached := contains(req.Method, c.methods) && !contains(req.RequestURI, c.exceptions)

	log.Debug().
		Str("request", req.RequestURI).
		Interface("methods", c.methods).
		Interface("exceptions", c.exceptions).
		Bool("methods", contains(req.Method, c.methods)).
		Bool("exceptions", !contains(req.RequestURI, c.exceptions)).
		Bool("pistached", pistached).
		Msg("Skip")

	return !pistached
}

func contains(str string, slice []string) bool {
	for _, p := range slice {
		if p == str {
			return true
		}
	}
	return false
}
