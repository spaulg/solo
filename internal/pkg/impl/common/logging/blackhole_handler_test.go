package logging

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestBlackholeHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(BlackholeHandlerTestSuite))
}
