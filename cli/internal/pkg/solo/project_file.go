package solo

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type ProjectFile struct {
	Directory string
	FilePath  string
}

func NewProjectFile(projectFilePath string) *ProjectFile {
	return &ProjectFile{
		Directory: filepath.Dir(projectFilePath),
		FilePath:  projectFilePath,
	}
}

func (p *ProjectFile) Marshall() (*ProjectConfig, error) {
	fileContents, err := os.ReadFile(p.FilePath)
	if err != nil {
		return nil, err
	}

	projectConfig := ProjectConfig{}
	if err := yaml.Unmarshal(fileContents, &projectConfig); err != nil {
		return nil, err
	}

	return &projectConfig, nil
}
