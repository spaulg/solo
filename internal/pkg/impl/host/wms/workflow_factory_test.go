package wms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestWorkflowFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowFactoryTestSuite))
}
