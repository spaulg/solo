package project

import (
	"errors"
	"os"
	"path/filepath"
)

const ProjectFileName = "solo.yml"

// FindProjectFile Find the project file by navigating up the
// filesystem tree until the project file is found, or
// return error if no project file is found
func FindProjectFile() (*Project, error) {
	var projectFilePath = ""

	path, err := filepath.Abs("./")
	if err != nil {
		return nil, err
	}

	for {
		projectFilePath = filepath.Join(path, ProjectFileName)
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
			return NewProject(projectFilePath), nil
		}
	}

	return nil, errors.New("filesystem root reached, project file not found")
}
