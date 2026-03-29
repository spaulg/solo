package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/app/wms"
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
