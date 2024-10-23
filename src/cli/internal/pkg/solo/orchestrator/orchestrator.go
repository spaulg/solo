package orchestrator

import (
	"github.com/spaulg/solo/cli/internal/pkg/solo/config"
)

type Orchestrator interface {
	Up(projectDirectory string, composeFile string) error
	Down(projectDirectory string, composeFile string) error
	Destroy(projectDirectory string, composeFile string) error
	ExportComposeConfiguration(globalConfig *config.Config, projectPath string) ([]byte, error)
}

func BuildOrchestrator() Orchestrator {
	return &DockerOrchestrator{}
}
