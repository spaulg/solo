package grpc

import "github.com/spaulg/solo/internal/pkg/solo/project"

type ServerFactory interface {
	Build(hostname string, port uint16, project *project.Project) (Server, error)
}
