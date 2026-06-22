package wms

import (
	"testing"

	"github.com/stretchr/testify/suite"

	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
	compose2 "github.com/spaulg/solo/internal/pkg/host/domain/compose"
	config2 "github.com/spaulg/solo/internal/pkg/host/domain/config"
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

	mockConfig    *domain2.Config
	expectedSteps []expectedStep
	config        compose2.ServiceWorkflowConfig
}

func (t *WorkflowTestSuite) SetupTest() {
	cwd := "/"

	t.mockConfig = &domain2.Config{
		Entrypoint: config2.EntrypointConfig{
			HostEntrypointPath: test.GetTestDataFilePath("entrypoint.sh"),
		},
		Workflow: config2.WorkflowConfig{
			Grpc: config2.GrpcConfig{
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

	t.config = compose2.NewServiceWorkflowConfig(
		[]domain2.WorkflowStep{
			compose2.NewWorkflowStep("step1", "/bin/foo arg1 arg2", nil, &cwd),
			compose2.NewWorkflowStep("step2", "/bin/bar arg3 arg4", nil, &cwd),
			compose2.NewWorkflowStep("step3", "/bin/baz arg5 arg6", nil, &cwd),
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
