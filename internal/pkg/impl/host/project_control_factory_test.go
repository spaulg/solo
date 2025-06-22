package host

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestProjectControlFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectControlFactoryTestSuite))
}
