package grpc

import (
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/app/wms"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/domain/project"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/infra/container"
)

type ServerFactory interface {
	Build(
		orchestrator container_types.Orchestrator,
		workflowExecutionTracker wms_types.WorkflowExecTracker,
		project project_types.Project,
		port int,
	) (Server, error)
}
