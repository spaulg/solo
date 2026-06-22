package logging

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestBlackholeHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(BlackholeHandlerTestSuite))
}

type BlackholeHandlerTestSuite struct {
	suite.Suite
}

func (t *BlackholeHandlerTestSuite) TestHandler() {
	handler := NewBlackHoleHandler()

	t.False(handler.Enabled(context.Background(), slog.LevelDebug))

	t.Nil(handler.Handle(context.Background(), slog.Record{
		Level:   slog.LevelDebug,
		Message: "This is a test message",
	}))

	t.Equal(handler, handler.WithAttrs([]slog.Attr{
		slog.String("key", "value"),
	}))

	t.Equal(handler, handler.WithGroup("test-group"))
}
