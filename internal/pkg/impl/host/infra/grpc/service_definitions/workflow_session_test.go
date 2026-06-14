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
	"github.com/spaulg/solo/internal/pkg/impl/host/app/wms/wf"
	interceptors2 "github.com/spaulg/solo/internal/pkg/impl/host/infra/grpc/interceptors"
	"github.com/spaulg/solo/test/mocks/grpc"
	"github.com/spaulg/solo/test/mocks/host/app/wms"
	"github.com/spaulg/solo/test/mocks/host/domain/compose"
	"github.com/spaulg/solo/test/mocks/host/infra/container"
	"github.com/spaulg/solo/test/mocks/logging"
)

func TestWorkflowSessionTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowSessionTestSuite))
}

type WorkflowSessionTestSuite struct {
	suite.Suite

	mockLogger              *slog.Logger
	mockProject             *compose.MockProject
	mockLogHandler          *logging.MockHandler
	mockOrchestrator        *container.MockOrchestrator
	mockWorkflowExecTracker *wms.MockWorkflowExecTracker
	mockGrpcServer          *grpc.MockBidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse]
}

func (t *WorkflowSessionTestSuite) SetupTest() {
	t.mockGrpcServer = &grpc.MockBidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse]{}
	t.mockWorkflowExecTracker = &wms.MockWorkflowExecTracker{}
	t.mockOrchestrator = &container.MockOrchestrator{}
	t.mockProject = &compose.MockProject{}
	t.mockLogHandler = &logging.MockHandler{}

	t.mockLogHandler.On("Enabled", mock.Anything, mock.Anything).
		Maybe().
		Return(true)

	t.mockLogger = slog.New(t.mockLogHandler)
}

func (t *WorkflowSessionTestSuite) TestMissingServiceName() {
	ctx := context.Background()
	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Service name not found"
	})).Return(nil)

	workflowSession, err := NewWorkflowSession(
		t.mockLogger,
		t.mockProject,
		commonworkflow.PreStartService,
		t.mockGrpcServer,
		t.mockWorkflowExecTracker,
		t.mockOrchestrator,
	)

	t.Nil(workflowSession)
	t.Error(err, "unauthorized")

	t.mockLogHandler.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowSessionTestSuite) TestMissingContainerName() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Container name not found"
	})).Return(nil)

	workflowSession, err := NewWorkflowSession(
		t.mockLogger,
		t.mockProject,
		commonworkflow.PreStartService,
		t.mockGrpcServer,
		t.mockWorkflowExecTracker,
		t.mockOrchestrator,
	)

	t.Nil(workflowSession)
	t.Error(err, "unauthorized")

	t.mockLogHandler.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowSessionTestSuite) TestMissingFullContainerName() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Full container name not found"
	})).Return(nil)

	workflowSession, err := NewWorkflowSession(
		t.mockLogger,
		t.mockProject,
		commonworkflow.PreStartService,
		t.mockGrpcServer,
		t.mockWorkflowExecTracker,
		t.mockOrchestrator,
	)

	t.Nil(workflowSession)
	t.Error(err, "unauthorized")

	t.mockLogHandler.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowSessionTestSuite) TestAccessors() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.mockLogger,
		t.mockProject,
		commonworkflow.PreStartService,
		t.mockGrpcServer,
		t.mockWorkflowExecTracker,
		t.mockOrchestrator,
	)

	t.NotNil(workflowSession)
	t.NoError(err)

	t.Equal(commonworkflow.PreStartService, workflowSession.GetWorkflowName())
	t.Equal("test_service", workflowSession.GetServiceName())
	t.Equal("test_service-1", workflowSession.GetFullContainerName())
	t.Equal("service-1", workflowSession.GetContainerName())
	t.Equal(false, workflowSession.HasFirstContainerWorkflowRun())

	t.mockLogHandler.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowSessionTestSuite) TestHasFirstContainerWorkflowRunTrue() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors2.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "true")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.mockLogger,
		t.mockProject,
		commonworkflow.FirstPreStartContainer,
		t.mockGrpcServer,
		t.mockWorkflowExecTracker,
		t.mockOrchestrator,
	)

	t.NotNil(workflowSession)
	t.NoError(err)

	t.Equal(true, workflowSession.HasFirstContainerWorkflowRun())

	t.mockLogHandler.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowSessionTestSuite) TestHasFirstContainerWorkflowRunFalse() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors2.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.mockLogger,
		t.mockProject,
		commonworkflow.FirstPreStartContainer,
		t.mockGrpcServer,
		t.mockWorkflowExecTracker,
		t.mockOrchestrator,
	)

	t.NotNil(workflowSession)
	t.NoError(err)

	t.Equal(false, workflowSession.HasFirstContainerWorkflowRun())

	t.mockLogHandler.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowSessionTestSuite) TestRunCommand() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors2.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.mockLogger,
		t.mockProject,
		commonworkflow.FirstPreStartContainer,
		t.mockGrpcServer,
		t.mockWorkflowExecTracker,
		t.mockOrchestrator,
	)

	t.NotNil(workflowSession)
	t.NoError(err)

	t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
		Action: services.WorkflowAction_RUN_COMMAND_ACTION,
		RunCommand: &services.WorkflowRunCommand{
			Command:          "test_command",
			Arguments:        []string{"arg1", "arg2"},
			WorkingDirectory: "/tmp",
		},
	}).Return(nil)

	err = workflowSession.RunCommand(&wf.RunCommandRequest{
		Command:          "test_command",
		Arguments:        []string{"arg1", "arg2"},
		WorkingDirectory: "/tmp",
	})

	t.NoError(err)

	t.mockLogHandler.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowSessionTestSuite) TestRecvReturnsError() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors2.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.mockLogger,
		t.mockProject,
		commonworkflow.FirstPreStartContainer,
		t.mockGrpcServer,
		t.mockWorkflowExecTracker,
		t.mockOrchestrator,
	)

	t.NotNil(workflowSession)
	t.NoError(err)

	t.mockGrpcServer.On("Recv").Return(nil, errors.New("test error"))
	response, err := workflowSession.RecvCommandResponse()
	t.Nil(response)
	t.ErrorContains(err, "test error")

	t.mockLogHandler.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
	t.mockWorkflowExecTracker.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowSessionTestSuite) TestRecvIncorrectCommandResult() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors2.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.mockLogger,
		t.mockProject,
		commonworkflow.FirstPreStartContainer,
		t.mockGrpcServer,
		t.mockWorkflowExecTracker,
		t.mockOrchestrator,
	)

	t.NotNil(workflowSession)
	t.NoError(err)

	t.mockGrpcServer.On("Recv").Return(&services.WorkflowStreamRequest{
		Result: services.WorkflowResult_UNKNOWN_WORKFLOW_RESULT,
	}, nil)

	response, err := workflowSession.RecvCommandResponse()
	t.Nil(response)
	t.ErrorContains(err, "unknown result")
}

func (t *WorkflowSessionTestSuite) TestRecvWithoutExitCode() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors2.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.mockLogger,
		t.mockProject,
		commonworkflow.FirstPreStartContainer,
		t.mockGrpcServer,
		t.mockWorkflowExecTracker,
		t.mockOrchestrator,
	)

	t.NotNil(workflowSession)
	t.NoError(err)

	t.mockGrpcServer.On("Recv").Return(&services.WorkflowStreamRequest{
		Result: services.WorkflowResult_RUN_COMMAND_RESULT,
		RunCommandResult: &services.WorkflowRunResult{
			Stdout:   "stdout data",
			Stderr:   "stderr data",
			ExitCode: nil,
		},
	}, nil)

	response, err := workflowSession.RecvCommandResponse()
	t.NoError(err)

	t.Equal("stdout data", response.Stdout)
	t.Equal("stderr data", response.Stderr)
	t.Nil(response.ExitCode)
}

func (t *WorkflowSessionTestSuite) TestRecvWithExitCode() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors2.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.mockLogger,
		t.mockProject,
		commonworkflow.FirstPreStartContainer,
		t.mockGrpcServer,
		t.mockWorkflowExecTracker,
		t.mockOrchestrator,
	)

	t.NotNil(workflowSession)
	t.NoError(err)

	exitCode := uint32(100)
	expectedExitCode := uint8(exitCode)

	t.mockGrpcServer.On("Recv").Return(&services.WorkflowStreamRequest{
		Result: services.WorkflowResult_RUN_COMMAND_RESULT,
		RunCommandResult: &services.WorkflowRunResult{
			Stdout:   "stdout data",
			Stderr:   "stderr data",
			ExitCode: &exitCode,
		},
	}, nil)

	response, err := workflowSession.RecvCommandResponse()
	t.NoError(err)

	t.Equal("stdout data", response.Stdout)
	t.Equal("stderr data", response.Stderr)
	t.Equal(&expectedExitCode, response.ExitCode)
}

func (t *WorkflowSessionTestSuite) TestMarkComplete() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors2.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.mockLogger,
		t.mockProject,
		commonworkflow.FirstPreStartContainer,
		t.mockGrpcServer,
		t.mockWorkflowExecTracker,
		t.mockOrchestrator,
	)

	t.NotNil(workflowSession)
	t.NoError(err)

	t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
		Action: services.WorkflowAction_COMPLETE_ACTION,
	}).Return(nil)

	err = workflowSession.MarkCompletion()
	t.NoError(err)
}
