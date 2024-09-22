package project

import (
	"errors"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/project_file"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path"
)

type Project struct {
	Config      *viper.Viper
	ProjectFile *project_file.ProjectFile
	ComposeFile string
}

func New(config *viper.Viper, projectFile *project_file.ProjectFile) Project {
	return Project{
		Config:      config,
		ProjectFile: projectFile,
		ComposeFile: path.Join(projectFile.Directory, ".solo", "docker-compose.yml"),
	}
}

func (d Project) DumpComposeConfig() {
	composeYml, _ := d.ProjectFile.ExportComposeConfiguration(d.Config)
	fmt.Println(string(composeYml))
}

func (d Project) Start() {
	composeYml, _ := d.ProjectFile.ExportComposeConfiguration(d.Config)

	composeDirectory := path.Dir(d.ComposeFile)
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

	if err := os.WriteFile(d.ComposeFile, composeYml, 0640); err != nil {
		fmt.Println(fmt.Errorf("failed to write compose file: %v", err))
		os.Exit(1)
	}

	composeCmd := exec.Command("/usr/local/bin/docker", "compose", "-f", d.ComposeFile, "up", "-d")

	if err := composeCmd.Run(); err != nil {
		output, _ := composeCmd.CombinedOutput()

		fmt.Println(fmt.Errorf("error running composeCmd: %v", err))
		fmt.Println(string(output))
		os.Exit(1)
	}
}

func (d Project) Stop() {
	if _, err := os.Stat(d.ComposeFile); err != nil {
		if errors.Is(os.ErrNotExist, err) {
			fmt.Println("compose file not found")
			os.Exit(1)
		} else {
			fmt.Println(fmt.Errorf("error running composeCmd: %v", err))
			os.Exit(1)
		}
	}

	composeCmd := exec.Command("/usr/local/bin/docker", "compose", "-f", d.ComposeFile, "down")

	if err := composeCmd.Run(); err != nil {
		fmt.Println(fmt.Errorf("error running compose: %v", err))
	}
}

func (d Project) Destroy() {

}
