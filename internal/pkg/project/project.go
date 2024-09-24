package project

import (
	"errors"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/compose_exporter"
	"github.com/spaulg/solo/internal/pkg/config"
	"github.com/spaulg/solo/internal/pkg/project_file"
	"os"
	"os/exec"
	"path"
)

type Project struct {
	Config      *config.Config
	ProjectFile *project_file.ProjectFile
	ComposeFile string
}

func New(config *config.Config, projectFile *project_file.ProjectFile) Project {
	return Project{
		Config:      config,
		ProjectFile: projectFile,
		ComposeFile: path.Join(projectFile.Directory, ".solo", "docker-compose.yml"),
	}
}

func (p Project) DumpComposeConfig() {
	composeYml, _ := compose_exporter.ExportComposeConfiguration(p.Config, p.ProjectFile)
	fmt.Println(string(composeYml))
}

func (p Project) Start() {
	composeYml, _ := compose_exporter.ExportComposeConfiguration(p.Config, p.ProjectFile)

	composeDirectory := path.Dir(p.ComposeFile)
	if _, err := os.Stat(composeDirectory); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Println(fmt.Errorf("failed to check .solo directory existence: %v", err))
			os.Exit(1)
		}

		if err := os.MkdirAll(composeDirectory, 0755); err != nil {
			fmt.Println(fmt.Errorf("failed to create .solo directory: %v", err))
			os.Exit(1)
		}
	}

	if err := os.WriteFile(p.ComposeFile, composeYml, 0640); err != nil {
		fmt.Println(fmt.Errorf("failed to write compose file: %v", err))
		os.Exit(1)
	}

	composeCmd := exec.Command("/usr/local/bin/docker", "compose", "-f", p.ComposeFile, "up", "-d")

	if err := composeCmd.Run(); err != nil {
		fmt.Println(fmt.Errorf("error running composeCmd: %v", err))
		os.Exit(1)
	}
}

func (p Project) Stop() {
	if _, err := os.Stat(p.ComposeFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("compose file not found")
			os.Exit(1)
		} else {
			fmt.Println(fmt.Errorf("error running composeCmd: %v", err))
			os.Exit(1)
		}
	}

	composeCmd := exec.Command("/usr/local/bin/docker", "compose", "-f", p.ComposeFile, "stop")

	if err := composeCmd.Run(); err != nil {
		fmt.Println(fmt.Errorf("error running compose: %v", err))
	}
}

func (p Project) Destroy() {
	if _, err := os.Stat(p.ComposeFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("compose file not found")
			os.Exit(1)
		} else {
			fmt.Println(fmt.Errorf("error running composeCmd: %v", err))
			os.Exit(1)
		}
	}

	composeCmd := exec.Command("/usr/local/bin/docker", "compose", "-f", p.ComposeFile, "down", "-v")

	if err := composeCmd.Run(); err != nil {
		fmt.Println(fmt.Errorf("error running compose: %v", err))
	}

	if err := os.Remove(p.ComposeFile); err != nil {
		fmt.Println(fmt.Errorf("failed to remove compose file: %v", err))
		os.Exit(1)
	}
}
