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

	hasServiceWorkflowRun        bool
	hasServiceWorkflowRunError   error
	hasFirstContainerWorkflowRun bool
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

	t.hasServiceWorkflowRun = false
	t.hasServiceWorkflowRunError = nil
	t.hasFirstContainerWorkflowRun = false
}

func (t *WorkflowRunnerTestSuite) testWorkflowExecFor(workflow commonworkflow.WorkflowName, exitCode uint8) {
	t.workflowSession.On("HasServiceWorkflowRun", "test_service").Return(t.hasServiceWorkflowRun, t.hasServiceWorkflowRunError)
	t.workflowSession.On("HasFirstContainerWorkflowRun").Return(t.hasFirstContainerWorkflowRun)
	t.workflowSession.On("MarkCompletion").Return(nil)

	t.workflowSession.On("GetServiceName").Return("test_service")
	t.workflowSession.On("GetContainerName").Return("service-1")
	t.workflowSession.On("GetFullContainerName").Return("test_service-1")
	t.workflowSession.On("GetWorkflowName").Return(workflow)

	if t.hasServiceWorkflowRun || t.hasFirstContainerWorkflowRun {
		t.mockEventManager.On("Publish", &wms_types.WorkflowSkippedEvent{
			BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
				ServiceName:       "test_service",
				ContainerName:     "service-1",
				FullContainerName: "test_service-1",
				WorkflowName:      workflow,
			},
			Successful: true,
		}).Return()

		return
	}

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

	t.mockEventManager.On("Publish", &wms_types.WorkflowCompleteEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      workflow,
		},
		Successful: exitCode == 0,
	}).Return()
}

func (t *WorkflowRunnerTestSuite) TestServiceWorkflowRun() {
	t.hasServiceWorkflowRun = true
	t.hasServiceWorkflowRunError = nil
	t.hasFirstContainerWorkflowRun = false

	t.testWorkflowExecFor(commonworkflow.FirstPreStartContainer, 0)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	err := workflowService.RunWorkflow(t.workflowSession)
	t.NoError(err)

	t.workflowSession.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestContainerWorkflowRun() {
	t.hasServiceWorkflowRun = false
	t.hasServiceWorkflowRunError = nil
	t.hasFirstContainerWorkflowRun = true

	t.testWorkflowExecFor(commonworkflow.FirstPreStartContainer, 0)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	err := workflowService.RunWorkflow(t.workflowSession)
	t.NoError(err)

	t.workflowSession.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) TestFirstPreStartZeroWorkflowStepTrigger() {
	t.testWorkflowExecFor(commonworkflow.FirstPreStartContainer, 0)

	workflowService := NewWorkflowRunner(
		t.soloCtx,
		t.mockEventManager,
		t.mockWorkflowFactory,
	)

	err := workflowService.RunWorkflow(t.workflowSession)
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

	err := workflowService.RunWorkflow(t.workflowSession)
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

	err := workflowService.RunWorkflow(t.workflowSession)
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

	err := workflowService.RunWorkflow(t.workflowSession)
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

	err := workflowService.RunWorkflow(t.workflowSession)
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

	err := workflowService.RunWorkflow(t.workflowSession)
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

	err := workflowService.RunWorkflow(t.workflowSession)
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

	err := workflowService.RunWorkflow(t.workflowSession)
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

	err := workflowService.RunWorkflow(t.workflowSession)
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

	err := workflowService.RunWorkflow(t.workflowSession)
	t.NoError(err)

	t.workflowSession.AssertExpectations(t.T())
	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}

func (t *WorkflowRunnerTestSuite) testServerRecvErrorFor(workflow commonworkflow.WorkflowName) {
	t.workflowSession.On("GetServiceName").Return("test_service")
	t.workflowSession.On("GetContainerName").Return("service-1")
	t.workflowSession.On("GetFullContainerName").Return("test_service-1")
	t.workflowSession.On("GetWorkflowName").Return(workflow)
	t.workflowSession.On("GetWorkingDirectory").Return("/", nil)

	t.workflowSession.On("HasServiceWorkflowRun", "test_service").Return(false, nil)
	t.workflowSession.On("HasFirstContainerWorkflowRun").Return(false)
	t.workflowSession.On("MarkCompletion").Return(nil)

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

	err := workflowService.RunWorkflow(t.workflowSession)
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

	err := workflowService.RunWorkflow(t.workflowSession)
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

	err := workflowService.RunWorkflow(t.workflowSession)
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

	err := workflowService.RunWorkflow(t.workflowSession)
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

	err := workflowService.RunWorkflow(t.workflowSession)
	t.Error(err)

	t.mockEventManager.AssertExpectations(t.T())
	t.mockWorkflowFactory.AssertExpectations(t.T())
	t.mockWorkflowOrchestrator.AssertExpectations(t.T())
}
