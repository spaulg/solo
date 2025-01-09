package wms

import (
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"iter"
)

type Orchestrator interface {
	StepIterator() iter.Seq[Step]
}

type DefaultOrchestrator struct {
	steps []project.WorkflowStep
}

func NewOrchestrator(
	steps []project.WorkflowStep,
) Orchestrator {
	return &DefaultOrchestrator{
		steps: steps,
	}
}

func (t *DefaultOrchestrator) StepIterator() iter.Seq[Step] {
	stepNumber := 0
	stepCount := len(t.steps)

	return func(yield func(Step) bool) {
		for stepNumber < stepCount {
			if !yield(NewStep(t.steps[stepNumber].Name, t.steps[stepNumber].Command, t.steps[stepNumber].Cwd)) {
				return
			}

			stepNumber++
		}
	}
}
