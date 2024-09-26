package project_file

import (
	"github.com/spaulg/solo/internal/pkg/schema"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type ProjectFile struct {
	Directory string
	FilePath  string
}

func New(projectFilePath string) *ProjectFile {
	return &ProjectFile{
		Directory: filepath.Dir(projectFilePath),
		FilePath:  projectFilePath,
	}
}

func (p *ProjectFile) Marshall() (*schema.Config, error) {
	fileContents, err := os.ReadFile(p.FilePath)
	if err != nil {
		return nil, err
	}

	projectConfig := schema.Config{}
	if err := yaml.Unmarshal(fileContents, &projectConfig); err != nil {
		return nil, err
	}

	return &projectConfig, nil
}
