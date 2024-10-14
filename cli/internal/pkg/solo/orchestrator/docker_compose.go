package orchestrator

import (
	"os/exec"
)

type DockerComposeOrchestrator struct{}

func (o *DockerComposeOrchestrator) Start(projectDirectory string, composeFile string) error {
	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", composeFile,
		"--project-directory", projectDirectory,
		"up", "-d")

	if err := composeCmd.Run(); err != nil {
		return err
	}

	return nil
}

func (o *DockerComposeOrchestrator) Stop(projectDirectory string, composeFile string) error {
	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", composeFile,
		"--project-directory", projectDirectory,
		"stop")

	if err := composeCmd.Run(); err != nil {
		return err
	}

	return nil
}

func (o *DockerComposeOrchestrator) Destroy(projectDirectory string, composeFile string) error {
	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", composeFile,
		"--project-directory", projectDirectory,
		"down", "-v")

	if err := composeCmd.Run(); err != nil {
		return err
	}

	return nil
}
