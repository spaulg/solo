package wms

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	cli_context "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	domain_config "github.com/spaulg/solo/internal/pkg/impl/host/domain/config"
	wms_shared "github.com/spaulg/solo/internal/pkg/impl/host/shared/wms"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/app/wms"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/app/events"
	"github.com/spaulg/solo/test/mocks/host/app/wms"
	"github.com/spaulg/solo/test/mocks/host/domain/project"
	"github.com/spaulg/solo/test/mocks/host/infra/grpc/service_definitions"
	"github.com/spaulg/solo/test/mocks/logging"
)

func TestWorkflowRunnerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowRunnerTestSuite))
}

type WorkflowRunnerTestSuite struct {
	suite.Suite

	soloCtx             *cli_context.CliContext
	mockProject         *project.MockProject
	mockLogHandler      *logging.MockHandler
	mockEventManager    *events.MockEventManager
	mockWorkflowFactory *wms.MockWorkflowFactory

	mockWorkflowOrchestrator *wms.MockWorkflow
	workflowSession          *service_definitions.MockWorkflowSession
}

func (t *WorkflowRunnerTestSuite) SetupTest() {
	t.mockProject = &project.MockProject{}
	t.mockWorkflowOrchestrator = &wms.MockWorkflow{}
	t.mockEventManager = &events.MockEventManager{}
	t.mockWorkflowFactory = &wms.MockWorkflowFactory{}
	t.workflowSession = &service_definitions.MockWorkflowSession{}
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

	NewWorkflowRunner(t.soloCtx, t.mockEventManager, t.mockWorkflowFactory)
}

//	func (t *WorkflowRunnerTestSuite) TestNilWorkflowFromFactory() {
//		serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
//		containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
//		fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
//
//		ctx := context.Background()
//		ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
//		ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
//		ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
//
//		t.mockGrpcServer.On("Context").Return(ctx)
//
//		t.mockEventManager.On("Publish", &wms_types.WorkflowStartedEvent{
//			BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
//				ServiceName:       "test_service",
//				ContainerName:     "service-1",
//				FullContainerName: "test_service-1",
//				WorkflowName:      commonworkflow.PreStartContainer,
//			},
//		}).Return()
//
//		t.mockWorkflowFactory.On("Make", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
//
//		t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
//			Request: &services.RunWorkflowStreamRequest_RunRequest{
//				RunRequest: &services.WorkflowRunRequest{
//					WorkflowName: commonworkflow.PreStartContainer.String(),
//				},
//			},
//		}, nil).Once()
//
//		t.mockEventManager.On("Publish", &wms_types.WorkflowCompleteEvent{
//			BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
//				ServiceName:       "test_service",
//				ContainerName:     "service-1",
//				FullContainerName: "test_service-1",
//				WorkflowName:      commonworkflow.PreStartContainer,
//			},
//			Successful: true,
//		}).Return()
//
//		t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
//			Action:     services.WorkflowAction_COMPLETE_ACTION,
//			RunCommand: nil,
//		}).Return(nil)
//
//		workflowService := NewWorkflowService(
//			t.soloCtx,
//			t.mockEventManager,
//			t.mockOrchestrator,
//			t.mockWorkflowFactory,
//			t.mockWorkflowExecTracker,
//			t.mockWorkflowRunner,
//		)
//
//		err := workflowService.RunWorkflowStream(t.mockGrpcServer)
//		t.NoError(err)
//
//		t.mockEventManager.AssertExpectations(t.T())
//		t.mockWorkflowFactory.AssertExpectations(t.T())
//		t.mockWorkflowOrchestrator.AssertExpectations(t.T())
//		t.mockGrpcServer.AssertExpectations(t.T())
//	}

func (t *WorkflowRunnerTestSuite) testWorkflowExecFor(workflow commonworkflow.WorkflowName, exitCode uint8) {
	t.workflowSession.On("GetServiceName").Return("test_service")
	t.workflowSession.On("GetContainerName").Return("service-1")
	t.workflowSession.On("GetFullContainerName").Return("test_service-1")
	t.workflowSession.On("GetWorkflowName").Return(workflow)
	t.workflowSession.On("GetWorkingDirectory").Return("/", nil)

	t.mockEventManager.On("Publish", &wms_types.WorkflowStartedEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      workflow,
		},
	}).Return()

	t.mockWorkflowFactory.On(
		"Make",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(t.mockWorkflowOrchestrator, nil)

	t.mockWorkflowOrchestrator.On("StepIterator").Return(func(yield func(wms_types.Step) bool) {
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
			t.mockEventManager.On("Publish", &wms_types.WorkflowStepStartedEvent{
				BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
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

			t.workflowSession.On("RunCommand", &wms_shared.RunCommandRequest{
				Command:          "/bin/sh",
				Arguments:        []string{"-c", "echo \"Hello World\""},
				WorkingDirectory: "/",
			}).Return(nil)

			trigger, ok := args.Get(0).(wms_types.StepTriggerLambda)
			t.True(ok)

			err := trigger()
			t.Nil(err)

			// progress
			t.workflowSession.On("RecvCommandResponse").Return(&wms_shared.CommandResponse{
				Stdout:   "Hello World",
				Stderr:   "",
				ExitCode: &exitCode,
			}, nil)

			t.mockEventManager.On("Publish", &wms_types.WorkflowStepOutputEvent{
				BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
					ServiceName:       "test_service",
					ContainerName:     "service-1",
					FullContainerName: "test_service-1",
					WorkflowName:      workflow,
				},
				StepID: "12345",
				Stdout: "Hello World",
				Stderr: "",
			}).Return()

			progress, ok := args.Get(1).(wms_types.StepProgressLambda)
			t.True(ok)

			exitCodePtr, err := progress()
			t.Nil(err)

			// completion
			t.mockEventManager.On("Publish", &wms_types.WorkflowStepCompleteEvent{
				BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
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

			complete, ok := args.Get(2).(wms_types.StepCompleteLambda)
			t.True(ok)

			err = complete(*exitCodePtr)
			t.Nil(err)
		}).Return(nil)

		yield(step)
	})

	//t.mockEventManager.On("Publish", &wms_types.WorkflowCompleteEvent{
	//	BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
	//		ServiceName:       "test_service",
	//		ContainerName:     "service-1",
	//		FullContainerName: "test_service-1",
	//		WorkflowName:      workflow,
	//	},
	//	Successful: exitCode == 0,
	//}).Return()
	//
	//t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
	//	Action:     services.WorkflowAction_COMPLETE_ACTION,
	//	RunCommand: nil,
	//}).Return(nil)
}

func (t *WorkflowRunnerTestSuite) TestFirstPreStartZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.FirstPreStartContainer, 0)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.True(success)
	t.NoError(err)

	t.workflowSession.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestPreStartZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreStartContainer, 0)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.True(success)
	t.NoError(err)

	t.workflowSession.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestPostStartZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PostStartContainer, 0)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.True(success)
	t.NoError(err)

	t.workflowSession.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestPreStopZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreStopContainer, 0)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.True(success)
	t.NoError(err)

	t.workflowSession.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestPreDestroyZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreDestroyContainer, 0)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.True(success)
	t.NoError(err)

	t.workflowSession.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestFirstPreStartNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.FirstPreStartContainer, 10)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.False(success)
	t.NoError(err)

	t.workflowSession.AssertExpectations(t.T())
	t.workflowSession.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestPreStartNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreStartContainer, 10)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.False(success)
	t.NoError(err)

	t.workflowSession.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestPostStartNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PostStartContainer, 10)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.False(success)
	t.NoError(err)

	t.workflowSession.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestPreStopNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreStopContainer, 10)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.False(success)
	t.NoError(err)

	t.workflowSession.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestPreDestroyNonZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.PreDestroyContainer, 10)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.False(success)
	t.NoError(err)

	t.workflowSession.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

//func (t *WorkflowRunnerTestSuite) testUnknownWorkflowResult(workflow commonworkflow.WorkflowName) {
//	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
//	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
//	fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
//
//	ctx := context.Background()
//	ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
//	ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
//	ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")
//
//	t.mockGrpcServer.On("Context").Return(ctx)
//
//	t.mockEventManager.On("Publish", &wms_types.WorkflowStartedEvent{
//		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
//			ServiceName:       "test_service",
//			ContainerName:     "service-1",
//			FullContainerName: "test_service-1",
//			WorkflowName:      workflow,
//		},
//	}).Return()
//
//	t.mockWorkflowFactory.On("Make", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(t.mockWorkflowOrchestrator, nil)
//
//	t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
//		Request: &services.RunWorkflowStreamRequest_RunRequest{
//			RunRequest: &services.WorkflowRunRequest{
//				WorkflowName: workflow.String(),
//			},
//		},
//	}, nil).Once()
//
//	t.mockWorkflowOrchestrator.On("StepIterator").Return(func(yield func(wms_types.Step) bool) {
//		step := &wms.MockStep{}
//		step.On("GetID").Return("12345")
//		step.On("GetName").Return("test")
//		step.On("GetCommand").Return("/bin/sh")
//		step.On("GetArguments").Return([]string{"-c", "echo \"Hello World\""})
//		step.On("GetWorkingDirectory").Return("/")
//		step.On("GetShell").Return("/bin/sh")
//
//		step.On(
//			"Trigger",
//			mock.AnythingOfType("wms.StepTriggerLambda"),
//			mock.AnythingOfType("wms.StepProgressLambda"),
//			mock.AnythingOfType("wms.StepCompleteLambda"),
//		).Run(func(args mock.Arguments) {
//			// trigger
//			t.mockEventManager.On("Publish", &wms_types.WorkflowStepStartedEvent{
//				BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
//					ServiceName:       "test_service",
//					ContainerName:     "service-1",
//					FullContainerName: "test_service-1",
//					WorkflowName:      workflow,
//				},
//				StepID:    "12345",
//				Name:      "test",
//				Command:   "/bin/sh",
//				Arguments: []string{"-c", "echo \"Hello World\""},
//				Cwd:       "/",
//				Shell:     "/bin/sh",
//			}).Return()
//
//			t.mockGrpcServer.On("Send", &services.WorkflowStreamResponse{
//				Action: services.WorkflowAction_RUN_COMMAND_ACTION,
//				RunCommand: &services.WorkflowRunCommand{
//					Command:          "/bin/sh",
//					Arguments:        []string{"-c", "echo \"Hello World\""},
//					WorkingDirectory: "/",
//				},
//			}).Return(nil)
//
//			trigger, ok := args.Get(0).(wms_types.StepTriggerLambda)
//			t.True(ok)
//
//			err := trigger()
//			t.Nil(err)
//
//			// progress
//			var exitCode uint32
//
//			t.mockGrpcServer.On("Recv").Return(&services.RunWorkflowStreamRequest{
//				Request: &services.RunWorkflowStreamRequest_StreamRequest{
//					StreamRequest: &services.WorkflowStreamRequest{
//						Result: -9999,
//						RunCommandResult: &services.WorkflowRunResult{
//							Stdout:   "Hello World",
//							Stderr:   "",
//							ExitCode: &exitCode,
//						},
//					},
//				},
//			}, nil)
//
//			progress, ok := args.Get(1).(wms_types.StepProgressLambda)
//			t.True(ok)
//
//			_, err = progress()
//			t.Error(err)
//		}).Return(errors.New("unknown result"))
//
//		yield(step)
//	})
//
//	t.mockEventManager.On("Publish", &wms_types.WorkflowErrorEvent{
//		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
//			ServiceName:       "test_service",
//			ContainerName:     "service-1",
//			FullContainerName: "test_service-1",
//			WorkflowName:      workflow,
//		},
//		Err: errors.New("unknown result"),
//	}).Return()
//}

func (t *WorkflowRunnerTestSuite) testServerRecvErrorFor(workflow commonworkflow.WorkflowName) {
	//serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	//containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	//fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
	//
	//ctx := context.Background()
	//ctx = context.WithValue(ctx, serviceNameContextValueName, "test_service")
	//ctx = context.WithValue(ctx, containerNameContextValueName, "service-1")
	//ctx = context.WithValue(ctx, fullContainerNameContextValueName, "test_service-1")

	t.workflowSession.On("GetServiceName").Return("test_service")
	t.workflowSession.On("GetContainerName").Return("service-1")
	t.workflowSession.On("GetFullContainerName").Return("test_service-1")
	t.workflowSession.On("GetWorkflowName").Return(workflow)
	t.workflowSession.On("GetWorkingDirectory").Return("/", nil)

	//t.mockGrpcServer.On("Context").Return(ctx)

	t.mockEventManager.On("Publish", &wms_types.WorkflowStartedEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      workflow,
		},
	}).Return()

	t.mockWorkflowFactory.On(
		"Make",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(t.mockWorkflowOrchestrator, nil)

	t.mockWorkflowOrchestrator.On("StepIterator").Return(func(yield func(wms_types.Step) bool) {
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
			t.mockEventManager.On("Publish", &wms_types.WorkflowStepStartedEvent{
				BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
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

			t.workflowSession.On("RunCommand", &wms_shared.RunCommandRequest{
				Command:          "/bin/sh",
				Arguments:        []string{"-c", "echo \"Hello World\""},
				WorkingDirectory: "/",
			}).Return(nil)

			trigger, ok := args.Get(0).(wms_types.StepTriggerLambda)
			t.True(ok)

			err := trigger()
			t.Nil(err)

			// progress
			t.workflowSession.On("RecvCommandResponse").Return(nil, errors.New("mock recv error"))

			progress, ok := args.Get(1).(wms_types.StepProgressLambda)
			t.True(ok)

			_, err = progress()
			t.Error(err)
		}).Return(errors.New("mock recv error"))

		yield(step)
	})

	t.mockEventManager.On("Publish", &wms_types.WorkflowErrorEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      workflow,
		},
		Err: errors.New("mock recv error"),
	}).Return()
}

func (t *WorkflowRunnerTestSuite) TestFirstPreStartRecvError() {
	t.testServerRecvErrorFor(commonworkflow.FirstPreStartContainer)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.False(success)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestPreStartRecvError() {
	t.testServerRecvErrorFor(commonworkflow.PreStartContainer)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.False(success)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestPostStartRecvError() {
	t.testServerRecvErrorFor(commonworkflow.PostStartContainer)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.False(success)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestPreStopRecvError() {
	t.testServerRecvErrorFor(commonworkflow.PreStopContainer)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.False(success)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestPreDestroyRecvError() {
	t.testServerRecvErrorFor(commonworkflow.PreDestroyContainer)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	success, err := workflowService.RunWorkflow(t.workflowSession)
	t.False(success)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

//func (t *WorkflowRunnerTestSuite) TestFirstPreStartUnknownWorkflowResult() {
//	t.testUnknownWorkflowResult(commonworkflow.FirstPreStartContainer)
//
//	workflowService := NewWorkflowService(
//		t.soloCtx,
//		t.mockEventManager,
//		t.mockOrchestrator,
//		t.mockWorkflowFactory,
//		t.mockWorkflowExecTracker,
//		t.mockWorkflowRunner,
//	)
//
//	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
//	t.Error(err)
//
//	t.mockEventManager.AssertExpectations(t.T())
//	t.mockWorkflowFactory.AssertExpectations(t.T())
//	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
//	t.mockGrpcServer.AssertExpectations(t.T())
//}
//
//func (t *WorkflowRunnerTestSuite) TestPreStartUnknownWorkflowResult() {
//	t.testUnknownWorkflowResult(commonworkflow.PreStartContainer)
//
//	workflowService := NewWorkflowService(
//		t.soloCtx,
//		t.mockEventManager,
//		t.mockOrchestrator,
//		t.mockWorkflowFactory,
//		t.mockWorkflowExecTracker,
//		t.mockWorkflowRunner,
//	)
//
//	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
//	t.Error(err)
//
//	t.mockEventManager.AssertExpectations(t.T())
//	t.mockWorkflowFactory.AssertExpectations(t.T())
//	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
//	t.mockGrpcServer.AssertExpectations(t.T())
//}
//
//func (t *WorkflowRunnerTestSuite) TestPostStartUnknownWorkflowResult() {
//	t.testUnknownWorkflowResult(commonworkflow.PostStartContainer)
//
//	workflowService := NewWorkflowService(
//		t.soloCtx,
//		t.mockEventManager,
//		t.mockOrchestrator,
//		t.mockWorkflowFactory,
//		t.mockWorkflowExecTracker,
//		t.mockWorkflowRunner,
//	)
//
//	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
//	t.Error(err)
//
//	t.mockEventManager.AssertExpectations(t.T())
//	t.mockWorkflowFactory.AssertExpectations(t.T())
//	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
//	t.mockGrpcServer.AssertExpectations(t.T())
//}
//
//func (t *WorkflowRunnerTestSuite) TestPreStopUnknownWorkflowResult() {
//	t.testUnknownWorkflowResult(commonworkflow.PreStopContainer)
//
//	workflowService := NewWorkflowService(
//		t.soloCtx,
//		t.mockEventManager,
//		t.mockOrchestrator,
//		t.mockWorkflowFactory,
//		t.mockWorkflowExecTracker,
//		t.mockWorkflowRunner,
//	)
//
//	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
//	t.Error(err)
//
//	t.mockEventManager.AssertExpectations(t.T())
//	t.mockWorkflowFactory.AssertExpectations(t.T())
//	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
//	t.mockGrpcServer.AssertExpectations(t.T())
//}
//
//func (t *WorkflowRunnerTestSuite) TestPreDestroyUnknownWorkflowResult() {
//	t.testUnknownWorkflowResult(commonworkflow.PreDestroyContainer)
//
//	workflowService := NewWorkflowService(
//		t.soloCtx,
//		t.mockEventManager,
//		t.mockOrchestrator,
//		t.mockWorkflowFactory,
//		t.mockWorkflowExecTracker,
//		t.mockWorkflowRunner,
//	)
//
//	err := workflowService.RunWorkflowStream(t.mockGrpcServer)
//	t.Error(err)
//
//	t.mockEventManager.AssertExpectations(t.T())
//	t.mockWorkflowFactory.AssertExpectations(t.T())
//	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
//	t.mockGrpcServer.AssertExpectations(t.T())
//}
//
