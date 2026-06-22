package service_definitions

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	commonworkflow "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/common/infra/grpc/services"
	interceptors2 "github.com/spaulg/solo/internal/pkg/host/infra/grpc/interceptors"
	"github.com/spaulg/solo/test/mocks/grpc"
	"github.com/spaulg/solo/test/mocks/host/app/wms"
	"github.com/spaulg/solo/test/mocks/host/domain/compose"
	"github.com/spaulg/solo/test/mocks/host/infra/container"
	"github.com/spaulg/solo/test/mocks/logging"
)

func TestWorkflowServiceTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowServiceTestSuite))
}

type WorkflowServiceTestSuite struct {
	suite.Suite

	mockLogger              *slog.Logger
	mockProject             *compose.MockProject
	mockLogHandler          *logging.MockHandler
	mockWorkflowExecTracker *wms.MockWorkflowExecTracker
	mockOrchestrator        *container.MockOrchestrator
	mockGrpcServer          *grpc.MockBidiStreamingServer[services.RunWorkflowStreamRequest, services.WorkflowStreamResponse]
	mockWorkflowRunner      *wms.MockWorkflowRunner
}

func (t *WorkflowServiceTestSuite) SetupTest() {
	t.mockProject = &compose.MockProject{}
	t.mockOrchestrator = &container.MockOrchestrator{}
	t.mockGrpcServer = &grpc.MockBidiStreamingServer[services.RunWorkflowStreamRequest, services.WorkflowStreamResponse]{}
	t.mockWorkflowExecTracker = &wms.MockWorkflowExecTracker{}
	t.mockWorkflowRunner = &wms.MockWorkflowRunner{}

	t.mockLogHandler = &logging.MockHandler{}
	t.mockLogHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)

	t.mockLogger = slog.New(t.mockLogHandler)
}

func (t *WorkflowServiceTestSuite) TestRecvRunWorkflowStreamRequestFailed() {
	workflowService := NewWorkflowService(
		t.mockLogger,
		t.mockProject,
		t.mockOrchestrator,
		t.mockWorkflowExecTracker,
		t.mockWorkflowRunner,
	)

	t.NotNil(workflowService)

	t.mockGrpcServer.On("Recv").Return(nil, errors.New("recv error")).Once()

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.Error(err)

	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockWorkflowRunner.AssertExpectations(t.T())
}

func (t *WorkflowServiceTestSuite) TestRecvRunWorkflowStreamRequestMessageUnsupported() {
	workflowService := NewWorkflowService(
		t.mockLogger,
		t.mockProject,
		t.mockOrchestrator,
		t.mockWorkflowExecTracker,
		t.mockWorkflowRunner,
	)

	t.NotNil(workflowService)

	message := services.RunWorkflowStreamRequest{Request: nil}
	t.mockGrpcServer.On("Recv").Return(&message, nil).Once()

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.ErrorContains(err, "unsupported first message")

	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockWorkflowRunner.AssertExpectations(t.T())
}

func (t *WorkflowServiceTestSuite) TestFirstContainerCompleteSkipsWorkflow() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)
	firstPreStartCompleteContextValueName := interceptors2.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

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

	t.mockWorkflowRunner.On("RunWorkflow", mock.AnythingOfType("*service_definitions.WorkflowSession")).Return(nil)

	workflowService := NewWorkflowService(
		t.mockLogger,
		t.mockProject,
		t.mockOrchestrator,
		t.mockWorkflowExecTracker,
		t.mockWorkflowRunner,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockWorkflowRunner.AssertExpectations(t.T())
}

func (t *WorkflowServiceTestSuite) TestRunWorkflowSucceeds() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)
	firstPreStartCompleteContextValueName := interceptors2.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

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
	).Return(nil).Once()

	workflowService := NewWorkflowService(
		t.mockLogger,
		t.mockProject,
		t.mockOrchestrator,
		t.mockWorkflowExecTracker,
		t.mockWorkflowRunner,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockWorkflowRunner.AssertExpectations(t.T())
}

func (t *WorkflowServiceTestSuite) TestRunWorkflowFails() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)
	firstPreStartCompleteContextValueName := interceptors2.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

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
	).Return(errors.New("mock wf error")).Once()

	workflowService := NewWorkflowService(
		t.mockLogger,
		t.mockProject,
		t.mockOrchestrator,
		t.mockWorkflowExecTracker,
		t.mockWorkflowRunner,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.ErrorContains(err, "mock wf error")

	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockWorkflowRunner.AssertExpectations(t.T())
}
