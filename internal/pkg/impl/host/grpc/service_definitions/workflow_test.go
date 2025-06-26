package service_definitions

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuite))
}
