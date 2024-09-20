package project_finder

import (
	"errors"
	"github.com/spaulg/solo/internal/pkg/project"
	"os"
	"path/filepath"
)

// FindProject Find the project by navigating up the
// filesystem tree until the project file is found, or
// return error if no project is found
func FindProject() (project.Project, error) {
	var projectFilePath = ""

	path, err := filepath.Abs("./")
	if err != nil {
		return nil, err
	}

	for {
		projectFilePath = filepath.Join(path, "docker-compose.yml")
		fileInfo, err := os.Stat(projectFilePath)

		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if path == "/" {
					break
				}

				path, err = filepath.Abs(filepath.Join(path, ".."))
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		} else if fileInfo != nil {
			return project.New(projectFilePath), nil
		}
	}

	return nil, errors.New("filesystem root reached, project file not found")
}
