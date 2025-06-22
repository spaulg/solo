package grpc

import (
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
)

type ServerFactory interface {
	Build(orchestrator container_types.Orchestrator, project project_types.Project, port int) (Server, error)
}
