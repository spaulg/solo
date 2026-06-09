package grpc

import (
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/grpc/service_definitions"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/domain"
)

type ServerFactory interface {
	Build(
		orchestrator ContainerResolver,
		workflowExecutionTracker service_definitions.WorkflowExecTracker,
		project project_types.Project,
		port int,
	) (Server, error)
}
