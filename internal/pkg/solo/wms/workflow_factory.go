package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/common/wms"
	"github.com/spaulg/solo/internal/pkg/solo/project"
)

type Factory interface {
	Make(project *project.Project, service string, workflowName workflowcommon.WorkflowName) Orchestrator
}

type DefaultFactory struct{}

func NewWorkflowFactory() Factory {
	return &DefaultFactory{}
}

func (t *DefaultFactory) Make(
	project *project.Project,
	serviceName string,
	workflowName workflowcommon.WorkflowName,
) Orchestrator {
	return NewOrchestrator(project.GetServiceWorkflow(serviceName, workflowName.String()))
}
