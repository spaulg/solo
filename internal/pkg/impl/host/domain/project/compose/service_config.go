package compose

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"

	"github.com/spaulg/solo/internal/pkg/types/host/domain"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/domain/project"
	compose_types "github.com/spaulg/solo/internal/pkg/types/host/domain/project/compose"
)

type ServiceConfig struct {
	project       domain.Project
	serviceConfig types.ServiceConfig
}

func NewServiceConfig(project domain.Project, serviceConfig types.ServiceConfig) compose_types.ServiceConfig {
	return &ServiceConfig{
		project:       project,
		serviceConfig: serviceConfig,
	}
}

func (t *ServiceConfig) GetServiceWorkflow(eventName string) compose_types.ServiceWorkflowConfig {
	serviceWorkflows := t.serviceConfig.Extensions[project_types.ServiceWorkflowExtensionName].(compose_types.ServiceWorkflows)
	return serviceWorkflows[eventName]
}

func (t *ServiceConfig) GetConfig() types.ServiceConfig {
	return t.serviceConfig
}

func (t *ServiceConfig) ResolveContainerWorkingDirectory(cwd string) string {
	cwd = filepath.Clean(cwd) + string(filepath.Separator)

	projectDirectory := t.project.GetDirectory()
	stateDirectoryRoot := t.project.GetStateDirectoryRoot() + string(filepath.Separator)

	var preferredHostWorkingDirectory string
	var preferredContainerWorkingDirectory string

	for _, volume := range t.serviceConfig.Volumes {
		if volume.Type == "bind" {
			volumeSource := volume.Source

			if !filepath.IsAbs(volume.Source) {
				volumeSource = filepath.Join(projectDirectory, volumeSource)
			}

			volumeSource, err := filepath.Abs(volumeSource)
			if err != nil {
				continue // Skip directories that cannot be resolved
			}

			volumeSource += string(filepath.Separator)

			fileStat, err := os.Stat(volumeSource)
			if err != nil {
				continue // Skip directories that cannot be accessed
			}

			if !fileStat.IsDir() {
				continue // skip non-directory volumes
			}

			if strings.HasPrefix(volumeSource, stateDirectoryRoot) {
				continue // skip volumes that are in the state directory
			}

			if cwd != volumeSource && !strings.HasPrefix(cwd, volumeSource) {
				continue // skip volumes that are not a root of the current working directory
			}

			if len(volumeSource) > len(preferredHostWorkingDirectory) {
				preferredHostWorkingDirectory = volumeSource

				preferredContainerWorkingDirectory, err = filepath.Rel(volumeSource, cwd)
				if err != nil {
					continue // Skip if relative path cannot be calculated
				}

				preferredContainerWorkingDirectory = filepath.Join(volume.Target, preferredContainerWorkingDirectory)
			}
		}
	}

	return preferredContainerWorkingDirectory
}
