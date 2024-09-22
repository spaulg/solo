package project

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/project_file"
	"github.com/spf13/viper"
	"os/exec"
)

type Project struct {
	Config      *viper.Viper
	ProjectFile *project_file.ProjectFile
}

func New(config *viper.Viper, projectFile *project_file.ProjectFile) Project {
	return Project{
		Config:      config,
		ProjectFile: projectFile,
	}
}

func (d Project) DumpComposeConfig() {
	composeYml, _ := d.ProjectFile.ExportComposeConfiguration(d.Config)
	fmt.Println(string(composeYml))
}

func (d Project) Start() {
	d.ProjectFile.ExportComposeConfiguration(d.Config)

	// todo: write the the new yml to a hidden file

	composeCmd := exec.Command("/usr/local/bin/docker", "composeCmd", "-f", d.ProjectFile.FilePath, "up", "-d")

	if err := composeCmd.Run(); err != nil {
		fmt.Println(fmt.Errorf("error running composeCmd: %v", err))
	}
}

func (d Project) Stop() {
	// todo: find the temp docker-compose.yml and turn off containers

	// todo: run docker compose down on the project
	composeCmd := exec.Command("/usr/local/bin/docker", "compose", "-f", d.ProjectFile.FilePath, "down")

	if err := composeCmd.Run(); err != nil {
		fmt.Println(fmt.Errorf("error running compose: %v", err))
	}
}

func (d Project) Destroy() {

}
