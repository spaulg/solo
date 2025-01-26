package container

import (
	"fmt"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"os"
	"os/exec"
)

type DockerOrchestrator struct {
	projectDirectory string
	composeFile      string
}

func (t *DockerOrchestrator) Up() error {
	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", t.composeFile,
		"--project-directory", t.projectDirectory,
		"up", "-d")

	if err := composeCmd.Run(); err != nil {
		return err
	}

	return nil
}

func (t *DockerOrchestrator) Down() error {
	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", t.composeFile,
		"--project-directory", t.projectDirectory,
		"stop")

	if err := composeCmd.Run(); err != nil {
		return err
	}

	return nil
}

func (t *DockerOrchestrator) Destroy() error {
	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", t.composeFile,
		"--project-directory", t.projectDirectory,
		"down", "-v")

	if err := composeCmd.Run(); err != nil {
		return err
	}

	return nil
}

func (t *DockerOrchestrator) Execute(serviceNames []string, command []string) error {
	for _, serviceName := range serviceNames {
		composeCmd := exec.Command("/usr/local/bin/docker", append([]string{"compose",
			"-f", t.composeFile,
			"--project-directory", t.projectDirectory,
			"exec", "-d", serviceName,
		}, command...)...)

		if err := composeCmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (t *DockerOrchestrator) GetHostGatewayHostname() string {
	return "host.docker.internal"
}

func (t *DockerOrchestrator) ExportComposeConfiguration(config *config.Config, project *project.Project) ([]byte, error) {
	soloEntrypoint := config.Entrypoint

	allServicesDataPath := project.GetAllServicesStateDirectory()
	_, err := os.Stat(allServicesDataPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(allServicesDataPath, 0750); err != nil {
				return nil, fmt.Errorf("failed to create all services data directoiry: %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to create all services data directoiry: %v", err)
		}
	}

	for index, service := range project.Services {
		// Replace the entrypoint of each service. if an existing entrypoint has been set, prepend this to command
		if len(service.Entrypoint) > 0 {
			service.Command = append(service.Entrypoint, service.Command...)
		}

		service.Entrypoint = []string{"/usr/local/sbin/solo", "entrypoint"}

		serviceDataPath := project.GetServiceMountDirectory(service.Name)
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
			Target:   "/usr/local/sbin/solo",
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

		project.Services[index] = service
	}

	return project.MarshalYAML()
}
