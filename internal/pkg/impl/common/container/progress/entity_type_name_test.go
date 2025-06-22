package progress

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestProgressEntityTypeNameTestSuite(t *testing.T) {
	suite.Run(t, new(ProgressEntityTypeNameTestSuite))
}
