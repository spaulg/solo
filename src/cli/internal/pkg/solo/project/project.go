package project

import (
	"path/filepath"
)

type Project struct {
	projectStateDirectory string
	directory             string
	filePath              string
}

func NewProject(projectFilePath string) *Project {
	workingDirectory := filepath.Dir(projectFilePath)

	return &Project{
		projectStateDirectory: workingDirectory + "/.solo",
		directory:             workingDirectory,
		filePath:              projectFilePath,
	}
}

func (t *Project) ResolveStateDirectory(relativePath string) string {
	return t.projectStateDirectory + "/" + relativePath
}

func (t *Project) GetAllServicesStateDirectory() string {
	return t.projectStateDirectory + "/services_all"
}

func (t *Project) GetServiceStateDirectory(serviceName string) string {
	return t.projectStateDirectory + "/services/" + serviceName
}

func (t *Project) GetDirectory() string {
	return t.directory
}

func (t *Project) GetFilePath() string {
	return t.filePath
}
