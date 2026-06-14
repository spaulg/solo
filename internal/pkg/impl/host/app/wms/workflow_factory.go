package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/wms/wf"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
)

type WorkflowFactory struct{}

func NewWorkflowFactory() *WorkflowFactory {
	return &WorkflowFactory{}
}

func (t *WorkflowFactory) Make(
	config *domain.Config,
	project domain.Project,
	serviceName string,
	serviceWorkingDirectory string,
	workflowName workflowcommon.WorkflowName,
) (wf.Definition, error) {
	serviceWorkflow := project.Services().GetService(serviceName).GetServiceWorkflow(workflowName.String())
	return NewWorkflow(config, serviceWorkingDirectory, serviceWorkflow), nil
}
