package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	context_types "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/app/wms"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/infra/container"
)

type WorkflowFactory struct{}

func NewWorkflowFactory() wms_types.WorkflowFactory {
	return &WorkflowFactory{}
}

func (t *WorkflowFactory) Make(
	soloCtx *context_types.CliContext,
	orchestrator container_types.Orchestrator,
	serviceName string,
	workflowName workflowcommon.WorkflowName,
) (wms_types.Workflow, error) {
	service := soloCtx.Project.Services().GetService(serviceName)
	var err error

	// Use project context to lookup the services working_directory option
	serviceWorkingDirectory := service.GetConfig().WorkingDir
	if serviceWorkingDirectory == "" {
		// Project does not define a working directory for the service
		// Use orchestrator to lookup the working directory from the service image
		serviceWorkingDirectory, err = orchestrator.ResolveImageWorkingDirectory(serviceName)
		if err != nil {
			return nil, err
		}
	}

	return NewWorkflow(
		soloCtx,
		serviceWorkingDirectory,
		service.GetServiceWorkflow(workflowName.String()),
	), nil
}
