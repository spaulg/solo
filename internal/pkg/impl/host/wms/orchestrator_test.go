package wms

import (
	"testing"

	"github.com/stretchr/testify/suite"

	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
)

func TestOrchestratorTestSuite(t *testing.T) {
	suite.Run(t, new(OrchestratorTestSuite))
}

type expectedStep struct {
	Name      string
	Command   string
	Arguments []string
	Cwd       string
}

type OrchestratorTestSuite struct {
	suite.Suite

	expectedSteps []expectedStep
	config        project_types.ServiceWorkflowConfig
}

func (t *OrchestratorTestSuite) SetupTest() {
	cwd := "/"

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

	t.config = project_types.ServiceWorkflowConfig{
		Steps: []project_types.WorkflowStep{
			{
				Name: "step1",
				Run:  "/bin/foo arg1 arg2",
				Cwd:  &cwd,
			},
			{
				Name: "step2",
				Run:  "/bin/bar arg3 arg4",
				Cwd:  &cwd,
			},
			{
				Name: "step3",
				Run:  "/bin/baz arg5 arg6",
				Cwd:  &cwd,
			},
		},
	}
}

func (t *OrchestratorTestSuite) TestIteration() {
	orchestrator := NewOrchestrator(t.config)

	counter := 0
	for step := range orchestrator.StepIterator() {
		t.Equal(t.expectedSteps[counter].Name, step.GetName())
		t.Equal(t.expectedSteps[counter].Command, step.GetCommand())
		t.Equal(t.expectedSteps[counter].Arguments, step.GetArguments())
		t.Equal(t.expectedSteps[counter].Cwd, step.GetWorkingDirectory())

		counter++
	}
}

func (t *OrchestratorTestSuite) TestIterationWithEarlyBreak() {
	orchestrator := NewOrchestrator(t.config)

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
