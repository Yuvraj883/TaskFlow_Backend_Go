package utils

import (
	"log/slog"
	"os"
)

// InitLogger sets up slog to use JSON logging to stdout
func InitLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
