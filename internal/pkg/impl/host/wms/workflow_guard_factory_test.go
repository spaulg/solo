package wms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestWorkflowGuardFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowGuardFactoryTestSuite))
}
