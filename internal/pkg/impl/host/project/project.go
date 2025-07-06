package project

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/sirupsen/logrus"
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
)

const generatedComposeFileName = "docker-compose.yml"

type compose = types.Project

type Project struct {
	*compose

	projectStateDirectory string
	directory             string
	filePath              string
}

func NewProject(projectFilePath string, config *config_types.Config) (project_types.Project, error) {
	projectOptions, err := cli.NewProjectOptions(nil,
		WithComposeFiles(projectFilePath, config),
		cli.WithLoadOptions(func(option *loader.Options) {
			option.ResolvePaths = false     // Keep paths relative in case the user moves their project folder
			option.SkipInterpolation = true // Disable interpolation to avoid issues with environment variables in the project file
		}),
		cli.WithExtension(project_types.ServiceWorkflowExtensionName, NewServiceWorkflows()),
	)

	if err != nil {
		return nil, fmt.Errorf("error building project options: %w", err)
	}

	compose, err := projectOptions.LoadProject(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error loading project: %w", err)
	}

	projectDirectory := filepath.Dir(projectFilePath)
	project := &Project{
		projectStateDirectory: path.Join(projectDirectory, config.StateDirectoryName),
		directory:             projectDirectory,
		filePath:              projectFilePath,
		compose:               compose,
	}

	// Set default values in extensions
	project.loadServiceExtensionDefaults()

	return project, nil
}

func (t *Project) ResolveStateDirectory(relativePath string) string {
	return path.Join(t.projectStateDirectory, relativePath)
}

func (t *Project) GetAllServicesStateDirectory() string {
	return path.Join(t.projectStateDirectory, "services_all")
}

func (t *Project) GetServiceStateDirectoryRoot() string {
	return path.Join(t.projectStateDirectory, "services")
}

func (t *Project) GetServiceStateDirectory(serviceName string) string {
	return path.Join(t.GetServiceStateDirectoryRoot(), serviceName)
}

func (t *Project) GetServiceLogDirectory(serviceName string) string {
	return path.Join(t.GetServiceStateDirectoryRoot(), serviceName, "logs")
}

func (t *Project) GetServiceMountDirectory(serviceName string) string {
	return path.Join(t.GetServiceStateDirectoryRoot(), serviceName, "mount")
}

func (t *Project) GetStateDirectoryRoot() string {
	return t.projectStateDirectory
}

func (t *Project) GetDirectory() string {
	return t.directory
}

func (t *Project) GetFilePath() string {
	return t.filePath
}

func (t *Project) GetServiceWorkflow(serviceName string, eventName string) project_types.ServiceWorkflowConfig {
	serviceWorkflows := t.compose.Services[serviceName].Extensions[project_types.ServiceWorkflowExtensionName].(project_types.ServiceWorkflows)
	return serviceWorkflows[eventName]
}

func (t *Project) GetGeneratedComposeFilePath() string {
	return path.Join(t.projectStateDirectory, generatedComposeFileName)
}

func (t *Project) GetMaxWorkflowTimeout(eventName string) time.Duration {
	maxTimeout := types.Duration(0)

	for _, serviceConfig := range t.compose.Services {
		serviceWorkflows := serviceConfig.Extensions[project_types.ServiceWorkflowExtensionName].(project_types.ServiceWorkflows)

		if v, ok := serviceWorkflows[eventName]; ok && *v.Timeout > maxTimeout {
			maxTimeout = *v.Timeout
		}
	}

	return time.Duration(maxTimeout)
}

func (t *Project) Name() string {
	return t.compose.Name
}

func (t *Project) MarshalYAML() ([]byte, error) {
	return t.compose.MarshalYAML()
}

func (t *Project) Services() types.Services {
	return t.compose.Services
}

func (t *Project) ServiceNames() []string {
	return t.compose.ServiceNames()
}

func (t *Project) ContainerNames(serviceNames []string) ([]string, error) {
	var containerNames []string

	if err := t.ForEachService(serviceNames, func(name string, service *types.ServiceConfig) error {
		replicas := 1

		if service.Deploy != nil && service.Deploy.Replicas != nil {
			replicas = *service.Deploy.Replicas
		}

		if len(service.ContainerName) > 0 && replicas == 1 {
			// single container with a name defined by the container_name option
			containerNames = append(containerNames, service.ContainerName)
		} else {
			// one or more containers defined by the format {project}-{service}-{number}
			// consider moving this format to the orchestrator
			for i := 1; i <= replicas; i++ {
				containerName := fmt.Sprintf("%s-%s-%d", t.Name(), name, i)
				containerNames = append(containerNames, containerName)
			}
		}

		return nil
	}, types.IncludeDependencies); err != nil {
		return nil, err
	}

	return containerNames, nil
}

func WithComposeFiles(projectFilePath string, config *config_types.Config) func(o *cli.ProjectOptions) error {
	return func(o *cli.ProjectOptions) error {
		projectDirectory := filepath.Dir(projectFilePath)

		// Look for generated compose yaml file in the state directory
		// If found, use this, if not, find normal compose and override
		candidates := findFiles([]string{"docker-compose.yml"}, projectDirectory+"/"+config.StateDirectoryName)
		if len(candidates) == 1 {
			o.ConfigPaths = append(o.ConfigPaths, candidates[0])
			return nil
		}

		// Look for compose files in the project directory
		candidates = findFiles(cli.DefaultFileNames, projectDirectory)

		if len(candidates) > 0 {
			winner := candidates[0]
			if len(candidates) > 1 {
				// todo: fix use of unsupported logger
				logrus.Warnf("Found multiple config files with supported names: %s", strings.Join(candidates, ", "))
				logrus.Warnf("Using %s", winner)
			}

			o.ConfigPaths = append(o.ConfigPaths, winner)

			overrides := findFiles(cli.DefaultOverrideFileNames, projectDirectory)
			if len(overrides) > 0 {
				if len(overrides) > 1 {
					// todo: fix use of unsupported logger
					logrus.Warnf("Found multiple override files with supported names: %s", strings.Join(overrides, ", "))
					logrus.Warnf("Using %s", overrides[0])
				}

				o.ConfigPaths = append(o.ConfigPaths, overrides[0])
			}
		}

		o.ConfigPaths = append(o.ConfigPaths, projectFilePath)

		return nil
	}
}

func (t *Project) loadServiceExtensionDefaults() {
	defaultDuration := types.Duration(60 * time.Second)

	for serviceName, serviceConfig := range t.compose.Services {
		if serviceConfig.Extensions == nil {
			serviceConfig.Extensions = make(types.Extensions)
		}

		v, ok := serviceConfig.Extensions[project_types.ServiceWorkflowExtensionName]
		if !ok {
			v = NewServiceWorkflows()
			serviceConfig.Extensions[project_types.ServiceWorkflowExtensionName] = v
		}

		workflows := v.(project_types.ServiceWorkflows)
		for _, workflowName := range workflowcommon.WorkflowNames {
			if _, ok := workflows[workflowName.String()]; !ok {
				workflows[workflowName.String()] = project_types.ServiceWorkflowConfig{
					Timeout: &defaultDuration,
				}
			} else if workflows[workflowName.String()].Timeout == nil {
				workflowConfig := workflows[workflowName.String()]
				workflowConfig.Timeout = &defaultDuration

				workflows[workflowName.String()] = workflowConfig
			}
		}

		t.compose.Services[serviceName] = serviceConfig
	}
}

func findFiles(names []string, findDirectory string) []string {
	var candidates []string

	for _, n := range names {
		f := filepath.Join(findDirectory, n)
		if _, err := os.Stat(f); err == nil {
			candidates = append(candidates, f)
		}
	}

	return candidates
}
