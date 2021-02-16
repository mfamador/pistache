// Package config defines the config loading for the app
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jinzhu/configor"
	"github.com/mfamador/pistache/internal/server"
	"github.com/mfamador/pistache/internal/services"
)

// LoggerConfig has the log level and if we should pretty print
type loggerConfig struct {
	Level  string `yaml:"level"`
	Pretty bool   `yaml:"pretty"`
}

// appConfig is main app config
type appConfig struct {
	Logger        loggerConfig    `yaml:"logger"`
	Server        server.Config   `yaml:"server"`
	DeploymentEnv string          `yaml:"deploymentEnv" env:"DEPLOYMENT_ENV" default:"unset"`
	Services      services.Config `yaml:"services"`
}

var (
	configFiles = []string{"config.yaml", "config.yml"}
	// Config contains all configuration values
	Config appConfig
	// Dir contains the location of the config directory
	Dir string
)

func searchConfig(dir string) (string, error) {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	dirPath := filepath.Join(absPath, "configs")
	if fileInfo, err := os.Stat(dirPath); err == nil && fileInfo.IsDir() {
		return dirPath, nil
	}

	if absPath == "/" {
		return "", errors.New("not found")
	}

	return searchConfig(filepath.Join(absPath, ".."))
}

//nolint:gochecknoinits
func init() {
	Dir = os.Getenv("CONFIGOR_DIR")
	if Dir == "" {
		var err error
		if Dir, err = searchConfig("."); err != nil {
			panic("Config dir not found")
		}
	}

	for i, v := range configFiles {
		configFiles[i] = filepath.Join(Dir, v)
	}

	if err := configor.New(&configor.Config{ENVPrefix: "-"}).Load(&Config, configFiles...); err != nil {
		panic("Invalid config")
	}

	hashPrefix := fmt.Sprintf("%s_%s", Config.DeploymentEnv, Config.Services.Cache.Hash.Prefix)
	Config.Services.Cache.Hash.Prefix = hashPrefix
}
