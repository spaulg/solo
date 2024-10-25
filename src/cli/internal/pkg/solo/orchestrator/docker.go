package orchestrator

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/spaulg/solo/cli/internal/pkg/solo/config"
	"os/exec"
)

type DockerOrchestrator struct{}

func (o *DockerOrchestrator) Up(projectDirectory string, composeFile string) error {
	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", composeFile,
		"--project-directory", projectDirectory,
		"up", "-d")

	if err := composeCmd.Run(); err != nil {
		return err
	}

	return nil
}

func (o *DockerOrchestrator) Down(projectDirectory string, composeFile string) error {
	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", composeFile,
		"--project-directory", projectDirectory,
		"stop")

	if err := composeCmd.Run(); err != nil {
		return err
	}

	return nil
}

func (o *DockerOrchestrator) Destroy(projectDirectory string, composeFile string) error {
	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", composeFile,
		"--project-directory", projectDirectory,
		"down", "-v")

	if err := composeCmd.Run(); err != nil {
		return err
	}

	return nil
}

func (o *DockerOrchestrator) ExportComposeConfiguration(globalConfig *config.Config, projectPath string) ([]byte, error) {
	projectOptionsLoader := cli.WithLoadOptions(func(option *loader.Options) {
		option.SkipValidation = true // Prevent validation failures from preventing the global config from being loaded
		option.ResolvePaths = false  // Keep paths relative in case the user moves their project folder
	})

	projectOptions, err := cli.NewProjectOptions([]string{projectPath}, projectOptionsLoader)

	if err != nil {
		fmt.Println(fmt.Errorf("error building project options: %v", err))
	}

	soloEntrypoint := globalConfig.Entrypoint

	project, err := projectOptions.LoadProject(context.Background())
	if err != nil {
		fmt.Println(fmt.Errorf("error loading project: %v", err))
	}

	for index, service := range project.Services {
		// Override any user globalConfig to force root
		// todo: allow the user to switch in the entrypoint script
		service.User = "0"

		// Replace the entrypoint of each service. if an existing entrypoint has been set, prepend this to command
		if len(service.Entrypoint) > 0 {
			service.Command = append(service.Entrypoint, service.Command...)
		}

		service.Entrypoint = []string{"/solo-entrypoint"}

		// Append volume mounts for the new entrypoint
		service.Volumes = append(service.Volumes, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   soloEntrypoint,
			Target:   "/solo-entrypoint",
			ReadOnly: true,
		})

		project.Services[index] = service
	}

	return project.MarshalYAML()
}
