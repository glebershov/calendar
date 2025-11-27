package logger

import (
    "log/slog"
    "os"
    "strings"
)

// Logger — просто алиас, чтобы было удобно передавать.
type Logger = *slog.Logger

// InitLogger создаёт и настраивает логгер по строковому уровню.
func InitLogger(level string) Logger {
    slogLevel := parseLevel(level)

    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slogLevel,
    })

    logger := slog.New(handler)

    // По желанию: сделать его глобальным по умолчанию.
    slog.SetDefault(logger)

    return logger
}

// parseLevel переводит строковый уровень в slog.Level.
func parseLevel(level string) slog.Level {
    switch strings.ToLower(level) {
    case "debug":
        return slog.LevelDebug
    case "info":
        return slog.LevelInfo
    case "warn", "warning":
        return slog.LevelWarn
    case "error":
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}
