// Package logging provides a small structured-logger factory so every service
// logs consistently (slog text handler, tagged with the service name).
package logging

import (
	"log/slog"
	"os"
)

// New returns a logger writing structured text to stderr, tagged with service.
func New(service string) *slog.Logger {
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(h).With("service", service)
}
