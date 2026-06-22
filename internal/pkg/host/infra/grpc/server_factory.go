package grpc

import (
	"github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/internal/pkg/host/infra/grpc/service_definitions"
)

type ServerFactory interface {
	Build(
		orchestrator ContainerResolver,
		workflowExecutionTracker service_definitions.WorkflowExecTracker,
		project domain.Project,
		port int,
	) (Server, error)
}
