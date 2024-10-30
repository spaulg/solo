package project

import (
	"path/filepath"
)

type Project struct {
	ProjectStateDirectory string
	Directory             string
	FilePath              string
}

func NewProject(projectFilePath string) *Project {
	workingDirectory := filepath.Dir(projectFilePath)

	return &Project{
		ProjectStateDirectory: workingDirectory + "/.solo",
		Directory:             workingDirectory,
		FilePath:              projectFilePath,
	}
}

func (t *Project) ResolveStateDirectory(relativePath string) string {
	return t.ProjectStateDirectory + "/" + relativePath
}

func (t *Project) GetAllServicesStateDirectory() string {
	return t.ProjectStateDirectory + "/services_all"
}

func (t *Project) GetServiceStateDirectory(serviceName string) string {
	return t.ProjectStateDirectory + "/services/" + serviceName
}
