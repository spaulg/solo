package wms

import (
	"iter"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type Orchestrator struct {
	steps []project_types.WorkflowStep
}

func NewOrchestrator(
	workflow project_types.ServiceWorkflowConfig,
) wms_types.Orchestrator {
	return &Orchestrator{
		steps: workflow.Steps,
	}
}

func (t *Orchestrator) StepIterator() iter.Seq[wms_types.Step] {
	stepNumber := 0
	stepCount := len(t.steps)

	return func(yield func(wms_types.Step) bool) {
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
