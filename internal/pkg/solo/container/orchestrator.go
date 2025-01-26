package container

import (
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/project"
)

type Orchestrator interface {
	Up() error
	Down() error
	Destroy() error
	Execute(serviceNames []string, command []string) error
	GetHostGatewayHostname() string
	ExportComposeConfiguration(config *config.Config, project *project.Project) ([]byte, error)
}
