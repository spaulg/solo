package project

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/project_file"
	"os/exec"
)

type Project struct {
	ProjectFile *project_file.ProjectFile
}

func New(projectFile *project_file.ProjectFile) Project {
	return Project{
		ProjectFile: projectFile,
	}
}

func (d Project) DumpComposeConfig() {
	composeYml, _ := d.ProjectFile.ExportComposeConfiguration()
	fmt.Println(string(composeYml))
}

func (d Project) Start() {
	d.ProjectFile.ExportComposeConfiguration()

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
