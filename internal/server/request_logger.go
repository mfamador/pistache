// Package server defines the app server boot
package server

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// RequestLogger is an echo middleware to log HTTP requests
func RequestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var l *zerolog.Event
		req := c.Request()
		res := c.Response()

		start := time.Now()
		if err = next(c); err != nil {
			c.Error(err)
			l = log.Warn().Err(err)
		} else {
			l = log.Debug()
		}
		stop := time.Now()

		bytesIn, err := strconv.Atoi(req.Header.Get(echo.HeaderContentLength))
		if err != nil {
			bytesIn = 0
		}
		pistached := false
		if v := c.Get("pistached"); v != nil {
			pistached = v.(bool)
		}
		l.Str("remote_ip", c.RealIP()).
			Str("host", req.Host).
			Str("method", req.Method).
			Str("uri", req.RequestURI).
			Str("user_agent", req.UserAgent()).
			Int("status", res.Status).
			Float64("latency", stop.Sub(start).Seconds()).
			Int("bytes_in", bytesIn).
			Int64("bytes_out", res.Size).
			Bool("pistached", pistached).
			Msg("request")

		return nil
	}
}
