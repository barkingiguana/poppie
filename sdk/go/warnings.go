package poppie

import "log/slog"

// VersionWarning contains version negotiation information from the server.
type VersionWarning struct {
	// Status is the server's assessment: "supported", "deprecated", or "unknown".
	Status string
	// Message is an optional human-readable deprecation message.
	Message string
}

// WarningHandler is a callback invoked when the server sends a non-"supported" version status.
type WarningHandler func(VersionWarning)

// DefaultWarningHandler logs version warnings via slog.
func DefaultWarningHandler(w VersionWarning) {
	slog.Warn("poppie server version warning",
		"status", w.Status,
		"message", w.Message,
	)
}
