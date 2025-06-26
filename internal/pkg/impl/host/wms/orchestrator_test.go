package wms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestOrchestratorTestSuite(t *testing.T) {
	suite.Run(t, new(OrchestratorTestSuite))
}
