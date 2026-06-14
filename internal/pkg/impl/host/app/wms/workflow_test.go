package wms

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain/compose"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain/config"
	"github.com/spaulg/solo/test"
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

	mockConfig    *domain.Config
	expectedSteps []expectedStep
	config        compose.ServiceWorkflowConfig
}

func (t *WorkflowTestSuite) SetupTest() {
	cwd := "/"

	t.mockConfig = &domain.Config{
		Entrypoint: config.EntrypointConfig{
			HostEntrypointPath: test.GetTestDataFilePath("entrypoint.sh"),
		},
		Workflow: config.WorkflowConfig{
			Grpc: config.GrpcConfig{
				ServerPort: 0,
			},
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

	t.config = compose.NewServiceWorkflowConfig(
		[]domain.WorkflowStep{
			compose.NewWorkflowStep("step1", "/bin/foo arg1 arg2", nil, &cwd),
			compose.NewWorkflowStep("step2", "/bin/bar arg3 arg4", nil, &cwd),
			compose.NewWorkflowStep("step3", "/bin/baz arg5 arg6", nil, &cwd),
		},
		nil,
		nil,
	)
}

func (t *WorkflowTestSuite) TestIteration() {
	orchestrator := NewWorkflow(t.mockConfig, "/", t.config)

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
	orchestrator := NewWorkflow(t.mockConfig, "/", t.config)

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
