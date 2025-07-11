package project

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestServiceWorkflowsTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceWorkflowsTestSuite))
}

type ServiceWorkflowsTestSuite struct {
	suite.Suite
}

func (t *ServiceWorkflowsTestSuite) TestNewServiceWorkflows() {
	serviceWorkflow := NewServiceWorkflows()
	t.NotNil(serviceWorkflow)
}
