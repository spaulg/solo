package compose

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

	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
)

const generatedComposeFileName = "docker-compose.yml"

type composeProject = types.Project

type Project struct {
	*composeProject

	projectStateDirectory string
	directory             string
	filePath              string
}

func NewProject(projectFilePath string, config *domain2.Config, profiles []string) (*Project, error) {
	paths, isProjectFile := findComposeFiles(projectFilePath, config)

	projectOptions, err := cli.NewProjectOptions(nil,
		WithComposeFiles(paths),
		cli.WithLoadOptions(func(option *loader.Options) {
			option.ResolvePaths = false              // Keep paths relative in case the user moves their project folder
			option.SkipInterpolation = isProjectFile // Disable interpolation on the project file
		}),
		cli.WithExtension(ServiceWorkflowExtensionName, NewServiceWorkflows()),
		cli.WithExtension(ToolExtensionName, NewTools()),
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
		composeProject:        compose,
	}

	// Set default values in extensions
	project.loadToolExtensionDefaults()

	return project, nil
}

func (t *Project) GetCompose() *types.Project {
	return t.composeProject
}

func (t *Project) Profiles() []string {
	return t.composeProject.Profiles
}

func (t *Project) ReloadWithProfiles(profiles []string) error {
	composeProject, err := t.WithProfiles(profiles)
	if err != nil {
		return fmt.Errorf("error reloading project with profiles: %w", err)
	}

	t.composeProject = composeProject

	// Set default values in extensions
	t.loadToolExtensionDefaults()

	return nil
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
	maxTimeout := types.Duration(60 * time.Second)

	for _, serviceConfig := range t.composeProject.Services {
		if serviceConfig.Extensions == nil {
			continue
		}

		serviceWorkflows, ok := serviceConfig.Extensions[ServiceWorkflowExtensionName].(ServiceWorkflows)
		if !ok {
			continue
		}

		serviceWorkflow, ok := serviceWorkflows[eventName]
		if !ok {
			continue
		}

		workflowTimeout := serviceWorkflow.Timeout()
		if workflowTimeout > maxTimeout {
			maxTimeout = workflowTimeout
		}
	}

	return time.Duration(maxTimeout)
}

func (t *Project) Name() string {
	return t.composeProject.Name
}

func (t *Project) Services() domain2.Services {
	return NewServices(t, t.composeProject)
}

func (t *Project) Tools() domain2.Tools {
	if toolConfig, ok := t.Extensions[ToolExtensionName].(domain2.Tools); ok {
		return toolConfig
	}

	return domain2.Tools{}
}

func findComposeFiles(projectFilePath string, config *domain2.Config) ([]string, bool) {
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
