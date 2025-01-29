package container

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"io"
	"os"
	"os/exec"
)

type DockerOrchestrator struct {
	projectDirectory string
	composeFile      string
}

type RunningService struct {
	Service string `json:"service"`
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

func (t *DockerOrchestrator) RunningServices() ([]string, error) {
	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", t.composeFile,
		"--project-directory", t.projectDirectory,
		"ps",
		"--format", "json",
		"--filter", "status=running",
	)

	var stdoutBuf bytes.Buffer
	composeCmd.Stdout = &stdoutBuf

	if err := composeCmd.Run(); err != nil {
		return nil, err
	}

	var serviceNames []string
	serviceNameMap := make(map[string]bool)

	buffer := stdoutBuf.Bytes()
	reader := bufio.NewReader(bytes.NewReader(buffer))

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		runningService := RunningService{}
		if err := json.Unmarshal(line, &runningService); err != nil {
			return nil, err
		}

		if !serviceNameMap[runningService.Service] {
			serviceNameMap[runningService.Service] = true
			serviceNames = append(serviceNames, runningService.Service)
		}
	}

	return serviceNames, nil
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

		serviceLogPath := project.GetServiceLogDirectory(service.Name)
		_, err := os.Stat(serviceLogPath)
		if err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(serviceLogPath, 0750); err != nil {
					return nil, fmt.Errorf("failed to create service data directoiry: %v", err)
				}
			} else {
				return nil, fmt.Errorf("failed to create service data directoiry: %v", err)
			}
		}

		serviceDataPath := project.GetServiceMountDirectory(service.Name)
		_, err = os.Stat(serviceDataPath)
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
			Source:   serviceLogPath,
			Target:   "/solo/service/logs",
			ReadOnly: false,
		}, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   serviceDataPath,
			Target:   "/solo/service/data",
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
