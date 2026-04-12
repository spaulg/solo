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
	wms_shared "github.com/spaulg/solo/internal/pkg/impl/host/shared/wms"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/grpc"
	"github.com/spaulg/solo/test/mocks/host/app/wms"
	"github.com/spaulg/solo/test/mocks/host/domain/project"
	"github.com/spaulg/solo/test/mocks/host/infra/container"
	"github.com/spaulg/solo/test/mocks/logging"
)

func TestWorkflowSessionTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowSessionTestSuite))
}

type WorkflowSessionTestSuite struct {
	suite.Suite

	soloCtx                 *cli_context.CliContext
	mockProject             *project.MockProject
	mockLogHandler          *logging.MockHandler
	mockOrchestrator        *container.MockOrchestrator
	mockWorkflowExecTracker *wms.MockWorkflowExecTracker
	mockGrpcServer          *grpc.MockBidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse]
}

func (t *WorkflowSessionTestSuite) SetupTest() {
	t.mockGrpcServer = &grpc.MockBidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse]{}
	t.mockWorkflowExecTracker = &wms.MockWorkflowExecTracker{}
	t.mockOrchestrator = &container.MockOrchestrator{}
	t.mockProject = &project.MockProject{}
	t.mockLogHandler = &logging.MockHandler{}

	t.mockLogHandler.On("Enabled", mock.Anything, mock.Anything).
		Maybe().
		Return(true)

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

func (t *WorkflowSessionTestSuite) TestMissingServiceName() {
	ctx := context.Background()
	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Service name not found"
	})).Return(nil)

	workflowSession, err := NewWorkflowSession(
		t.soloCtx,
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
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Container name not found"
	})).Return(nil)

	workflowSession, err := NewWorkflowSession(
		t.soloCtx,
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
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Full container name not found"
	})).Return(nil)

	workflowSession, err := NewWorkflowSession(
		t.soloCtx,
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
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.soloCtx,
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
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "true")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.soloCtx,
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
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.soloCtx,
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
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.soloCtx,
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

	err = workflowSession.RunCommand(&wms_shared.RunCommandRequest{
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
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.soloCtx,
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
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.soloCtx,
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
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.soloCtx,
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
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	firstContainerCompleteValueName := interceptors.FirstContainerComplete(commonworkflow.FirstPreStartContainer)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, firstContainerCompleteValueName, "false")

	t.mockGrpcServer.On("Context").Return(ctx)

	workflowSession, err := NewWorkflowSession(
		t.soloCtx,
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
