package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	context_types "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/wms/wf"
)

type WorkflowFactory struct{}

func NewWorkflowFactory() *WorkflowFactory {
	return &WorkflowFactory{}
}

func (t *WorkflowFactory) Make(
	soloCtx *context_types.CliContext,
	serviceName string,
	serviceWorkingDirectory string,
	workflowName workflowcommon.WorkflowName,
) (wf.Definition, error) {
	serviceWorkflow := soloCtx.Project.Services().GetService(serviceName).GetServiceWorkflow(workflowName.String())
	return NewWorkflow(soloCtx, serviceWorkingDirectory, serviceWorkflow), nil
}
