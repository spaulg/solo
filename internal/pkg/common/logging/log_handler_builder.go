package logging

import (
	"fmt"
	"log/slog"
	"os"
)

type LogHandlerBuilder struct {
	logFilePath string
	level       slog.Level
	handler     string
}

func NewLogHandlerBuilder() *LogHandlerBuilder {
	return &LogHandlerBuilder{}
}

func (t *LogHandlerBuilder) WithLogFilePath(logFilePath string) *LogHandlerBuilder {
	t.logFilePath = logFilePath
	return t
}

func (t *LogHandlerBuilder) WithLogLevel(level string) *LogHandlerBuilder {
	switch level {
	case "debug":
		t.level = slog.LevelDebug
	case "info":
		t.level = slog.LevelInfo
	case "error":
		t.level = slog.LevelError
	case "warning":
		fallthrough
	default:
		t.level = slog.LevelWarn
	}

	return t
}

func (t *LogHandlerBuilder) WithLogHandlerName(handler string) *LogHandlerBuilder {
	t.handler = handler
	return t
}

func (t *LogHandlerBuilder) Build() (slog.Handler, error) {
	logFile, err := os.OpenFile(t.logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	switch t.handler {
	case "json":
		return slog.NewJSONHandler(logFile, &slog.HandlerOptions{
			Level: t.level,
		}), nil

	case "text":
		fallthrough
	default:
		return slog.NewTextHandler(logFile, &slog.HandlerOptions{
			Level: t.level,
		}), nil
	}
}
