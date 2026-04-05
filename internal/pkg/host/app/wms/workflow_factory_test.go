package wms

import (
	"log/slog"
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	cli_context "github.com/spaulg/solo/internal/pkg/host/app/context"
	"github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/internal/pkg/host/domain/config"
	compose_types "github.com/spaulg/solo/internal/pkg/shared/domain/project/compose"
	workflowcommon "github.com/spaulg/solo/internal/pkg/shared/domain/wms"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/domain/project"
	"github.com/spaulg/solo/test/mocks/host/domain/project/compose"
	"github.com/spaulg/solo/test/mocks/host/infra/container"
	"github.com/spaulg/solo/test/mocks/logging"
)

func TestWorkflowFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowFactoryTestSuite))
}

type WorkflowFactoryTestSuite struct {
	suite.Suite

	soloCtx          *cli_context.CliContext
	mockProject      *project.MockProject
	mockOrchestrator *container.MockOrchestrator
	mockLogHandler   *logging.MockHandler
}

func (t *WorkflowFactoryTestSuite) SetupTest() {
	t.mockProject = &project.MockProject{}
	t.mockOrchestrator = &container.MockOrchestrator{}

	t.mockLogHandler = &logging.MockHandler{}
	t.mockLogHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)

	t.soloCtx = &cli_context.CliContext{
		Project: t.mockProject,
		Logger:  slog.New(t.mockLogHandler),
		Config: &domain.Config{
			Entrypoint: config.EntrypointConfig{
				HostEntrypointPath: test.GetTestDataFilePath("entrypoint.sh"),
			},
			Workflow: config.WorkflowConfig{
				Grpc: config.GrpcConfig{
					ServerPort: 0,
				},
			},
		},
	}
}

func (t *WorkflowFactoryTestSuite) TestBuild() {
	serviceName := "test"
	workflowName := workflowcommon.FirstPreStartContainer

	workflowConfig := compose_types.ServiceWorkflowConfig{
		Steps: make([]compose_types.WorkflowStep, 0),
	}

	mockServices := &compose.MockServices{}
	mockServiceConfig := &compose.MockServiceConfig{}

	t.mockProject.On("Services").Return(mockServices)
	mockServices.On("GetService", serviceName).Return(mockServiceConfig)
	mockServiceConfig.On("GetServiceWorkflow", workflowName.String()).Return(workflowConfig)
	mockServiceConfig.On("GetConfig").Return(types.ServiceConfig{})

	t.mockOrchestrator.On("ResolveImageWorkingDirectory", serviceName).Return("/", nil)

	workflowFactory := NewWorkflowFactory()
	workflow, err := workflowFactory.Make(t.soloCtx, t.mockOrchestrator, serviceName, workflowName)

	t.NotNil(workflow)
	t.NoError(err)

	t.mockProject.AssertExpectations(t.T())
}
