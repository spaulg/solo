package wms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestWorkflowGuardTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowGuardTestSuite))
}
