package project

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

type WorkflowStep struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	Cwd     string `yaml:"cwd"`
}

type Workflows map[string][]WorkflowStep

type Project struct {
	projectStateDirectory string
	directory             string
	filePath              string
	compose               *types.Project
}

func NewProject(projectFilePath string) (*Project, error) {
	projectOptions, err := cli.NewProjectOptions(nil,
		WithComposeFiles(projectFilePath),
		cli.WithLoadOptions(func(option *loader.Options) {
			option.ResolvePaths = false // Keep paths relative in case the user moves their project folder
		}),
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

	return project, nil
}

func (t *Project) GetCompose() *types.Project {
	composeCopy := *t.compose
	return &composeCopy
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

func (t *Project) ServiceNames() []string {
	return t.compose.ServiceNames()
}

func (t *Project) GetServiceWorkflow(serviceName string, eventName string) []WorkflowStep {
	workflows := Workflows{}
	if ok, _ := t.compose.Services[serviceName].Extensions.Get("x-workflows", &workflows); !ok {
		return nil
	}

	return workflows[eventName]
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
