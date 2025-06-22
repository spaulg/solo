package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
	"github.com/spaulg/solo/test/mocks/host/project"
	"github.com/stretchr/testify/suite"
)

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

	workflowConfig := project_types.ServiceWorkflowConfig{
		Steps: make([]project_types.WorkflowStep, 0),
	}

	t.mockProject.On("GetServiceWorkflow", serviceName, workflowName.String()).Return(workflowConfig)

	workflowFactory := NewWorkflowFactory()
	workflow := workflowFactory.Make(t.mockProject, serviceName, workflowName)

	t.NotNil(workflow)
}
