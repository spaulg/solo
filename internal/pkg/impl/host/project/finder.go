package project

import (
	"errors"
	"os"
	"path/filepath"

	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
)

const DefaultProjectFileName = "solo.yml"

// FindProject Find the project file by navigating up the
// filesystem tree until the project file is found, or
// return error if no project file is found
func FindProject(startPath string, config *config_types.Config, profiles []string) (project_types.Project, error) {
	var projectFilePath string

	path, err := filepath.Abs(startPath)
	if err != nil {
		return nil, err
	}

	for {
		projectFilePath = filepath.Join(path, DefaultProjectFileName)
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
			return NewProject(projectFilePath, config, profiles)
		}
	}

	return nil, errors.New("filesystem root reached, project file not found")
}
