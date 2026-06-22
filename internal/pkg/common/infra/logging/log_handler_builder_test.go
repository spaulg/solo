package logging

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestLogHandlerBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(LogHandlerBuilderTestSuite))
}

type LogHandlerBuilderTestSuite struct {
	suite.Suite

	logFilePath string
	builder     *LogHandlerBuilder
}

func (t *LogHandlerBuilderTestSuite) SetupTest() {
	t.builder = NewLogHandlerBuilder()
	t.logFilePath = t.T().TempDir() + "/logfile.log"
}

func (t *LogHandlerBuilderTestSuite) TestTextHandler() {
	t.builder.WithLogFilePath(t.logFilePath).
		WithLogLevel("error").
		WithLogHandlerName("text")

	handler, err := t.builder.Build()
	t.NoError(err)
	t.NotNil(handler)

	t.IsType(&slog.TextHandler{}, handler)
}

func (t *LogHandlerBuilderTestSuite) TestJsonHandler() {
	t.builder.WithLogFilePath(t.logFilePath).
		WithLogLevel("error").
		WithLogHandlerName("json")

	handler, err := t.builder.Build()
	t.NoError(err)
	t.NotNil(handler)

	t.IsType(&slog.JSONHandler{}, handler)
}
