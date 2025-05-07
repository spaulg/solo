package container

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/container/progress"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"google.golang.org/grpc/metadata"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

type ServiceStatus struct {
	Service string `json:"Service"`
	State   string `json:"State"`
}

type DockerInspect struct {
	Name string `json:"Name"`
}

type DockerOrchestrator struct {
	soloCtx          *context.CliContext
	eventManager     events.Manager
	projectDirectory string
	composeFile      string
}

func NewDockerOrchestrator(
	soloCtx *context.CliContext,
	eventManager events.Manager,
) Orchestrator {
	return &DockerOrchestrator{
		soloCtx:          soloCtx,
		eventManager:     eventManager,
		projectDirectory: soloCtx.Project.GetDirectory(),
		composeFile:      soloCtx.Project.GetGeneratedComposeFilePath(),
	}
}

func (t *DockerOrchestrator) runComposeCommandWithProgress(arguments ...string) error {
	composeCmd := exec.Command("/usr/local/bin/docker", append([]string{"compose"}, arguments...)...)

	stderr, err := composeCmd.StderrPipe()
	if err != nil {
		return err
	}

	defer stderr.Close()

	if err := composeCmd.Start(); err != nil {
		return err
	}

	// Stream the progress events
	eventStreamReader := progress.NewProgressEventPublisher(t.soloCtx, t.eventManager, t.soloCtx.Project.Name, stderr)
	go func(eventStreamReader *progress.ProgressEventStreamer) {
		eventStreamReader.PublishStreamedProgressEvents()
	}(eventStreamReader)

	if err := composeCmd.Wait(); err != nil {
		return err
	}

	return nil
}

func (t *DockerOrchestrator) Up() error {
	return t.runComposeCommandWithProgress(
		"--progress", "json",
		"-f", t.composeFile,
		"--project-directory", t.projectDirectory,
		"up", "-d",
	)
}

func (t *DockerOrchestrator) Down() error {
	return t.runComposeCommandWithProgress(
		"--progress", "json",
		"-f", t.composeFile,
		"--project-directory", t.projectDirectory,
		"stop",
	)
}

func (t *DockerOrchestrator) Destroy() error {
	return t.runComposeCommandWithProgress(
		"--progress", "json",
		"-f", t.composeFile,
		"--project-directory", t.projectDirectory,
		"down", "-v",
	)
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

func (t *DockerOrchestrator) ServicesStatus() ([]string, []string, error) {
	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", t.composeFile,
		"--project-directory", t.projectDirectory,
		"ps",
		"--format", "json",
	)

	var stdoutBuf bytes.Buffer
	composeCmd.Stdout = &stdoutBuf

	if err := composeCmd.Run(); err != nil {
		return nil, nil, err
	}

	var runningServiceNames []string
	runningServiceNameMap := make(map[string]bool)

	var stoppedServiceNames []string
	stoppedServiceNameMap := make(map[string]bool)

	buffer := stdoutBuf.Bytes()
	reader := bufio.NewReader(bytes.NewReader(buffer))

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, nil, err
		}

		serviceStatus := ServiceStatus{}
		if err := json.Unmarshal(line, &serviceStatus); err != nil {
			return nil, nil, err
		}

		// Filter for running/stopped services
		if serviceStatus.State == "running" && !runningServiceNameMap[serviceStatus.Service] {
			runningServiceNameMap[serviceStatus.Service] = true
			runningServiceNames = append(runningServiceNames, serviceStatus.Service)
		} else if serviceStatus.State == "stopped" && !stoppedServiceNameMap[serviceStatus.Service] {
			stoppedServiceNameMap[serviceStatus.Service] = true
			stoppedServiceNames = append(stoppedServiceNames, serviceStatus.Service)
		}
	}

	// Add services with no container
	for _, service := range t.soloCtx.Project.ServiceNames() {
		if !runningServiceNameMap[service] && !stoppedServiceNameMap[service] {
			// Service is neither running nor stopped
			stoppedServiceNames = append(stoppedServiceNames, service)
		}
	}

	return runningServiceNames, stoppedServiceNames, nil
}

func (t *DockerOrchestrator) GetHostGatewayHostname() string {
	return "host.docker.internal"
}

func (t *DockerOrchestrator) ExportComposeConfiguration(config *config.Config, project *project.Project) ([]byte, error) {
	soloEntrypoint := path.Join(project.GetStateDirectoryRoot(), "solo-entrypoint")

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

		service.Entrypoint = []string{config.Entrypoint.ContainerEntrypointPath, "entrypoint"}

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
			Target:   config.Entrypoint.ContainerEntrypointPath,
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

func (t *DockerOrchestrator) ResolveServiceNameFromContainerName(containerName string) (*string, error) {
	serviceName := "example"
	return &serviceName, nil
}

func (t *DockerOrchestrator) ResolveContainerNameFromMetadata(md metadata.MD) (*string, error) {
	containerNames := md.Get("hostname")

	if len(containerNames) == 0 {
		return nil, fmt.Errorf("unable to resolve container name")
	}

	composeCmd := exec.Command("/usr/local/bin/docker", "inspect",
		"--format", "{{ json . }}",
		"--type", "container",
		containerNames[0],
	)

	var stdoutBuf bytes.Buffer
	composeCmd.Stdout = &stdoutBuf

	if err := composeCmd.Run(); err != nil {
		t.soloCtx.Logger.Error("Failed to run")
		return nil, err
	}

	inspect := DockerInspect{}
	if err := json.Unmarshal(stdoutBuf.Bytes(), &inspect); err != nil {
		t.soloCtx.Logger.Error("Failed to unmarshall")
		return nil, err
	}

	containerName := inspect.Name
	containerName = strings.TrimLeft(containerName, "/")

	return &containerName, nil
}
