package grpc

import (
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/app/wms"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/domain"
)

type ServerFactory interface {
	Build(
		orchestrator ContainerResolver,
		workflowExecutionTracker wms_types.WorkflowExecTracker,
		project project_types.Project,
		port int,
	) (Server, error)
}
