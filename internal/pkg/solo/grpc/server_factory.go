package grpc

import "github.com/spaulg/solo/internal/pkg/solo/project"

type ServerFactory interface {
	Build(project *project.Project) (Server, error)
}
