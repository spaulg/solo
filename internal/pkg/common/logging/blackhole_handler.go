package logging

import (
	"context"
	"log/slog"
)

type BlackHoleHandler struct{}

func NewBlackHoleHandler() slog.Handler {
	return &BlackHoleHandler{}
}

func (h *BlackHoleHandler) Enabled(context.Context, slog.Level) bool {
	return false
}

func (h *BlackHoleHandler) Handle(context.Context, slog.Record) error {
	return nil
}

func (h *BlackHoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *BlackHoleHandler) WithGroup(name string) slog.Handler {
	return h
}
