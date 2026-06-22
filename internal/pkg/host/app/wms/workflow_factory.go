package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/host/app/wms/wf"
	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
)

type WorkflowFactory struct{}

func NewWorkflowFactory() *WorkflowFactory {
	return &WorkflowFactory{}
}

func (t *WorkflowFactory) Make(
	config *domain2.Config,
	project domain2.Project,
	serviceName string,
	serviceWorkingDirectory string,
	workflowName workflowcommon.WorkflowName,
) (wf.Definition, error) {
	serviceWorkflow := project.Services().GetService(serviceName).GetServiceWorkflow(workflowName.String())
	return NewWorkflow(config, serviceWorkingDirectory, serviceWorkflow), nil
}
