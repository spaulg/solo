package events

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestEventManagerTestSuite(t *testing.T) {
	suite.Run(t, new(EventManagerTestSuite))
}
