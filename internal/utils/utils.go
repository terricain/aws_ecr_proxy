package utils

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func LogNameToLevel(logName string) zerolog.Level {
	logLevel := zerolog.InfoLevel
	switch logName {
	case "INFO":
		logLevel = zerolog.InfoLevel
	case "DEBUG":
		logLevel = zerolog.DebugLevel
	case "WARN":
		logLevel = zerolog.WarnLevel
	case "ERROR":
		logLevel = zerolog.ErrorLevel
	default:
		log.Warn().Msgf("Unknown log level \"%s\", defaulting to INFO", logName)
	}
	return logLevel
}
