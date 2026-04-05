package wms

import (
	"github.com/spaulg/solo/internal/pkg/host/app/context"
	wms_types "github.com/spaulg/solo/internal/pkg/shared/app/wms"
	workflowcommon "github.com/spaulg/solo/internal/pkg/shared/domain/wms"
)

type WorkflowGuardFactory struct {
	soloCtx *context.CliContext
}

func NewWorkflowGuardFactory(soloCtx *context.CliContext) *WorkflowGuardFactory {
	return &WorkflowGuardFactory{
		soloCtx: soloCtx,
	}
}

func (t *WorkflowGuardFactory) Build(workflowNames []workflowcommon.WorkflowName, containerNames []string) wms_types.WorkflowGuard {
	return NewWorkflowGuard(t.soloCtx, workflowNames, containerNames)
}
