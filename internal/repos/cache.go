// Package repos defines the entity interface to interact with a data source
package repos

import (
	"time"

	"github.com/mfamador/pistache/internal/models"
)

// Cache exposes the interface to access a data source
type Cache interface {
	Fetch(string) (*models.Response, error)
	Store(string, *models.Response, time.Duration) bool
}
