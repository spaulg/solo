package interceptors

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestServiceNameTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceNameTestSuite))
}
