package service_definitions

import (
	"context"
	"errors"
	"log/slog"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/internal/pkg/impl/common/grpc/services"
	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	cli_context "github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/grpc/interceptors"
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/grpc"
	"github.com/spaulg/solo/test/mocks/host/events"
	"github.com/spaulg/solo/test/mocks/host/logging"
	"github.com/spaulg/solo/test/mocks/host/project"
	"github.com/spaulg/solo/test/mocks/host/wms"
)

type WorkflowTestSuite struct {
	suite.Suite

	soloCtx             *cli_context.CliContext
	mockProject         *project.MockProject
	mockLogHandler      *logging.MockHandler
	mockEventManager    *events.MockEventManager
	mockOrchestrator    *wms.MockOrchestrator
	mockWorkflowFactory *wms.MockWorkflowFactory
	mockGrpcServer      *grpc.MockBidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse]
}

func (t *WorkflowTestSuite) SetupTest() {
	t.mockEventManager = &events.MockEventManager{}
	t.mockOrchestrator = &wms.MockOrchestrator{}
	t.mockWorkflowFactory = &wms.MockWorkflowFactory{}
	t.mockProject = &project.MockProject{}
	t.mockGrpcServer = &grpc.MockBidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse]{}

	t.mockLogHandler = &logging.MockHandler{}
	t.mockLogHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)

	t.soloCtx = &cli_context.CliContext{
		Project: t.mockProject,
		Logger:  slog.New(t.mockLogHandler),
		Config: &config_types.Config{
			Entrypoint: config_types.Entrypoint{
				HostEntrypointPath: test.GetTestDataFilePath("entrypoint.sh"),
			},
			GrpcServerPort: 0,
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

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	err := workflowService.FirstPreStartWorkflowStream(t.mockGrpcServer)

	t.Error(err, "unauthorized")

	t.mockLogHandler.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestMissingContainerName() {
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Container name not found"
	})).Return(nil)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	err := workflowService.FirstPreStartWorkflowStream(t.mockGrpcServer)

	t.Error(err, "unauthorized")

	t.mockLogHandler.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestNilWorkflowFromFactory() {
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "test_service-1")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockEventManager.On("Publish", &wms_types.WorkflowStartedEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   "test_service",
			ContainerName: "test_service-1",
			WorkflowName:  commonworkflow.PreStart,
		},
	}).Return()

	t.mockWorkflowFactory.On("Make", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	t.mockEventManager.On("Publish", &wms_types.WorkflowCompleteEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   "test_service",
			ContainerName: "test_service-1",
			WorkflowName:  commonworkflow.PreStart,
		},
		Successful: true,
	}).Return()

	t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
		Action:     services.WorkflowAction_COMPLETE_ACTION,
		RunCommand: nil,
	}).Return(nil)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PreStartWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestFirstPreStartSkippedWorkflowStepTrigger() {
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	firstPreStartCompleteContextValueName := interceptors.FirstPreStartComplete(interceptors.FirstPreStartCompleteContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "test_service-1")
	ctx = context.WithValue(ctx, firstPreStartCompleteContextValueName, "true")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockEventManager.On("Publish", &wms_types.WorkflowSkippedEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   "test_service",
			ContainerName: "test_service-1",
			WorkflowName:  commonworkflow.FirstPreStart,
		},
		Successful: true,
	}).Return()

	t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
		Action:     services.WorkflowAction_COMPLETE_ACTION,
		RunCommand: nil,
	}).Return(nil)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.FirstPreStartWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestFirstPreStartRecvError() {
	t.testServerRecvErrorFor(commonworkflow.FirstPreStart)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.FirstPreStartWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStartRecvError() {
	t.testServerRecvErrorFor(commonworkflow.PreStart)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PreStartWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPostStartRecvError() {
	t.testServerRecvErrorFor(commonworkflow.PostStart)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PostStartWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStopRecvError() {
	t.testServerRecvErrorFor(commonworkflow.PreStop)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PreStopWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreDestroyRecvError() {
	t.testServerRecvErrorFor(commonworkflow.PreDestroy)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PreDestroyWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestFirstPreStartUnknownWorkflowResult() {
	t.testUnknownWorkflowResult(commonworkflow.FirstPreStart)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.FirstPreStartWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStartUnknownWorkflowResult() {
	t.testUnknownWorkflowResult(commonworkflow.PreStart)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PreStartWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPostStartUnknownWorkflowResult() {
	t.testUnknownWorkflowResult(commonworkflow.PostStart)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PostStartWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStopUnknownWorkflowResult() {
	t.testUnknownWorkflowResult(commonworkflow.PreStop)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PreStopWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreDestroyUnknownWorkflowResult() {
	t.testUnknownWorkflowResult(commonworkflow.PreDestroy)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PreDestroyWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestFirstPreStartZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.FirstPreStart, 0)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.FirstPreStartWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStartZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreStart, 0)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PreStartWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPostStartZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PostStart, 0)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PostStartWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStopZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreStop, 0)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PreStopWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreDestroyZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreDestroy, 0)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PreDestroyWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestFirstPreStartNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.FirstPreStart, 10)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.FirstPreStartWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStartNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreStart, 10)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PreStartWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPostStartNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PostStart, 10)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PostStartWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreStopNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreStop, 10)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PreStopWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) TestPreDestroyNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreDestroy, 10)

	workflowService := NewWorkflowService(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
	_ = workflowService.PreDestroyWorkflowStream(t.mockGrpcServer)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockOrchestrator.AssertExpectations(t.T())
	t.mockGrpcServer.AssertExpectations(t.T())
}

func (t *WorkflowTestSuite) testWorkflowExecFor(workflow commonworkflow.WorkflowName, exitCode uint32) {
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "test_service-1")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockEventManager.On("Publish", &wms_types.WorkflowStartedEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   "test_service",
			ContainerName: "test_service-1",
			WorkflowName:  workflow,
		},
	}).Return()

	t.mockWorkflowFactory.On("Make", mock.Anything, mock.Anything, mock.Anything).Return(t.mockOrchestrator)

	t.mockOrchestrator.On("StepIterator").Return(func(yield func(wms_types.Step) bool) {
		step := &wms.MockStep{}
		step.On("GetId").Return("12345")
		step.On("GetName").Return("test")
		step.On("GetCommand").Return("/bin/sh")
		step.On("GetArguments").Return([]string{"-c", "echo \"Hello World\""})
		step.On("GetWorkingDirectory").Return("/")

		step.On(
			"Trigger",
			mock.AnythingOfType("wms.StepTriggerLambda"),
			mock.AnythingOfType("wms.StepProgressLambda"),
			mock.AnythingOfType("wms.StepCompleteLambda"),
		).Run(func(args mock.Arguments) {
			// trigger
			t.mockEventManager.On("Publish", &wms_types.WorkflowStepStartedEvent{
				BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
					ServiceName:   "test_service",
					ContainerName: "test_service-1",
					WorkflowName:  workflow,
				},
				StepId:    "12345",
				Name:      "test",
				Command:   "/bin/sh",
				Arguments: []string{"-c", "echo \"Hello World\""},
				Cwd:       "/",
			}).Return()

			t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
				Action: services.WorkflowAction_RUN_COMMAND_ACTION,
				RunCommand: &services.WorkflowRunCommand{
					Command:          "/bin/sh",
					Arguments:        []string{"-c", "echo \"Hello World\""},
					WorkingDirectory: "/",
				},
			}).Return(nil)

			trigger := args.Get(0).(wms_types.StepTriggerLambda)
			err := trigger()
			t.Nil(err)

			// progress
			t.mockGrpcServer.On("Recv").Return(&services.WorkflowStreamRequest{
				Result: services.WorkflowResult_RUN_COMMAND_RESULT,
				RunCommandResult: &services.WorkflowRunResult{
					Stdout:   "Hello World",
					Stderr:   "",
					ExitCode: &exitCode,
				},
			}, nil)

			t.mockEventManager.On("Publish", &wms_types.WorkflowStepOutputEvent{
				BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
					ServiceName:   "test_service",
					ContainerName: "test_service-1",
					WorkflowName:  workflow,
				},
				StepId: "12345",
				Stdout: "Hello World",
				Stderr: "",
			}).Return()

			progress := args.Get(1).(wms_types.StepProgressLambda)
			exitCodePtr, err := progress()
			t.Nil(err)

			// completion
			t.mockEventManager.On("Publish", &wms_types.WorkflowStepCompleteEvent{
				BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
					ServiceName:   "test_service",
					ContainerName: "test_service-1",
					WorkflowName:  workflow,
				},
				StepId:    "12345",
				Command:   "/bin/sh",
				Arguments: []string{"-c", "echo \"Hello World\""},
				Cwd:       "/",
				ExitCode:  uint8(exitCode),
			}).Return()

			complete := args.Get(2).(wms_types.StepCompleteLambda)
			err = complete(*exitCodePtr)
			t.Nil(err)
		}).Return(nil)

		yield(step)
	})

	t.mockEventManager.On("Publish", &wms_types.WorkflowCompleteEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   "test_service",
			ContainerName: "test_service-1",
			WorkflowName:  workflow,
		},
		Successful: exitCode == 0,
	}).Return()

	t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
		Action:     services.WorkflowAction_COMPLETE_ACTION,
		RunCommand: nil,
	}).Return(nil)
}

func (t *WorkflowTestSuite) testServerRecvErrorFor(workflow commonworkflow.WorkflowName) {
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "test_service-1")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockEventManager.On("Publish", &wms_types.WorkflowStartedEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   "test_service",
			ContainerName: "test_service-1",
			WorkflowName:  workflow,
		},
	}).Return()

	t.mockWorkflowFactory.On("Make", mock.Anything, mock.Anything, mock.Anything).Return(t.mockOrchestrator)

	t.mockOrchestrator.On("StepIterator").Return(func(yield func(wms_types.Step) bool) {
		step := &wms.MockStep{}
		step.On("GetId").Return("12345")
		step.On("GetName").Return("test")
		step.On("GetCommand").Return("/bin/sh")
		step.On("GetArguments").Return([]string{"-c", "echo \"Hello World\""})
		step.On("GetWorkingDirectory").Return("/")

		step.On(
			"Trigger",
			mock.AnythingOfType("wms.StepTriggerLambda"),
			mock.AnythingOfType("wms.StepProgressLambda"),
			mock.AnythingOfType("wms.StepCompleteLambda"),
		).Run(func(args mock.Arguments) {
			// trigger
			t.mockEventManager.On("Publish", &wms_types.WorkflowStepStartedEvent{
				BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
					ServiceName:   "test_service",
					ContainerName: "test_service-1",
					WorkflowName:  workflow,
				},
				StepId:    "12345",
				Name:      "test",
				Command:   "/bin/sh",
				Arguments: []string{"-c", "echo \"Hello World\""},
				Cwd:       "/",
			}).Return()

			t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
				Action: services.WorkflowAction_RUN_COMMAND_ACTION,
				RunCommand: &services.WorkflowRunCommand{
					Command:          "/bin/sh",
					Arguments:        []string{"-c", "echo \"Hello World\""},
					WorkingDirectory: "/",
				},
			}).Return(nil)

			trigger := args.Get(0).(wms_types.StepTriggerLambda)
			err := trigger()
			t.Nil(err)

			// progress
			t.mockGrpcServer.On("Recv").Return(nil, errors.New("mock recv error"))

			progress := args.Get(1).(wms_types.StepProgressLambda)
			_, err = progress()
			t.Error(err)
		}).Return(errors.New("mock recv error"))

		yield(step)
	})

	t.mockEventManager.On("Publish", &wms_types.WorkflowErrorEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   "test_service",
			ContainerName: "test_service-1",
			WorkflowName:  workflow,
		},
		Err: errors.New("mock recv error"),
	}).Return()
}

func (t *WorkflowTestSuite) testUnknownWorkflowResult(workflow commonworkflow.WorkflowName) {
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	ctx = context.WithValue(ctx, containerNameContextValueName, "test_service-1")

	t.mockGrpcServer.On("Context").Return(ctx)

	t.mockEventManager.On("Publish", &wms_types.WorkflowStartedEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   "test_service",
			ContainerName: "test_service-1",
			WorkflowName:  workflow,
		},
	}).Return()

	t.mockWorkflowFactory.On("Make", mock.Anything, mock.Anything, mock.Anything).Return(t.mockOrchestrator)

	t.mockOrchestrator.On("StepIterator").Return(func(yield func(wms_types.Step) bool) {
		step := &wms.MockStep{}
		step.On("GetId").Return("12345")
		step.On("GetName").Return("test")
		step.On("GetCommand").Return("/bin/sh")
		step.On("GetArguments").Return([]string{"-c", "echo \"Hello World\""})
		step.On("GetWorkingDirectory").Return("/")

		step.On(
			"Trigger",
			mock.AnythingOfType("wms.StepTriggerLambda"),
			mock.AnythingOfType("wms.StepProgressLambda"),
			mock.AnythingOfType("wms.StepCompleteLambda"),
		).Run(func(args mock.Arguments) {
			// trigger
			t.mockEventManager.On("Publish", &wms_types.WorkflowStepStartedEvent{
				BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
					ServiceName:   "test_service",
					ContainerName: "test_service-1",
					WorkflowName:  workflow,
				},
				StepId:    "12345",
				Name:      "test",
				Command:   "/bin/sh",
				Arguments: []string{"-c", "echo \"Hello World\""},
				Cwd:       "/",
			}).Return()

			t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
				Action: services.WorkflowAction_RUN_COMMAND_ACTION,
				RunCommand: &services.WorkflowRunCommand{
					Command:          "/bin/sh",
					Arguments:        []string{"-c", "echo \"Hello World\""},
					WorkingDirectory: "/",
				},
			}).Return(nil)

			trigger := args.Get(0).(wms_types.StepTriggerLambda)
			err := trigger()
			t.Nil(err)

			// progress
			var exitCode uint32 = 0
			t.mockGrpcServer.On("Recv").Return(&services.WorkflowStreamRequest{
				Result: -9999,
				RunCommandResult: &services.WorkflowRunResult{
					Stdout:   "Hello World",
					Stderr:   "",
					ExitCode: &exitCode,
				},
			}, nil)

			progress := args.Get(1).(wms_types.StepProgressLambda)
			_, err = progress()
			t.Error(err)
		}).Return(errors.New("unknown result"))

		yield(step)
	})

	t.mockEventManager.On("Publish", &wms_types.WorkflowErrorEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   "test_service",
			ContainerName: "test_service-1",
			WorkflowName:  workflow,
		},
		Err: errors.New("unknown result"),
	}).Return()
}
