package container

import (
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"google.golang.org/grpc/metadata"
)

type Orchestrator interface {
	ComposeUp() error
	ComposeStop() error
	ComposeDown() error
	Execute(containerName string, command []string) error
	GetHostGatewayHostname() string
	ServicesStatus() (*ServiceStatus, error)
	ExportComposeConfiguration(config *config.Config, project *project.Project) ([]byte, error)
	ResolveServiceNameFromContainerName(containerName string) (*string, error)
	ResolveContainerNameFromMetadata(md metadata.MD) (*string, error)
}
