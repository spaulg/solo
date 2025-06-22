package container

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestOrchestratorFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(OrchestratorFactoryTestSuite))
}
