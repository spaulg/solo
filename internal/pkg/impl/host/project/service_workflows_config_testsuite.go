package project

import "github.com/stretchr/testify/suite"

type ServiceWorkflowsTestSuite struct {
	suite.Suite
}

func (t *ServiceWorkflowsTestSuite) TestNewServiceWorkflows() {
	serviceWorkflow := NewServiceWorkflows()
	t.NotNil(serviceWorkflow)
}
