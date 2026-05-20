// Package logging — zerolog в JSON (по умолчанию) или консоль (LOG_FORMAT=text) для ELK.
package logging

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// NewZerolog возвращает логгер с полем service и уровнем из LOG_LEVEL.
func NewZerolog(defaultService string) zerolog.Logger {
	if s := strings.TrimSpace(os.Getenv("SERVICE_NAME")); s != "" {
		defaultService = s
	}
	zerolog.TimeFieldFormat = time.RFC3339Nano

	var out io.Writer = os.Stderr
	if strings.EqualFold(strings.TrimSpace(os.Getenv("LOG_FORMAT")), "text") {
		out = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	}

	zl := parseZerologLevel(os.Getenv("LOG_LEVEL"))
	return zerolog.New(out).
		Level(zl).
		With().
		Timestamp().
		Str("service", defaultService).
		Logger()
}

func parseZerologLevel(s string) zerolog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return zerolog.DebugLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}
