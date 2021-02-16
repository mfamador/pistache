// Package services defines the base service
package services

// Config contains the configuration options for all the services
type Config struct {
	Cache CacheConfig `yaml:"cache"`
	Proxy ProxyConfig `yaml:"proxy"`
}
