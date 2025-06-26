package interceptors

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestFirstPreStartCompleteTestSuite(t *testing.T) {
	suite.Run(t, new(FirstPreStartCompleteTestSuite))
}
