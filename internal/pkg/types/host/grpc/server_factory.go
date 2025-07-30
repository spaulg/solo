package grpc

import (
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type ServerFactory interface {
	Build(
		orchestrator container_types.Orchestrator,
		workflowExecutionTracker wms_types.WorkflowExecTracker,
		project project_types.Project,
		port int,
	) (Server, error)
}
