package logger

import (
	"log/slog"
	"os"
)

func InitStructuredLogger(level slog.Leveler) {
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	slog.SetDefault(slog.New(jsonHandler))
}
