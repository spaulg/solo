package logging

import (
	"context"
	"log/slog"

	"github.com/stretchr/testify/mock"
)

type MockHandler struct {
	mock.Mock
}

func (h *MockHandler) Enabled(ctx context.Context, record slog.Level) bool {
	args := h.Called(ctx, record)
	return args.Bool(0)
}

func (h *MockHandler) Handle(ctx context.Context, record slog.Record) error {
	args := h.Called(ctx, record)
	return args.Error(0)
}

func (h *MockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.Called(attrs)
	return h
}

func (h *MockHandler) WithGroup(name string) slog.Handler {
	h.Called(name)
	return h
}
