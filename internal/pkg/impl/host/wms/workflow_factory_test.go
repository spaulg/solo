package wms

import (
	"testing"

	"github.com/stretchr/testify/suite"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	compose_types "github.com/spaulg/solo/internal/pkg/types/host/project/compose"
	"github.com/spaulg/solo/test/mocks/host/project"
	"github.com/spaulg/solo/test/mocks/host/project/compose"
)

func TestWorkflowFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowFactoryTestSuite))
}

type WorkflowFactoryTestSuite struct {
	suite.Suite

	mockProject *project.MockProject
}

func (t *WorkflowFactoryTestSuite) SetupTest() {
	t.mockProject = &project.MockProject{}
}

func (t *WorkflowFactoryTestSuite) TestBuild() {
	serviceName := "test"
	workflowName := workflowcommon.FirstPreStart

	workflowConfig := compose_types.ServiceWorkflowConfig{
		Steps: make([]compose_types.WorkflowStep, 0),
	}

	mockServices := &compose.MockServices{}
	mockServiceConfig := &compose.MockServiceConfig{}

	t.mockProject.On("Services").Return(mockServices)
	mockServices.On("GetService", serviceName).Return(mockServiceConfig)
	mockServiceConfig.On("GetServiceWorkflow", workflowName.String()).Return(workflowConfig)

	workflowFactory := NewWorkflowFactory()
	workflow := workflowFactory.Make(t.mockProject, serviceName, workflowName)

	t.NotNil(workflow)

	t.mockProject.AssertExpectations(t.T())
}
