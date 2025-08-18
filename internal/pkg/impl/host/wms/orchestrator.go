package wms

import (
	"iter"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"

	compose_types "github.com/spaulg/solo/internal/pkg/types/host/project/compose"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type Orchestrator struct {
	serviceWorkingDirectory string
	steps                   []compose_types.WorkflowStep
}

func NewOrchestrator(
	serviceWorkingDirectory string,
	workflow compose_types.ServiceWorkflowConfig,
) wms_types.Orchestrator {
	return &Orchestrator{
		serviceWorkingDirectory: serviceWorkingDirectory,
		steps:                   workflow.Steps,
	}
}

func (t *Orchestrator) StepIterator() iter.Seq[wms_types.Step] {
	stepNumber := 0
	stepCount := len(t.steps)

	return func(yield func(wms_types.Step) bool) {
		entropy := rand.New(rand.NewSource(time.Now().UnixNano()))

		for stepNumber < stepCount {
			id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)

			workingDirectory := t.serviceWorkingDirectory
			if t.steps[stepNumber].WorkingDirectory != nil {
				workingDirectory = *t.steps[stepNumber].WorkingDirectory
			}

			step := NewStep(id.String(), t.steps[stepNumber].Name, t.steps[stepNumber].Run, workingDirectory)
			if !yield(step) {
				return
			}

			stepNumber++
		}
	}
}
