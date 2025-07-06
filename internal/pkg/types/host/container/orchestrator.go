package container

import (
	"google.golang.org/grpc/metadata"

	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
)

type Orchestrator interface {
	ComposeUp() error
	ComposeStop() error
	ComposeDown() error
	Execute(containerName string, command []string) error
	GetHostGatewayHostname() string
	ServicesStatus() (*ServiceStatus, error)
	ExportComposeConfiguration(config *config_types.Config, project project_types.Project) ([]byte, error)
	ResolveContainerNameFromMetadata(md metadata.MD) (*string, error)
}
