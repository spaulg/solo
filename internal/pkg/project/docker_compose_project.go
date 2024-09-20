package project

import (
	"fmt"
	"os/exec"
)

type DockerComposeProject struct {
	projectFilePath string
}

func New(projectFilePath string) Project {
	return DockerComposeProject{
		projectFilePath: projectFilePath,
	}
}

func (d DockerComposeProject) Start() {
	// todo: read the docker-compose.yml based project file
	// todo: strip out anything not official docker-compose
	// todo: replace the entrypoint of each service. if an existing entrypoint has been set, prepend this to command
	// todo: override any user config to apply root
	// todo: append volume mounts for the new entrypoint, build scripts, run scripts and preferred user id config
	// todo: write the the new docker-compose file to a working hidden directory

	compose := exec.Command("/usr/local/bin/docker", "compose", "-f", d.projectFilePath, "up", "-d")

	if err := compose.Run(); err != nil {
		fmt.Println(fmt.Errorf("error running compose: %v", err))
	}
}

func (d DockerComposeProject) Stop() {
	// todo: find the temp docker-compose.yml and turn off containers

	// todo: run docker compose down on the project
	compose := exec.Command("/usr/local/bin/docker", "compose", "-f", d.projectFilePath, "down")

	if err := compose.Run(); err != nil {
		fmt.Println(fmt.Errorf("error running compose: %v", err))
	}
}

func (d DockerComposeProject) Destroy() {

}
