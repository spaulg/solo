package app

import (
	"github.com/spaulg/solo/internal/pkg/host/infra/grpc"
	wms_types "github.com/spaulg/solo/internal/pkg/shared/app/wms"
	project_types "github.com/spaulg/solo/internal/pkg/shared/domain"
	container_types "github.com/spaulg/solo/internal/pkg/shared/infra/container"
)

type ServerFactory interface {
	Build(
		orchestrator container_types.Orchestrator,
		workflowExecutionTracker wms_types.WorkflowExecTracker,
		project project_types.Project,
		port int,
	) (grpc.Server, error)
}
