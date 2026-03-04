package logger

import (
	"log/slog"
	"os"
	"strings"
	"time"
)

var Log *slog.Logger

func Init(level string) {
	opts := &slog.HandlerOptions{
		Level: parseLevel(level),
	}
	Log = slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Helper функции для удобства
func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

func Error(msg string, args ...any) {
	Log.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}

// Middleware для логирования HTTP запросов
func HTTPMiddleware(next interface{}) interface{} {
	// Вернётся функция-обёртка для chi middleware
	return next
}

// Логгер для времени запуска
func Startup(msg string) {
	Log.Info(msg, "timestamp", time.Now().Format(time.RFC3339))
}
