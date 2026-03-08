// Package logger provides a global zerolog-based logger with level control.
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

// Init initializes the global logger. Call once at startup.
// If debug is true, sets the log level to Debug; otherwise Info.
func Init(debug bool) {
	level := zerolog.InfoLevel
	if debug {
		level = zerolog.DebugLevel
	}

	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	log = zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()
}

// Get returns the global logger instance.
func Get() *zerolog.Logger {
	return &log
}
