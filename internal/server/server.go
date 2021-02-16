// Package server defines the app server boot
package server

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mfamador/pistache/internal/services"

	echoPrometheus "github.com/globocom/echo-prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	"github.com/mfamador/pistache/internal/handlers"
)

// Config defines the handler configuration
type Config struct {
	Port int `yaml:"port"`
}

type httpErrorMessage struct {
	Message interface{} `json:"message"`
	Errors  []string    `json:"errors,omitempty"`
}

func customHTTPErrorHandler(err error, c echo.Context) {
	var message interface{}
	var errors []string

	var code int
	switch t := err.(type) {
	case *echo.HTTPError:
		code = t.Code
		message = t.Message
	default:
		code = http.StatusInternalServerError
		message = err.Error()
	}

	res := httpErrorMessage{Message: message, Errors: errors}

	// Send response
	if !c.Response().Committed {
		var writeErr error

		log.Error().Err(err).Interface("path", c.Path()).Msg(err.Error())

		if c.Request().Method == http.MethodHead {
			writeErr = c.NoContent(code)
		} else {
			writeErr = c.JSON(code, res)
		}

		if writeErr != nil {
			log.Error().Err(writeErr).Msg("Failed to reply with the HTTP error")
		}
	}
}

// Setup abstracts booting the echo framework
func Setup() (*echo.Echo, error) {
	e := echo.New()

	// Hide echo banner and port, so we only output valid logs
	e.HideBanner = true
	e.HidePort = true

	e.HTTPErrorHandler = customHTTPErrorHandler

	e.Use(RequestLogger)
	e.Use(middleware.Recover())

	return e, nil
}

// Start starts the echo http server
func Start(serverConfig Config, servicesConfig *services.Config) error {
	e, err := Setup()
	if err != nil {
		return err
	}

	c := handlers.Controller{}

	// Define routes and middleware
	e.Use(echoPrometheus.MetricsMiddleware())

	pistache := e.Group("/pistache")
	pistache.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	pistache.GET("/healthz", c.Healthz)

	pService, err := services.NewProxy(servicesConfig.Proxy)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create proxy")
		return err
	}

	cService, err := services.NewCache(&servicesConfig.Cache)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create cache")
		return err
	}

	cHandler := handlers.NewCache(cService, pService)

	// Use our handler in case we hit the 'Skipper' target in ProxyMiddleware
	e.Any("/*", cHandler.Handle)

	log.Info().Int("Starting Pistache on port", serverConfig.Port)
	return e.Start(fmt.Sprintf(":%d", serverConfig.Port))
}
