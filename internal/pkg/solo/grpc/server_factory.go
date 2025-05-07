package grpc

import (
	"github.com/spaulg/solo/internal/pkg/solo/container"
	"github.com/spaulg/solo/internal/pkg/solo/project"
)

type ServerFactory interface {
	Build(orchestrator container.Orchestrator, project *project.Project, port uint16) (Server, error)
}
