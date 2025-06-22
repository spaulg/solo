package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type WorkflowGuardFactory struct {
	soloCtx *context.CliContext
}

func NewWorkflowGuardFactory(soloCtx *context.CliContext) wms_types.WorkflowGuardFactory {
	return &WorkflowGuardFactory{
		soloCtx: soloCtx,
	}
}

func (t *WorkflowGuardFactory) Build(workflowNames []workflowcommon.WorkflowName, containerNames []string) wms_types.WorkflowGuard {
	return NewWorkflowGuard(t.soloCtx, workflowNames, containerNames)
}
