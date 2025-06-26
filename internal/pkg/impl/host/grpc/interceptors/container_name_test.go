package interceptors

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestContainerNameTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerNameTestSuite))
}
