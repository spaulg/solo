package progress

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestComposeProgressActionTestSuite(t *testing.T) {
	suite.Run(t, new(ComposeProgressActionTestSuite))
}
