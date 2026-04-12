package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	context_types "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/app/wms"
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
) (wms_types.Workflow, error) {
	serviceWorkflow := soloCtx.Project.Services().GetService(serviceName).GetServiceWorkflow(workflowName.String())
	return NewWorkflow(soloCtx, serviceWorkingDirectory, serviceWorkflow), nil
}
