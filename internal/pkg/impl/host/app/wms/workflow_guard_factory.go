package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/wms/wf"
)

type WorkflowGuardFactory struct {
	soloCtx *context.CliContext
}

func NewWorkflowGuardFactory(soloCtx *context.CliContext) *WorkflowGuardFactory {
	return &WorkflowGuardFactory{
		soloCtx: soloCtx,
	}
}

func (t *WorkflowGuardFactory) Build(workflowNames []workflowcommon.WorkflowName, containerNames []string) wf.Guard {
	return NewWorkflowGuard(t.soloCtx, workflowNames, containerNames)
}
