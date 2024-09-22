package project

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/compose"
	"github.com/spaulg/solo/internal/pkg/project_file"
	"os/exec"
)

type DockerComposeProject struct {
	ProjectFile *project_file.ProjectFile
}

func New(projectFile *project_file.ProjectFile) Project {
	return DockerComposeProject{
		ProjectFile: projectFile,
	}
}

func (d DockerComposeProject) ComposeConfig() {
	composeYml := compose.GenerateCompose(d.ProjectFile.FilePath)
	fmt.Println(string(composeYml))
}

func (d DockerComposeProject) Start() {
	compose.GenerateCompose(d.ProjectFile.FilePath)

	// todo: write the the new yml to a hidden file

	composeCmd := exec.Command("/usr/local/bin/docker", "composeCmd", "-f", d.ProjectFile.FilePath, "up", "-d")

	if err := composeCmd.Run(); err != nil {
		fmt.Println(fmt.Errorf("error running composeCmd: %v", err))
	}
}

func (d DockerComposeProject) Stop() {
	// todo: find the temp docker-compose.yml and turn off containers

	// todo: run docker compose down on the project
	composeCmd := exec.Command("/usr/local/bin/docker", "compose", "-f", d.ProjectFile.FilePath, "down")

	if err := composeCmd.Run(); err != nil {
		fmt.Println(fmt.Errorf("error running compose: %v", err))
	}
}

func (d DockerComposeProject) Destroy() {

}
