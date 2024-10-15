package project

import (
	"path/filepath"
)

type Project struct {
	Directory string
	FilePath  string
}

func NewProject(projectFilePath string) *Project {
	return &Project{
		Directory: filepath.Dir(projectFilePath),
		FilePath:  projectFilePath,
	}
}

//func (p *Project) Marshall() (*ProjectConfig, error) {
//	fileContents, err := os.ReadFile(p.FilePath)
//	if err != nil {
//		return nil, err
//	}
//
//	projectConfig := ProjectConfig{}
//	if err := yaml.Unmarshal(fileContents, &projectConfig); err != nil {
//		return nil, err
//	}
//
//	return &projectConfig, nil
//}
