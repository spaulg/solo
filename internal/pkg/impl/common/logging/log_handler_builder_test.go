package logging

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestLogHandlerBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(LogHandlerBuilderTestSuite))
}
