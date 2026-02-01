package services

import (
	"github.com/spaulg/solo/internal/pkg/impl/common/grpc/services"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
)

type WorkflowFactory interface {
	Build(orchestrator container_types.Orchestrator) services.WorkflowServer
}
