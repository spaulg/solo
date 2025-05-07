package container

import (
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"google.golang.org/grpc/metadata"
)

type Orchestrator interface {
	Up() error
	Down() error
	Destroy() error
	Execute(serviceNames []string, command []string) error
	GetHostGatewayHostname() string
	ServicesStatus() ([]string, []string, error)
	ExportComposeConfiguration(config *config.Config, project *project.Project) ([]byte, error)
	ResolveServiceNameFromContainerName(containerName string) (*string, error)
	ResolveContainerNameFromMetadata(md metadata.MD) (*string, error)
}
