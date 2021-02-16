// Package services has proxy services
package services

import (
	"fmt"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mfamador/pistache/internal/models"
)

// ProxyConfig contains the Proxy config options
type ProxyConfig struct {
	Upstreams []Upstream `yaml:"upstreams"`
}

// Upstream defines an upstream target
type Upstream struct {
	Host string
	Port int
}

// URL returns a url.URL object for this upstream
func (u Upstream) URL() (*url.URL, error) {
	return url.Parse(fmt.Sprintf("http://%s:%d", u.Host, u.Port))
}

// Proxy defines a proxy service
type Proxy interface {
	Request(c echo.Context) (*models.Response, error)
}

type proxy struct {
	proxyHandler echo.HandlerFunc
}

// NewProxy creates a new Configs service
func NewProxy(conf ProxyConfig) (Proxy, error) {
	targets := make([]*middleware.ProxyTarget, len(conf.Upstreams))
	for i, upstream := range conf.Upstreams {
		u, err := upstream.URL()
		if err != nil {
			return nil, err
		}

		targets[i] = &middleware.ProxyTarget{URL: u}
	}

	balancer := middleware.NewRoundRobinBalancer(targets)

	return &proxy{
		proxyHandler: middleware.Proxy(balancer)(noop),
	}, nil
}

// Request proxies an HTTP request
func (h proxy) Request(c echo.Context) (*models.Response, error) {
	// If we don't have the pistached value, something is wrong, so just forward
	// the request
	// If pistached is false, then we don't want to cache it, so just forward the
	// request
	if c.Get("pistached") == nil || !c.Get("pistached").(bool) {
		return nil, h.proxyHandler(c)
	}

	rp := &models.Response{}
	err := middleware.BodyDump(ResponseStorer(rp))(h.proxyHandler)(c)

	return rp, err
}

// ResponseStorer stores response information in a `models.Response`
func ResponseStorer(rp *models.Response) func(echo.Context, []byte, []byte) {
	return func(c echo.Context, reqBody, resBody []byte) {
		rp.StatusCode = c.Response().Status
		rp.Header = c.Response().Writer.Header()
		rp.Body = resBody
	}
}

func noop(c echo.Context) error { return nil }
