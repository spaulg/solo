package container

import (
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
	"google.golang.org/grpc/metadata"
)

type Orchestrator interface {
	ComposeUp() error
	ComposeStop() error
	ComposeDown() error
	Execute(containerName string, command []string) error
	GetHostGatewayHostname() string
	ServicesStatus() (*ServiceStatus, error)
	ExportComposeConfiguration(config *config_types.Config, project project_types.Project) ([]byte, error)
	ResolveServiceNameFromContainerName(containerName string) (*string, error)
	ResolveContainerNameFromMetadata(md metadata.MD) (*string, error)
}
