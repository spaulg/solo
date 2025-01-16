package project

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/sirupsen/logrus"
	workflowcommon "github.com/spaulg/solo/internal/pkg/common/wms"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type compose = types.Project

type Project struct {
	*compose

	projectStateDirectory string
	directory             string
	filePath              string
}

func NewProject(projectFilePath string) (*Project, error) {
	projectOptions, err := cli.NewProjectOptions(nil,
		WithComposeFiles(projectFilePath),
		cli.WithLoadOptions(func(option *loader.Options) {
			option.ResolvePaths = false // Keep paths relative in case the user moves their project folder
		}),
		cli.WithExtension(ServiceWorkflowExtensionName, NewServiceWorkflows()),
	)

	if err != nil {
		return nil, fmt.Errorf("error building project options: %v", err)
	}

	compose, err := projectOptions.LoadProject(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error loading project: %v", err)
	}

	projectDirectory := filepath.Dir(projectFilePath)
	project := &Project{
		projectStateDirectory: projectDirectory + "/.solo",
		directory:             projectDirectory,
		filePath:              projectFilePath,
		compose:               compose,
	}

	// Set default values in extensions
	project.loadServiceExtensionDefaults()

	return project, nil
}

func (t *Project) ResolveStateDirectory(relativePath string) string {
	return t.projectStateDirectory + "/" + relativePath
}

func (t *Project) GetAllServicesStateDirectory() string {
	return t.projectStateDirectory + "/services_all"
}

func (t *Project) GetServiceStateDirectoryRoot() string {
	return t.projectStateDirectory + "/services"
}

func (t *Project) GetServiceStateDirectory(serviceName string) string {
	return t.GetServiceStateDirectoryRoot() + "/" + serviceName
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

func (t *Project) GetServiceWorkflow(serviceName string, eventName string) ServiceWorkflowConfig {
	config := t.Services[serviceName].Extensions[ServiceWorkflowExtensionName].(ServiceWorkflows)
	return config[eventName]
}

func (t *Project) GetMaxWorkflowTimeout(eventName string) time.Duration {
	maxTimeout := types.Duration(0)

	for _, serviceConfig := range t.Services {
		config := serviceConfig.Extensions[ServiceWorkflowExtensionName].(ServiceWorkflows)

		if v, ok := config[eventName]; ok && *v.Timeout > maxTimeout {
			maxTimeout = *v.Timeout
		}
	}

	return time.Duration(maxTimeout)
}

func WithComposeFiles(projectFilePath string) func(o *cli.ProjectOptions) error {
	return func(o *cli.ProjectOptions) error {
		projectDirectory := filepath.Dir(projectFilePath)
		candidates := findFiles(cli.DefaultFileNames, projectDirectory)

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

	for serviceName, serviceConfig := range t.Services {
		if serviceConfig.Extensions == nil {
			serviceConfig.Extensions = make(types.Extensions)
		}

		v, ok := serviceConfig.Extensions[ServiceWorkflowExtensionName]
		if !ok {
			v = NewServiceWorkflows()
			serviceConfig.Extensions[ServiceWorkflowExtensionName] = v
		}

		workflows := v.(ServiceWorkflows)
		for _, workflowName := range workflowcommon.WorkflowNames {
			if _, ok := workflows[workflowName.String()]; !ok {
				workflows[workflowName.String()] = ServiceWorkflowConfig{
					Timeout: &defaultDuration,
				}
			} else if workflows[workflowName.String()].Timeout == nil {
				workflowConfig := workflows[workflowName.String()]
				workflowConfig.Timeout = &defaultDuration

				workflows[workflowName.String()] = workflowConfig
			}
		}

		t.Services[serviceName] = serviceConfig
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
