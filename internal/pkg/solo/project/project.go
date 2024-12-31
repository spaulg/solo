package project

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type StepConfig struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	Cwd     string `yaml:"cwd"`
	Timeout int    `yaml:"timeout"`
}

type StepsConfig struct {
	Provisioning []StepConfig `yaml:"provisioning"`
	PreStart     []StepConfig `yaml:"pre_start"`
	PostStart    []StepConfig `yaml:"post_start"`
	PreStop      []StepConfig `yaml:"pre_stop"`
	PostStop     []StepConfig `yaml:"post_stop"`
	PreDestroy   []StepConfig `yaml:"pre_destroy"`
	PostDestroy  []StepConfig `yaml:"post_destroy"`
}

type ServiceConfig struct {
	Steps StepsConfig `yaml:"steps"`
}

type Services map[string]ServiceConfig

type ProjectConfig struct {
	ComposeFile *string  `yaml:"compose_file"`
	Services    Services `yaml:"services"`
}

type Project struct {
	projectStateDirectory string
	directory             string
	filePath              string
	config                ProjectConfig
	serviceNames          []string
}

func NewProject(projectFilePath string) (*Project, error) {
	workingDirectory := filepath.Dir(projectFilePath)

	project := &Project{
		projectStateDirectory: workingDirectory + "/.solo",
		directory:             workingDirectory,
		filePath:              projectFilePath,
	}

	bytes, err := os.ReadFile(projectFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project file: %v", err)
	}

	if err := yaml.Unmarshal(bytes, &project.config); err != nil {
		return nil, fmt.Errorf("failed to parse project file: %v", err)
	}

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

func (t *Project) GetComposePath() string {
	if t.config.ComposeFile != nil {
		return *t.config.ComposeFile
	} else {
		return t.filePath
	}
}

func (t *Project) ServiceNames() []string {
	if t.serviceNames == nil {
		serviceNames := make([]string, len(t.config.Services))
		for serviceName := range t.config.Services {
			serviceNames = append(serviceNames, serviceName)
		}

		t.serviceNames = serviceNames
	}

	return t.serviceNames
}
