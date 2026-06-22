package wms

import (
	"crypto/rand"
	"iter"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/spaulg/solo/internal/pkg/host/app/wms/wf"
	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
)

type Workflow struct {
	config                  *domain2.Config
	serviceWorkingDirectory string
	workflow                domain2.ServiceWorkflowConfig
}

func NewWorkflow(
	config *domain2.Config,
	serviceWorkingDirectory string,
	workflow domain2.ServiceWorkflowConfig,
) *Workflow {
	return &Workflow{
		config:                  config,
		serviceWorkingDirectory: serviceWorkingDirectory,
		workflow:                workflow,
	}
}

func (t *Workflow) StepIterator() iter.Seq[wf.Step] {
	return func(yield func(wf.Step) bool) {
		for _, step := range t.workflow.Steps() {
			id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)

			workingDirectory := t.serviceWorkingDirectory
			if step.WorkingDirectory() != nil {
				workingDirectory = *step.WorkingDirectory()
			}

			shell := t.config.Workflow.DefaultStepShell
			if step.Shell() != nil {
				shell = *step.Shell()
			} else if t.workflow.Shell() != nil {
				shell = *t.workflow.Shell()
			}

			step := NewStep(
				id.String(),
				step.Name(),
				step.Run(),
				workingDirectory,
				shell,
			)

			if !yield(step) {
				return
			}
		}
	}
}
