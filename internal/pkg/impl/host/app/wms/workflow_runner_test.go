package wms

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	cli_context "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/wms/wf"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	domain_config "github.com/spaulg/solo/internal/pkg/impl/host/domain/config"
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

func (t *WorkflowRunnerTestSuite) testWorkflowExecFor(wfName commonworkflow.WorkflowName, exitCode uint8) {
	t.workflowSession.On("HasServiceWorkflowRun", "test_service").Return(t.hasServiceWorkflowRun, t.hasServiceWorkflowRunError)
	t.workflowSession.On("HasFirstContainerWorkflowRun").Return(t.hasFirstContainerWorkflowRun)
	t.workflowSession.On("MarkCompletion").Return(nil)

	t.workflowSession.On("GetServiceName").Return("test_service")
	t.workflowSession.On("GetContainerName").Return("service-1")
	t.workflowSession.On("GetFullContainerName").Return("test_service-1")
	t.workflowSession.On("GetWorkflowName").Return(wfName)

	if t.hasServiceWorkflowRun || t.hasFirstContainerWorkflowRun {
		t.mockEventManager.On("Publish", &wf.SkippedEvent{
			BaseWorkflowEvent: wf.BaseWorkflowEvent{
				ServiceName:       "test_service",
				ContainerName:     "service-1",
				FullContainerName: "test_service-1",
				WorkflowName:      wfName,
			},
			Successful: true,
		}).Return()

		return
	}

	t.workflowSession.On("GetWorkingDirectory").Return("/", nil)

	t.mockEventManager.On("Publish", &wf.StartedEvent{
		BaseWorkflowEvent: wf.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      wfName,
		},
	}).Return()

	t.mockWorkflowFactory.On(
		"Make",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(t.mockWorkflowOrchestrator, nil)

	t.mockWorkflowOrchestrator.On("StepIterator").Return(func(yield func(wf.Step) bool) {
		step := &wms.MockStep{}
		step.On("GetID").Return("12345")
		step.On("GetName").Return("test")
		step.On("GetCommand").Return("/bin/sh")
		step.On("GetArguments").Return([]string{"-c", "echo \"Hello World\""})
		step.On("GetWorkingDirectory").Return("/")
		step.On("GetShell").Return("/bin/sh")

		step.On(
			"Trigger",
			mock.AnythingOfType("StepTriggerFunc"),
			mock.AnythingOfType("StepProgressFunc"),
			mock.AnythingOfType("StepCompleteFunc"),
		).Run(func(args mock.Arguments) {
			// trigger
			t.mockEventManager.On("Publish", &wf.StepStartedEvent{
				BaseWorkflowEvent: wf.BaseWorkflowEvent{
					ServiceName:       "test_service",
					ContainerName:     "service-1",
					FullContainerName: "test_service-1",
					WorkflowName:      wfName,
				},
				StepID:    "12345",
				Name:      "test",
				Command:   "/bin/sh",
				Arguments: []string{"-c", "echo \"Hello World\""},
				Cwd:       "/",
				Shell:     "/bin/sh",
			}).Return()

			t.workflowSession.On("RunCommand", &wf.RunCommandRequest{
				Command:          "/bin/sh",
				Arguments:        []string{"-c", "echo \"Hello World\""},
				WorkingDirectory: "/",
			}).Return(nil)

			trigger, ok := args.Get(0).(wf.StepTriggerFunc)
			t.True(ok)

			err := trigger()
			t.Nil(err)

			// progress
			t.workflowSession.On("RecvCommandResponse").Return(&wf.CommandResponse{
				Stdout:   "Hello World",
				Stderr:   "",
				ExitCode: &exitCode,
			}, nil)

			t.mockEventManager.On("Publish", &wf.StepOutputEvent{
				BaseWorkflowEvent: wf.BaseWorkflowEvent{
					ServiceName:       "test_service",
					ContainerName:     "service-1",
					FullContainerName: "test_service-1",
					WorkflowName:      wfName,
				},
				StepID: "12345",
				Stdout: "Hello World",
				Stderr: "",
			}).Return()

			progress, ok := args.Get(1).(wf.StepProgressFunc)
			t.True(ok)

			exitCodePtr, err := progress()
			t.Nil(err)

			// completion
			t.mockEventManager.On("Publish", &wf.StepCompleteEvent{
				BaseWorkflowEvent: wf.BaseWorkflowEvent{
					ServiceName:       "test_service",
					ContainerName:     "service-1",
					FullContainerName: "test_service-1",
					WorkflowName:      wfName,
				},
				StepID:    "12345",
				Command:   "/bin/sh",
				Arguments: []string{"-c", "echo \"Hello World\""},
				Cwd:       "/",
				Shell:     "/bin/sh",
				ExitCode:  uint8(exitCode), // nolint:gosec
			}).Return()

			complete, ok := args.Get(2).(wf.StepCompleteFunc)
			t.True(ok)

			err = complete(*exitCodePtr)
			t.Nil(err)
		}).Return(nil)

		yield(step)
	})

	t.mockEventManager.On("Publish", &wf.CompleteEvent{
		BaseWorkflowEvent: wf.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      wfName,
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

func (t *WorkflowRunnerTestSuite) testServerRecvErrorFor(wfName commonworkflow.WorkflowName) {
	t.workflowSession.On("GetServiceName").Return("test_service")
	t.workflowSession.On("GetContainerName").Return("service-1")
	t.workflowSession.On("GetFullContainerName").Return("test_service-1")
	t.workflowSession.On("GetWorkflowName").Return(wfName)
	t.workflowSession.On("GetWorkingDirectory").Return("/", nil)

	t.workflowSession.On("HasServiceWorkflowRun", "test_service").Return(false, nil)
	t.workflowSession.On("HasFirstContainerWorkflowRun").Return(false)
	t.workflowSession.On("MarkCompletion").Return(nil)

	t.mockEventManager.On("Publish", &wf.StartedEvent{
		BaseWorkflowEvent: wf.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      wfName,
		},
	}).Return()

	t.mockWorkflowFactory.On(
		"Make",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(t.mockWorkflowOrchestrator, nil)

	t.mockWorkflowOrchestrator.On("StepIterator").Return(func(yield func(wf.Step) bool) {
		step := &wms.MockStep{}
		step.On("GetID").Return("12345")
		step.On("GetName").Return("test")
		step.On("GetCommand").Return("/bin/sh")
		step.On("GetArguments").Return([]string{"-c", "echo \"Hello World\""})
		step.On("GetWorkingDirectory").Return("/")
		step.On("GetShell").Return("/bin/sh")

		step.On(
			"Trigger",
			mock.AnythingOfType("StepTriggerFunc"),
			mock.AnythingOfType("StepProgressFunc"),
			mock.AnythingOfType("StepCompleteFunc"),
		).Run(func(args mock.Arguments) {
			// trigger
			t.mockEventManager.On("Publish", &wf.StepStartedEvent{
				BaseWorkflowEvent: wf.BaseWorkflowEvent{
					ServiceName:       "test_service",
					ContainerName:     "service-1",
					FullContainerName: "test_service-1",
					WorkflowName:      wfName,
				},
				StepID:    "12345",
				Name:      "test",
				Command:   "/bin/sh",
				Arguments: []string{"-c", "echo \"Hello World\""},
				Cwd:       "/",
				Shell:     "/bin/sh",
			}).Return()

			t.workflowSession.On("RunCommand", &wf.RunCommandRequest{
				Command:          "/bin/sh",
				Arguments:        []string{"-c", "echo \"Hello World\""},
				WorkingDirectory: "/",
			}).Return(nil)

			trigger, ok := args.Get(0).(wf.StepTriggerFunc)
			t.True(ok)

			err := trigger()
			t.Nil(err)

			// progress
			t.workflowSession.On("RecvCommandResponse").Return(nil, errors.New("mock recv error"))

			progress, ok := args.Get(1).(wf.StepProgressFunc)
			t.True(ok)

			_, err = progress()
			t.Error(err)
		}).Return(errors.New("mock recv error"))

		yield(step)
	})

	t.mockEventManager.On("Publish", &wf.ErrorEvent{
		BaseWorkflowEvent: wf.BaseWorkflowEvent{
			ServiceName:       "test_service",
			ContainerName:     "service-1",
			FullContainerName: "test_service-1",
			WorkflowName:      wfName,
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
