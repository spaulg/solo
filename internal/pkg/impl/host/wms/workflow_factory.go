package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type WorkflowFactory struct{}

func NewWorkflowFactory() wms_types.WorkflowFactory {
	return &WorkflowFactory{}
}

func (t *WorkflowFactory) Make(
	project project_types.Project,
	serviceName string,
	workflowName workflowcommon.WorkflowName,
) wms_types.Orchestrator {
	return NewOrchestrator(project.Services().GetService(serviceName).GetServiceWorkflow(workflowName.String()))
}
