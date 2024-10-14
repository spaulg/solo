package orchestrator

import (
	"github.com/spaulg/solo/cli/internal/pkg/solo/config"
)

type Orchestrator interface {
	Start(projectDirectory string, composeFile string) error
	Stop(projectDirectory string, composeFile string) error
	Destroy(projectDirectory string, composeFile string) error
	ExportComposeConfiguration(globalConfig *config.Config, projectPath string) ([]byte, error)
}

func BuildOrchestrator() Orchestrator {
	return &DockerOrchestrator{}
}
