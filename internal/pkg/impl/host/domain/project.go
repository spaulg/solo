package domain

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

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	domain_project "github.com/spaulg/solo/internal/pkg/impl/host/domain/project"
	domain_compose "github.com/spaulg/solo/internal/pkg/impl/host/domain/project/compose"
	domain_types "github.com/spaulg/solo/internal/pkg/types/host/domain"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/domain/project"
	compose_types "github.com/spaulg/solo/internal/pkg/types/host/domain/project/compose"
)

const generatedComposeFileName = "docker-compose.yml"

type compose = types.Project

type Project struct {
	*compose

	projectStateDirectory string
	directory             string
	filePath              string
}

func NewProject(projectFilePath string, config *Config, profiles []string) (domain_types.Project, error) {
	paths, isProjectFile := findComposeFiles(projectFilePath, config)

	projectOptions, err := cli.NewProjectOptions(nil,
		WithComposeFiles(paths),
		cli.WithLoadOptions(func(option *loader.Options) {
			option.ResolvePaths = false              // Keep paths relative in case the user moves their project folder
			option.SkipInterpolation = isProjectFile // Disable interpolation on the project file
		}),
		cli.WithExtension(project_types.ServiceWorkflowExtensionName, domain_project.NewServiceWorkflows()),
		cli.WithExtension(project_types.ToolExtensionName, domain_project.NewTools()),
		cli.WithProfiles(profiles),
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
	project.loadToolExtensionDefaults()
	project.loadServiceExtensionDefaults()

	return project, nil
}

func (t *Project) GetCompose() *types.Project {
	return t.compose
}

func (t *Project) Profiles() []string {
	return t.compose.Profiles
}

func (t *Project) ReloadWithProfiles(profiles []string) error {
	compose, err := t.WithProfiles(profiles)
	if err != nil {
		return fmt.Errorf("error reloading project with profiles: %w", err)
	}

	t.compose = compose

	// Set default values in extensions
	t.loadToolExtensionDefaults()
	t.loadServiceExtensionDefaults()

	return nil
}

func (t *Project) ReloadWithAllProfilesEnabled() (domain_types.Project, error) {
	compose, err := t.WithProfiles([]string{"*"})
	if err != nil {
		return nil, fmt.Errorf("error loading project: %w", err)
	}

	project := &Project{
		projectStateDirectory: t.projectStateDirectory,
		directory:             t.directory,
		filePath:              t.filePath,
		compose:               compose,
	}

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

func (t *Project) GetGeneratedComposeFilePath() string {
	return path.Join(t.projectStateDirectory, generatedComposeFileName)
}

func (t *Project) GetMaxWorkflowTimeout(eventName string) time.Duration {
	maxTimeout := types.Duration(0)

	for _, serviceConfig := range t.compose.Services {
		serviceWorkflows := serviceConfig.Extensions[project_types.ServiceWorkflowExtensionName].(compose_types.ServiceWorkflows)

		if v, ok := serviceWorkflows[eventName]; ok && *v.Timeout > maxTimeout {
			maxTimeout = *v.Timeout
		}
	}

	return time.Duration(maxTimeout)
}

func (t *Project) Name() string {
	return t.compose.Name
}

func (t *Project) Services() compose_types.Services {
	return domain_compose.NewServices(t, t.compose)
}

func (t *Project) Tools() compose_types.Tools {
	if t, ok := t.Extensions[project_types.ToolExtensionName].(compose_types.Tools); ok {
		return t
	}

	return compose_types.Tools{}
}

func findComposeFiles(projectFilePath string, config *Config) ([]string, bool) {
	projectDirectory := filepath.Dir(projectFilePath)
	var configPaths []string

	// Look for generated compose yaml file in the state directory
	// If found, use this, if not, find normal compose and override
	candidates := findFiles([]string{"docker-compose.yml"}, projectDirectory+"/"+config.StateDirectoryName)
	if len(candidates) == 1 {
		return []string{candidates[0]}, false
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

		configPaths = append(configPaths, winner)

		overrides := findFiles(cli.DefaultOverrideFileNames, projectDirectory)
		if len(overrides) > 0 {
			if len(overrides) > 1 {
				// todo: fix use of unsupported logger
				logrus.Warnf("Found multiple override files with supported names: %s", strings.Join(overrides, ", "))
				logrus.Warnf("Using %s", overrides[0])
			}

			configPaths = append(configPaths, overrides[0])
		}
	}

	configPaths = append(configPaths, projectFilePath)

	return configPaths, true
}

func WithComposeFiles(configPaths []string) func(o *cli.ProjectOptions) error {
	return func(o *cli.ProjectOptions) error {
		o.ConfigPaths = append(o.ConfigPaths, configPaths...)
		return nil
	}
}

func (t *Project) loadToolExtensionDefaults() {
	if t.Extensions == nil {
		t.Extensions = make(types.Extensions)
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
			v = domain_project.NewServiceWorkflows()
			serviceConfig.Extensions[project_types.ServiceWorkflowExtensionName] = v
		}

		workflows := v.(compose_types.ServiceWorkflows)
		for _, workflowName := range workflowcommon.WorkflowNames {
			if _, ok := workflows[workflowName.String()]; !ok {
				workflows[workflowName.String()] = compose_types.ServiceWorkflowConfig{
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
