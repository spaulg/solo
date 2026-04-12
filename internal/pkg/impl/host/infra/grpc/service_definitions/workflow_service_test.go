package service_definitions

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/common/infra/grpc/services"
	cli_context "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	domain_config "github.com/spaulg/solo/internal/pkg/impl/host/domain/config"
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/grpc/interceptors"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/app/wms"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/grpc"
	"github.com/spaulg/solo/test/mocks/host/app/events"
	"github.com/spaulg/solo/test/mocks/host/app/wms"
	"github.com/spaulg/solo/test/mocks/host/domain/project"
	"github.com/spaulg/solo/test/mocks/host/infra/container"
	"github.com/spaulg/solo/test/mocks/logging"
)

func TestWorkflowServiceTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowServiceTestSuite))
}

type WorkflowServiceTestSuite struct {
	suite.Suite

	soloCtx                 *cli_context.CliContext
	mockProject             *project.MockProject
	mockLogHandler          *logging.MockHandler
	mockEventManager        *events.MockEventManager
	mockWorkflowExecTracker *wms.MockWorkflowExecTracker
	mockOrchestrator        *container.MockOrchestrator
	mockGrpcServer          *grpc.MockBidiStreamingServer[services.RunWorkflowStreamRequest, services.WorkflowStreamResponse]
	mockWorkflowRunner      *wms.MockWorkflowRunner
}

func (t *WorkflowServiceTestSuite) SetupTest() {
	t.mockEventManager = &events.MockEventManager{}
	t.mockProject = &project.MockProject{}
	t.mockOrchestrator = &container.MockOrchestrator{}
	t.mockGrpcServer = &grpc.MockBidiStreamingServer[services.RunWorkflowStreamRequest, services.WorkflowStreamResponse]{}
	t.mockWorkflowExecTracker = &wms.MockWorkflowExecTracker{}
	t.mockWorkflowRunner = &wms.MockWorkflowRunner{}

	t.mockLogHandler = &logging.MockHandler{}
	t.mockLogHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)

	t.soloCtx = &cli_context.CliContext{
		Project: t.mockProject,
		Logger:  slog.New(t.mockLogHandler),
		Config: &domain.Config{
			Entrypoint: domain_config.EntrypointConfig{
				HostEntrypointPath: test.GetTestDataFilePath("entrypoint.sh"),
			},
			Workflow: domain_config.WorkflowConfig{
				Grpc: domain_config.GrpcConfig{
					ServerPort: 0,
				},
			},
		},
	}
}

func (t *WorkflowServiceTestSuite) TestRecvRunWorkflowStreamRequestFailed() {
	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowExecTracker,
		t.mockWorkflowRunner,
	)

	t.NotNil(workflowService)

	t.mockGrpcServer.On("Recv").Return(nil, errors.New("recv error")).Once()

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockWorkflowRunner.AssertExpectations(t.T())
}

func (t *WorkflowServiceTestSuite) TestRecvRunWorkflowStreamRequestMessageUnsupported() {
	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowExecTracker,
		t.mockWorkflowRunner,
	)

	t.NotNil(workflowService)

	message := services.RunWorkflowStreamRequest{Request: nil}
	t.mockGrpcServer.On("Recv").Return(&message, nil).Once()

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.ErrorContains(err, "unsupported first message")

	t.mockEventManager.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockWorkflowRunner.AssertExpectations(t.T())
}

func (t *WorkflowServiceTestSuite) TestFirstContainerCompleteSkipsWorkflow() {
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
	firstPreStartCompleteContextValueName := interceptors.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, firstPreStartCompleteContextValueName, "true")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
		Request: &services.RunWorkflowStreamRequest_RunRequest{
			RunRequest: &services.WorkflowRunRequest{
				WorkflowName: commonworkflow.FirstPreStartContainer.String(),
			},
		},
	}, nil).Once()

	t.mockEventManager.On("Publish", &wms_types.WorkflowSkippedEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      commonworkflow.FirstPreStartContainer,
		},
		Successful: true,
	}).Return()

	t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
		Action:     services.WorkflowAction_COMPLETE_ACTION,
		RunCommand: nil,
	}).Return(nil)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowExecTracker,
		t.mockWorkflowRunner,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockWorkflowRunner.AssertExpectations(t.T())
}

func (t *WorkflowServiceTestSuite) TestRunWorkflowSucceeds() {
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
	firstPreStartCompleteContextValueName := interceptors.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, firstPreStartCompleteContextValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
		Request: &services.RunWorkflowStreamRequest_RunRequest{
			RunRequest: &services.WorkflowRunRequest{
				WorkflowName: commonworkflow.FirstPreStartContainer.String(),
			},
		},
	}, nil).Once()

	t.mockWorkflowRunner.On(
		"RunWorkflow",
		mock.AnythingOfType("*service_definitions.WorkflowSession"),
	).Return(true, nil).Once()

	t.mockEventManager.On("Publish", &wms_types.WorkflowCompleteEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      commonworkflow.FirstPreStartContainer,
		},
		Successful: true,
	}).Return()

	t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
		Action:     services.WorkflowAction_COMPLETE_ACTION,
		RunCommand: nil,
	}).Return(nil)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowExecTracker,
		t.mockWorkflowRunner,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockWorkflowRunner.AssertExpectations(t.T())
}

func (t *WorkflowServiceTestSuite) TestRunWorkflowFails() {
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
	firstPreStartCompleteContextValueName := interceptors.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, firstPreStartCompleteContextValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
		Request: &services.RunWorkflowStreamRequest_RunRequest{
			RunRequest: &services.WorkflowRunRequest{
				WorkflowName: commonworkflow.FirstPreStartContainer.String(),
			},
		},
	}, nil).Once()

	t.mockWorkflowRunner.On(
		"RunWorkflow",
		mock.AnythingOfType("*service_definitions.WorkflowSession"),
	).Return(false, errors.New("mock workflow error")).Once()

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowExecTracker,
		t.mockWorkflowRunner,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.ErrorContains(err, "mock workflow error")

	t.mockEventManager.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockWorkflowRunner.AssertExpectations(t.T())
}
