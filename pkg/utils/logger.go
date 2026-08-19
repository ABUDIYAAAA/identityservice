package utils

import (
	"log/slog"
	"os"
)

func CreateLoggers() (*slog.Logger, *slog.Logger, *slog.Logger) {
	infoHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	infoLogger := slog.New(infoHandler)

	warnHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})
	warnLogger := slog.New(warnHandler)

	errorHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelError,
		AddSource: true,
	})
	errorLogger := slog.New(errorHandler)

	return infoLogger, warnLogger, errorLogger

}
