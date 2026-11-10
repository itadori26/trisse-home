// Package logging builds the structured slog logger used across the server.
//
// Ring: Frameworks & Drivers.
package logging

import (
	"io"
	"log/slog"
)

// New returns a logger writing to w in the given format ("json" or "text").
func New(w io.Writer, level slog.Level, format string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level}
	if format == "text" {
		return slog.New(slog.NewTextHandler(w, opts))
	}
	return slog.New(slog.NewJSONHandler(w, opts))
}
