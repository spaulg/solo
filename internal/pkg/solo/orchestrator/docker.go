package orchestrator

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"os"
	"os/exec"
)

type DockerOrchestrator struct{}

func (o *DockerOrchestrator) Up(projectDirectory string, composeFile string) error {
	composeCmd := exec.Command("/usr/bin/docker", "compose",
		"-f", composeFile,
		"--project-directory", projectDirectory,
		"up", "-d")

	if err := composeCmd.Run(); err != nil {
		return err
	}

	return nil
}

func (o *DockerOrchestrator) Down(projectDirectory string, composeFile string) error {
	composeCmd := exec.Command("/usr/bin/docker", "compose",
		"-f", composeFile,
		"--project-directory", projectDirectory,
		"stop")

	if err := composeCmd.Run(); err != nil {
		return err
	}

	return nil
}

func (o *DockerOrchestrator) Destroy(projectDirectory string, composeFile string) error {
	composeCmd := exec.Command("/usr/bin/docker", "compose",
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

	allServicesDataPath := project.GetAllServicesStateDirectory()
	_, err = os.Stat(allServicesDataPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(allServicesDataPath, 0750); err != nil {
				return nil, fmt.Errorf("failed to create all services data directoiry: %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to create all services data directoiry: %v", err)
		}
	}

	for index, service := range compose.Services {
		// Replace the entrypoint of each service. if an existing entrypoint has been set, prepend this to command
		if len(service.Entrypoint) > 0 {
			service.Command = append(service.Entrypoint, service.Command...)
		}

		service.Entrypoint = []string{"/solo-entrypoint"}

		serviceDataPath := project.GetServiceStateDirectory(service.Name)
		_, err := os.Stat(serviceDataPath)
		if err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(serviceDataPath, 0750); err != nil {
					return nil, fmt.Errorf("failed to create service data directoiry: %v", err)
				}
			} else {
				return nil, fmt.Errorf("failed to create service data directoiry: %v", err)
			}
		}

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
		}, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   allServicesDataPath,
			Target:   "/solo/services_all",
			ReadOnly: true,
		})

		service.ExtraHosts = types.HostsList{}
		service.ExtraHosts["host.docker.internal"] = []string{"host-gateway"}

		compose.Services[index] = service
	}

	return compose.MarshalYAML()
}
