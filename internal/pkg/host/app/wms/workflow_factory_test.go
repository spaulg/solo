package wms

import (
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	workflowcommon "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
	compose_types "github.com/spaulg/solo/internal/pkg/host/domain/compose"
	"github.com/spaulg/solo/internal/pkg/host/domain/config"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/domain/compose"
	"github.com/spaulg/solo/test/mocks/logging"
)

func TestWorkflowFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowFactoryTestSuite))
}

type WorkflowFactoryTestSuite struct {
	suite.Suite

	mockProject    *compose.MockProject
	mockConfig     *domain2.Config
	mockLogHandler *logging.MockHandler
}

func (t *WorkflowFactoryTestSuite) SetupTest() {
	t.mockProject = &compose.MockProject{}

	t.mockLogHandler = &logging.MockHandler{}
	t.mockLogHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)

	t.mockConfig = &domain2.Config{
		Entrypoint: config.EntrypointConfig{
			HostEntrypointPath: test.GetTestDataFilePath("entrypoint.sh"),
		},
		Workflow: config.WorkflowConfig{
			Grpc: config.GrpcConfig{
				ServerPort: 0,
			},
		},
	}
}

func (t *WorkflowFactoryTestSuite) TestBuild() {
	serviceName := "test"
	workflowName := workflowcommon.FirstPreStartContainer

	workflowConfig := compose_types.NewServiceWorkflowConfig(make([]domain2.WorkflowStep, 0), nil, nil)

	mockServices := &compose.MockServices{}
	mockServiceConfig := &compose.MockServiceConfig{}

	t.mockProject.On("Services").Return(mockServices)
	mockServices.On("GetService", serviceName).Return(mockServiceConfig)
	mockServiceConfig.On("GetServiceWorkflow", workflowName.String()).Return(workflowConfig)
	mockServiceConfig.On("GetConfig").Return(types.ServiceConfig{})

	workflowFactory := NewWorkflowFactory()
	workflow, err := workflowFactory.Make(t.mockConfig, t.mockProject, serviceName, "/", workflowName)

	t.NotNil(workflow)
	t.NoError(err)

	t.mockProject.AssertExpectations(t.T())
}
