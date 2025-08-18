package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type WorkflowFactory struct{}

func NewWorkflowFactory() wms_types.WorkflowFactory {
	return &WorkflowFactory{}
}

func (t *WorkflowFactory) Make(
	project project_types.Project,
	orchestrator container_types.Orchestrator,
	serviceName string,
	workflowName workflowcommon.WorkflowName,
) (wms_types.Orchestrator, error) {
	service := project.Services().GetService(serviceName)
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
		serviceWorkingDirectory,
		service.GetServiceWorkflow(workflowName.String()),
	), nil
}
