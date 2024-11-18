package orchestrator

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/project"
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

func (o *DockerOrchestrator) GetHostGatewayHostname() string {
	return "host.docker.internal"
}

func (o *DockerOrchestrator) ExportComposeConfiguration(config *config.Config, project *project.Project) ([]byte, error) {
	projectOptionsLoader := cli.WithLoadOptions(func(option *loader.Options) {
		option.SkipValidation = true // Prevent validation failures from preventing the global config from being loaded
		option.ResolvePaths = false  // Keep paths relative in case the user moves their project folder
	})

	projectOptions, err := cli.NewProjectOptions([]string{project.GetFilePath()}, projectOptionsLoader)
	if err != nil {
		fmt.Println(fmt.Errorf("error building project options: %v", err))
	}

	compose, err := projectOptions.LoadProject(context.Background())
	if err != nil {
		fmt.Println(fmt.Errorf("error loading project: %v", err))
	}

	soloEntrypoint := config.Entrypoint

	for index, service := range compose.Services {
		// Replace the entrypoint of each service. if an existing entrypoint has been set, prepend this to command
		if len(service.Entrypoint) > 0 {
			service.Command = append(service.Entrypoint, service.Command...)
		}

		service.Entrypoint = []string{"/solo-entrypoint"}

		serviceDataPath := project.GetServiceStateDirectory(service.Name)
		allServicesDataPath := project.GetAllServicesStateDirectory()

		// Append volume mounts for the new entrypoint
		service.Volumes = append(service.Volumes, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   soloEntrypoint,
			Target:   "/solo-entrypoint",
			ReadOnly: true,
		}, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   serviceDataPath,
			Target:   "/solo/service",
			ReadOnly: true,
			Bind: &types.ServiceVolumeBind{
				CreateHostPath: true,
			},
		}, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   allServicesDataPath,
			Target:   "/solo/services_all",
			ReadOnly: true,
			Bind: &types.ServiceVolumeBind{
				CreateHostPath: true,
			},
		})

		compose.Services[index] = service
	}

	return compose.MarshalYAML()
}
