package host

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestProjectControlTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectControlTestSuite))
}
