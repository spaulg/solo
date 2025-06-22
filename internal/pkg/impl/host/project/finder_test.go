package project

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestFinderTestSuite(t *testing.T) {
	suite.Run(t, new(FinderTestSuite))
}
