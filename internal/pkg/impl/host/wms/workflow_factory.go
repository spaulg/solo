package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	context_types "github.com/spaulg/solo/internal/pkg/impl/host/context"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
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
) (wms_types.Orchestrator, error) {
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

	return NewOrchestrator(
		soloCtx,
		serviceWorkingDirectory,
		service.GetServiceWorkflow(workflowName.String()),
	), nil
}
