package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

// Init initializes the global slog logger based on the environment.
// It uses a human-readable text handler for "dev" and a structured
// JSON handler for "prod".
func Init(env string) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		// You can set a log level here if needed, e.g., slog.LevelDebug
		Level: level(env),
	}

	if env == "prod" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// 设置为全局 Logger
	Log = slog.New(handler)
	slog.SetDefault(Log)
}

func level(env string) slog.Level {
	if env == "prod" {
		return slog.LevelInfo
	}
	return slog.LevelDebug
}
