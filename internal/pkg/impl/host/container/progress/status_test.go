package progress

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}
