package wms

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	cli_context "github.com/spaulg/solo/internal/pkg/impl/host/context"
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	compose_types "github.com/spaulg/solo/internal/pkg/types/host/project/compose"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/logging"
	"github.com/spaulg/solo/test/mocks/host/project"
)

func TestOrchestratorTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuite))
}

type expectedStep struct {
	Name      string
	Command   string
	Arguments []string
	Cwd       string
}

type WorkflowTestSuite struct {
	suite.Suite

	soloCtx        *cli_context.CliContext
	mockProject    *project.MockProject
	mockLogHandler *logging.MockHandler

	expectedSteps []expectedStep
	config        compose_types.ServiceWorkflowConfig
}

func (t *WorkflowTestSuite) SetupTest() {
	cwd := "/"

	t.mockProject = &project.MockProject{}

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

	t.expectedSteps = []expectedStep{
		{
			Name:      "step1",
			Command:   "/bin/foo",
			Arguments: []string{"arg1", "arg2"},
			Cwd:       cwd,
		},
		{
			Name:      "step2",
			Command:   "/bin/bar",
			Arguments: []string{"arg3", "arg4"},
			Cwd:       cwd,
		},
		{
			Name:      "step3",
			Command:   "/bin/baz",
			Arguments: []string{"arg5", "arg6"},
			Cwd:       cwd,
		},
	}

	t.config = compose_types.ServiceWorkflowConfig{
		Steps: []compose_types.WorkflowStep{
			{
				Name:             "step1",
				Run:              "/bin/foo arg1 arg2",
				WorkingDirectory: &cwd,
			},
			{
				Name:             "step2",
				Run:              "/bin/bar arg3 arg4",
				WorkingDirectory: &cwd,
			},
			{
				Name:             "step3",
				Run:              "/bin/baz arg5 arg6",
				WorkingDirectory: &cwd,
			},
		},
	}
}

func (t *WorkflowTestSuite) TestIteration() {
	orchestrator := NewWorkflow(t.soloCtx, "/", t.config)

	counter := 0
	for step := range orchestrator.StepIterator() {
		t.Equal(t.expectedSteps[counter].Name, step.GetName())
		t.Equal(t.expectedSteps[counter].Command, step.GetCommand())
		t.Equal(t.expectedSteps[counter].Arguments, step.GetArguments())
		t.Equal(t.expectedSteps[counter].Cwd, step.GetWorkingDirectory())

		counter++
	}
}

func (t *WorkflowTestSuite) TestIterationWithEarlyBreak() {
	orchestrator := NewWorkflow(t.soloCtx, "/", t.config)

	counter := 0
	for step := range orchestrator.StepIterator() {
		if counter == 1 {
			break
		}

		_ = step
		counter++
	}

	t.Equal(1, counter)
}
