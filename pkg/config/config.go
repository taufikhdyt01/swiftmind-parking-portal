// Package config provides small helpers for reading service configuration
// from environment variables, with sensible local-development defaults.
package config

import (
	"os"
	"time"
)

// Get returns the value of the environment variable named key, or def if unset/empty.
func Get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Duration parses a Go duration (e.g. "24h") from the env var named key,
// falling back to def on missing or invalid values.
func Duration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// Bool reads a boolean-ish env var ("1", "true", "yes"); def otherwise.
func Bool(key string, def bool) bool {
	switch os.Getenv(key) {
	case "1", "true", "TRUE", "yes":
		return true
	case "0", "false", "FALSE", "no":
		return false
	default:
		return def
	}
}
