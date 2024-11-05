package orchestrator

import (
	"github.com/spaulg/solo/cli/internal/pkg/solo/config"
	"github.com/spaulg/solo/cli/internal/pkg/solo/project"
)

type Orchestrator interface {
	Up(projectDirectory string, composeFile string) error
	Down(projectDirectory string, composeFile string) error
	Destroy(projectDirectory string, composeFile string) error
	GetHostGatewayHostname() string
	ExportComposeConfiguration(config *config.Config, project *project.Project) ([]byte, error)
}

func OrchestratorFactory(config *config.Config) Orchestrator {
	return &DockerOrchestrator{}
}
