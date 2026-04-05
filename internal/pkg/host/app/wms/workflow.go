package wms

import (
	"crypto/rand"
	"iter"
	"time"

	"github.com/oklog/ulid/v2"

	context_types "github.com/spaulg/solo/internal/pkg/host/app/context"
	wms_types "github.com/spaulg/solo/internal/pkg/shared/app/wms"
	compose_types "github.com/spaulg/solo/internal/pkg/shared/domain/project/compose"
)

type Workflow struct {
	soloCtx                 *context_types.CliContext
	serviceWorkingDirectory string
	workflow                compose_types.ServiceWorkflowConfig
}

func NewWorkflow(
	soloCtx *context_types.CliContext,
	serviceWorkingDirectory string,
	workflow compose_types.ServiceWorkflowConfig,
) *Workflow {
	return &Workflow{
		soloCtx:                 soloCtx,
		serviceWorkingDirectory: serviceWorkingDirectory,
		workflow:                workflow,
	}
}

func (t *Workflow) StepIterator() iter.Seq[wms_types.Step] {
	stepNumber := 0
	stepCount := len(t.workflow.Steps)

	return func(yield func(wms_types.Step) bool) {
		for stepNumber < stepCount {
			id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)

			workingDirectory := t.serviceWorkingDirectory
			if t.workflow.Steps[stepNumber].WorkingDirectory != nil {
				workingDirectory = *t.workflow.Steps[stepNumber].WorkingDirectory
			}

			shell := t.soloCtx.Config.Workflow.DefaultStepShell
			if t.workflow.Steps[stepNumber].Shell != nil {
				shell = *t.workflow.Steps[stepNumber].Shell
			} else if t.workflow.Shell != nil {
				shell = *t.workflow.Shell
			}

			step := NewStep(
				id.String(),
				t.workflow.Steps[stepNumber].Name,
				t.workflow.Steps[stepNumber].Run,
				workingDirectory,
				shell,
			)

			if !yield(step) {
				return
			}

			stepNumber++
		}
	}
}
