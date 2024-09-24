package project_file

import (
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
