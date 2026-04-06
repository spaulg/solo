package service_definitions

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	cli_context "github.com/spaulg/solo/internal/pkg/host/app/context"
	"github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/internal/pkg/host/domain/config"
	interceptors2 "github.com/spaulg/solo/internal/pkg/host/infra/grpc/interceptors"
	wms3 "github.com/spaulg/solo/internal/pkg/shared/app/wms"
	commonworkflow "github.com/spaulg/solo/internal/pkg/shared/domain/wms"
	"github.com/spaulg/solo/internal/pkg/shared/infra/grpc/services"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/grpc"
	"github.com/spaulg/solo/test/mocks/host/app/events"
	"github.com/spaulg/solo/test/mocks/host/app/wms"
	"github.com/spaulg/solo/test/mocks/host/domain/project"
	"github.com/spaulg/solo/test/mocks/host/infra/container"
	"github.com/spaulg/solo/test/mocks/logging"
)

func TestWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuite))
}

type WorkflowTestSuite struct {
	suite.Suite

	soloCtx                  *cli_context.CliContext
	mockProject              *project.MockProject
	mockLogHandler           *logging.MockHandler
	mockEventManager         *events.MockEventManager
	mockWorkflowExecTracker  *wms.MockWorkflowExecTracker
	mockWorkflowOrchestrator *wms.MockWorkflow
	mockWorkflowFactory      *wms.MockWorkflowFactory
	mockOrchestrator         *container.MockOrchestrator
	mockGrpcServer           *grpc.MockBidiStreamingServer[services.RunWorkflowStreamRequest, services.WorkflowStreamResponse]
}

func (t *WorkflowTestSuite) SetupTest() {
	t.mockEventManager = &events.MockEventManager{}
	t.mockWorkflowOrchestrator = &wms.MockWorkflow{}
	t.mockWorkflowFactory = &wms.MockWorkflowFactory{}
	t.mockProject = &project.MockProject{}
	t.mockOrchestrator = &container.MockOrchestrator{}
	t.mockGrpcServer = &grpc.MockBidiStreamingServer[services.RunWorkflowStreamRequest, services.WorkflowStreamResponse]{}
	t.mockWorkflowExecTracker = &wms.MockWorkflowExecTracker{}

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

// server.Recv() returns err in progress callback
// received packet not services.WorkflowResult_RUN_COMMAND_RESULT in progress
//   - both triggering WorkflowErrorEvent
// non-zero exit code in progress callback

func (t *WorkflowTestSuite) TestMissingServiceName() {
	ctx := context.Background()
	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Service name not found"
	})).Return(nil)

	t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
		Request: &services.RunWorkflowStreamRequest_RunRequest{
			RunRequest: &services.WorkflowRunRequest{
				WorkflowName: commonworkflow.FirstPreStartContainer.String(),
			},
		},
	}, nil).Once()

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)

	t.Error(err, "unauthorized")

	t.mockLogHandler.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestMissingContainerName() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
		Request: &services.RunWorkflowStreamRequest_RunRequest{
			RunRequest: &services.WorkflowRunRequest{
				WorkflowName: commonworkflow.FirstPreStartContainer.String(),
			},
		},
	}, nil).Once()

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Container name not found"
	})).Return(nil)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)

	t.Error(err, "unauthorized")

	t.mockLogHandler.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestMissingFullContainerName() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
		Request: &services.RunWorkflowStreamRequest_RunRequest{
			RunRequest: &services.WorkflowRunRequest{
				WorkflowName: commonworkflow.FirstPreStartContainer.String(),
			},
		},
	}, nil).Once()

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Full container name not found"
	})).Return(nil)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)

	t.Error(err, "unauthorized")

	t.mockLogHandler.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestNilWorkflowFromFactory() {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockEventManager.On("Publish", &wms3.WorkflowStartedEvent{
		BaseWorkflowEvent: wms3.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      commonworkflow.PreStartContainer,
		},
	}).Return()

	t.mockWorkflowFactory.On("Make", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
		Request: &services.RunWorkflowStreamRequest_RunRequest{
			RunRequest: &services.WorkflowRunRequest{
				WorkflowName: commonworkflow.PreStartContainer.String(),
			},
		},
	}, nil).Once()

	t.mockEventManager.On("Publish", &wms3.WorkflowCompleteEvent{
		BaseWorkflowEvent: wms3.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      commonworkflow.PreStartContainer,
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
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestFirstPreStartSkippedWorkflowStepTrigger() {
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

	t.mockEventManager.On("Publish", &wms3.WorkflowSkippedEvent{
		BaseWorkflowEvent: wms3.BaseWorkflowEvent{
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
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestFirstPreStartRecvError() {
	t.testServerRecvErrorFor(commonworkflow.FirstPreStartContainer)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStartRecvError() {
	t.testServerRecvErrorFor(commonworkflow.PreStartContainer)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPostStartRecvError() {
	t.testServerRecvErrorFor(commonworkflow.PostStartContainer)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStopRecvError() {
	t.testServerRecvErrorFor(commonworkflow.PreStopContainer)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreDestroyRecvError() {
	t.testServerRecvErrorFor(commonworkflow.PreDestroyContainer)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestFirstPreStartUnknownWorkflowResult() {
	t.testUnknownWorkflowResult(commonworkflow.FirstPreStartContainer)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStartUnknownWorkflowResult() {
	t.testUnknownWorkflowResult(commonworkflow.PreStartContainer)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPostStartUnknownWorkflowResult() {
	t.testUnknownWorkflowResult(commonworkflow.PostStartContainer)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStopUnknownWorkflowResult() {
	t.testUnknownWorkflowResult(commonworkflow.PreStopContainer)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreDestroyUnknownWorkflowResult() {
	t.testUnknownWorkflowResult(commonworkflow.PreDestroyContainer)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestFirstPreStartZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.FirstPreStartContainer, 0)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStartZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreStartContainer, 0)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPostStartZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PostStartContainer, 0)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStopZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreStopContainer, 0)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreDestroyZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreDestroyContainer, 0)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestFirstPreStartNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.FirstPreStartContainer, 10)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStartNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreStartContainer, 10)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPostStartNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PostStartContainer, 10)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStopNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreStopContainer, 10)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreDestroyNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreDestroyContainer, 10)

	workflowService := NewWorkflowService(
		t.soloCtx,
		t.mockEventManager,
		t.mockOrchestrator,
		t.mockWorkflowFactory,
		t.mockWorkflowExecTracker,
	)

	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
	t.NoError(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) testWorkflowExecFor(workflow commonworkflow.WorkflowName, exitCode uint32) {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockEventManager.On("Publish", &wms3.WorkflowStartedEvent{
		BaseWorkflowEvent: wms3.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      workflow,
		},
	}).Return()

	t.mockWorkflowFactory.On("Make", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(t.mockWorkflowOrchestrator, nil)

	t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
		Request: &services.RunWorkflowStreamRequest_RunRequest{
			RunRequest: &services.WorkflowRunRequest{
				WorkflowName: workflow.String(),
			},
		},
	}, nil).Once()

	t.mockWorkflowOrchestrator.On("StepIterator").Return(func(yield func(wms3.Step) bool) {
		step := &wms.MockStep{}
		step.On("GetID").Return("12345")
		step.On("GetName").Return("test")
		step.On("GetCommand").Return("/bin/sh")
		step.On("GetArguments").Return([]string{"-c", "echo \"Hello World\""})
		step.On("GetWorkingDirectory").Return("/")
		step.On("GetShell").Return("/bin/sh")

		step.On(
			"Trigger",
			mock.AnythingOfType("wms.StepTriggerLambda"),
			mock.AnythingOfType("wms.StepProgressLambda"),
			mock.AnythingOfType("wms.StepCompleteLambda"),
		).Run(func(args mock.Arguments) {
			// trigger
			t.mockEventManager.On("Publish", &wms3.WorkflowStepStartedEvent{
				BaseWorkflowEvent: wms3.BaseWorkflowEvent{
					ServiceName:       "test_service",
					ContainerName:     "service-1",
					FullContainerName: "test_service-1",
					WorkflowName:      workflow,
				},
				StepID:    "12345",
				Name:      "test",
				Command:   "/bin/sh",
				Arguments: []string{"-c", "echo \"Hello World\""},
				Cwd:       "/",
				Shell:     "/bin/sh",
			}).Return()

			t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
				Action: services.WorkflowAction_RUN_COMMAND_ACTION,
				RunCommand: &services.WorkflowRunCommand{
					Command:          "/bin/sh",
					Arguments:        []string{"-c", "echo \"Hello World\""},
					WorkingDirectory: "/",
				},
			}).Return(nil)

			trigger, ok := args.Get(0).(wms3.StepTriggerLambda)
			t.True(ok)

			err := trigger()
			t.Nil(err)

			// progress
			t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
				Request: &services.RunWorkflowStreamRequest_StreamRequest{
					StreamRequest: &services.WorkflowStreamRequest{
						Result: services.WorkflowResult_RUN_COMMAND_RESULT,
						RunCommandResult: &services.WorkflowRunResult{
							Stdout:   "Hello World",
							Stderr:   "",
							ExitCode: &exitCode,
						},
					},
				},
			}, nil)

			t.mockEventManager.On("Publish", &wms3.WorkflowStepOutputEvent{
				BaseWorkflowEvent: wms3.BaseWorkflowEvent{
					ServiceName:       "test_service",
					ContainerName:     "service-1",
					FullContainerName: "test_service-1",
					WorkflowName:      workflow,
				},
				StepID: "12345",
				Stdout: "Hello World",
				Stderr: "",
			}).Return()

			progress, ok := args.Get(1).(wms3.StepProgressLambda)
			t.True(ok)

			exitCodePtr, err := progress()
			t.Nil(err)

			// completion
			t.mockEventManager.On("Publish", &wms3.WorkflowStepCompleteEvent{
				BaseWorkflowEvent: wms3.BaseWorkflowEvent{
					ServiceName:       "test_service",
					ContainerName:     "service-1",
					FullContainerName: "test_service-1",
					WorkflowName:      workflow,
				},
				StepID:    "12345",
				Command:   "/bin/sh",
				Arguments: []string{"-c", "echo \"Hello World\""},
				Cwd:       "/",
				Shell:     "/bin/sh",
				ExitCode:  uint8(exitCode), // nolint:gosec
			}).Return()

			complete, ok := args.Get(2).(wms3.StepCompleteLambda)
			t.True(ok)

			err = complete(*exitCodePtr)
			t.Nil(err)
		}).Return(nil)

		yield(step)
	})

	t.mockEventManager.On("Publish", &wms3.WorkflowCompleteEvent{
		BaseWorkflowEvent: wms3.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      workflow,
		},
		Successful: exitCode == 0,
	}).Return()

	t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
		Action:     services.WorkflowAction_COMPLETE_ACTION,
		RunCommand: nil,
	}).Return(nil)
}

func (t *WorkflowTestSuite) testServerRecvErrorFor(workflow commonworkflow.WorkflowName) {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockEventManager.On("Publish", &wms3.WorkflowStartedEvent{
		BaseWorkflowEvent: wms3.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      workflow,
		},
	}).Return()

	t.mockWorkflowFactory.On("Make", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(t.mockWorkflowOrchestrator, nil)

	t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
		Request: &services.RunWorkflowStreamRequest_RunRequest{
			RunRequest: &services.WorkflowRunRequest{
				WorkflowName: workflow.String(),
			},
		},
	}, nil).Once()

	t.mockWorkflowOrchestrator.On("StepIterator").Return(func(yield func(wms3.Step) bool) {
		step := &wms.MockStep{}
		step.On("GetID").Return("12345")
		step.On("GetName").Return("test")
		step.On("GetCommand").Return("/bin/sh")
		step.On("GetArguments").Return([]string{"-c", "echo \"Hello World\""})
		step.On("GetWorkingDirectory").Return("/")
		step.On("GetShell").Return("/bin/sh")

		step.On(
			"Trigger",
			mock.AnythingOfType("wms.StepTriggerLambda"),
			mock.AnythingOfType("wms.StepProgressLambda"),
			mock.AnythingOfType("wms.StepCompleteLambda"),
		).Run(func(args mock.Arguments) {
			// trigger
			t.mockEventManager.On("Publish", &wms3.WorkflowStepStartedEvent{
				BaseWorkflowEvent: wms3.BaseWorkflowEvent{
					ServiceName:       "test_service",
					ContainerName:     "service-1",
					FullContainerName: "test_service-1",
					WorkflowName:      workflow,
				},
				StepID:    "12345",
				Name:      "test",
				Command:   "/bin/sh",
				Arguments: []string{"-c", "echo \"Hello World\""},
				Cwd:       "/",
				Shell:     "/bin/sh",
			}).Return()

			t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
				Action: services.WorkflowAction_RUN_COMMAND_ACTION,
				RunCommand: &services.WorkflowRunCommand{
					Command:          "/bin/sh",
					Arguments:        []string{"-c", "echo \"Hello World\""},
					WorkingDirectory: "/",
				},
			}).Return(nil)

			trigger, ok := args.Get(0).(wms3.StepTriggerLambda)
			t.True(ok)

			err := trigger()
			t.Nil(err)

			// progress
			t.mockGrpcServer.On("Recv").Return(nil, errors.New("mock recv error"))

			progress, ok := args.Get(1).(wms3.StepProgressLambda)
			t.True(ok)

			_, err = progress()
			t.Error(err)
		}).Return(errors.New("mock recv error"))

		yield(step)
	})

	t.mockEventManager.On("Publish", &wms3.WorkflowErrorEvent{
		BaseWorkflowEvent: wms3.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      workflow,
		},
		Err: errors.New("mock recv error"),
	}).Return()
}

func (t *WorkflowTestSuite) testUnknownWorkflowResult(workflow commonworkflow.WorkflowName) {
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockEventManager.On("Publish", &wms3.WorkflowStartedEvent{
		BaseWorkflowEvent: wms3.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      workflow,
		},
	}).Return()

	t.mockWorkflowFactory.On("Make", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(t.mockWorkflowOrchestrator, nil)

	t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
		Request: &services.RunWorkflowStreamRequest_RunRequest{
			RunRequest: &services.WorkflowRunRequest{
				WorkflowName: workflow.String(),
			},
		},
	}, nil).Once()

	t.mockWorkflowOrchestrator.On("StepIterator").Return(func(yield func(wms3.Step) bool) {
		step := &wms.MockStep{}
		step.On("GetID").Return("12345")
		step.On("GetName").Return("test")
		step.On("GetCommand").Return("/bin/sh")
		step.On("GetArguments").Return([]string{"-c", "echo \"Hello World\""})
		step.On("GetWorkingDirectory").Return("/")
		step.On("GetShell").Return("/bin/sh")

		step.On(
			"Trigger",
			mock.AnythingOfType("wms.StepTriggerLambda"),
			mock.AnythingOfType("wms.StepProgressLambda"),
			mock.AnythingOfType("wms.StepCompleteLambda"),
		).Run(func(args mock.Arguments) {
			// trigger
			t.mockEventManager.On("Publish", &wms3.WorkflowStepStartedEvent{
				BaseWorkflowEvent: wms3.BaseWorkflowEvent{
					ServiceName:       "test_service",
					ContainerName:     "service-1",
					FullContainerName: "test_service-1",
					WorkflowName:      workflow,
				},
				StepID:    "12345",
				Name:      "test",
				Command:   "/bin/sh",
				Arguments: []string{"-c", "echo \"Hello World\""},
				Cwd:       "/",
				Shell:     "/bin/sh",
			}).Return()

			t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
				Action: services.WorkflowAction_RUN_COMMAND_ACTION,
				RunCommand: &services.WorkflowRunCommand{
					Command:          "/bin/sh",
					Arguments:        []string{"-c", "echo \"Hello World\""},
					WorkingDirectory: "/",
				},
			}).Return(nil)

			trigger, ok := args.Get(0).(wms3.StepTriggerLambda)
			t.True(ok)

			err := trigger()
			t.Nil(err)

			// progress
			var exitCode uint32

			t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
				Request: &services.RunWorkflowStreamRequest_StreamRequest{
					StreamRequest: &services.WorkflowStreamRequest{
						Result: -9999,
						RunCommandResult: &services.WorkflowRunResult{
							Stdout:   "Hello World",
							Stderr:   "",
							ExitCode: &exitCode,
						},
					},
				},
			}, nil)

			progress, ok := args.Get(1).(wms3.StepProgressLambda)
			t.True(ok)

			_, err = progress()
			t.Error(err)
		}).Return(errors.New("unknown result"))

		yield(step)
	})

	t.mockEventManager.On("Publish", &wms3.WorkflowErrorEvent{
		BaseWorkflowEvent: wms3.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      workflow,
		},
		Err: errors.New("unknown result"),
	}).Return()
}
