// Package logger defines the logger bootstrap
package logger

import (
	"os"

	"github.com/mfamador/pistache/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//nolint:gochecknoinits
func init() {
	SetPretty(config.Config.Logger.Pretty)

	switch config.Config.Logger.Level {
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// SetPretty switches pretty logs on and off
func SetPretty(b bool) {
	if b {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
		zerolog.MessageFieldName = "msg"
	}
}
