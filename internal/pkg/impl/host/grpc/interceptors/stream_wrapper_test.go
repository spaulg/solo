package interceptors

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestStreamWrapperTestSuite(t *testing.T) {
	suite.Run(t, new(StreamWrapperTestSuite))
}
