package container

import (
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/project"
)

type Orchestrator interface {
	Up(projectDirectory string, composeFile string) error
	Down(projectDirectory string, composeFile string) error
	Destroy(projectDirectory string, composeFile string) error
	GetHostGatewayHostname() string
	ExportComposeConfiguration(config *config.Config, project *project.Project) ([]byte, error)
}

func OrchestratorFactory(soloCtx *context.SoloContext) Orchestrator {
	return &DockerOrchestrator{}
}
