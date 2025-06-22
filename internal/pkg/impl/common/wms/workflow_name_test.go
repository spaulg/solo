package wms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestWorkflowNameTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowNameTestSuite))
}
