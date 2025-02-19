package wms

import (
	"github.com/oklog/ulid/v2"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"iter"
	"math/rand"
	"time"
)

type Orchestrator interface {
	StepIterator() iter.Seq[Step]
}

type DefaultOrchestrator struct {
	steps []project.WorkflowStep
}

func NewOrchestrator(
	workflow project.ServiceWorkflowConfig,
) Orchestrator {
	return &DefaultOrchestrator{
		steps: workflow.Steps,
	}
}

func (t *DefaultOrchestrator) StepIterator() iter.Seq[Step] {
	stepNumber := 0
	stepCount := len(t.steps)

	return func(yield func(Step) bool) {
		entropy := rand.New(rand.NewSource(time.Now().UnixNano()))

		for stepNumber < stepCount {
			id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)

			step := NewStep(id.String(), t.steps[stepNumber].Name, t.steps[stepNumber].Run, t.steps[stepNumber].Cwd)
			if !yield(step) {
				return
			}

			stepNumber++
		}
	}
}
