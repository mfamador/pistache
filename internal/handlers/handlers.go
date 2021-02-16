// Package handlers defines all the request
// handlers for the API
package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Controller implements the logic to handle requests
type Controller struct{}

// Healthz GET /healthz controller function
func (ct *Controller) Healthz(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
