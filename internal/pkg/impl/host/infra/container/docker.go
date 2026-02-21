package container

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/compose-spec/compose-go/v2/types"
	"google.golang.org/grpc/metadata"

	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/container/progress"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/app/events"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/domain"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/infra/container"
)

type ComposeServiceStatus struct {
	Service string `json:"Service"`
	State   string `json:"State"`
}

type DockerInspect struct {
	Name   string `json:"Name"`
	Config struct {
		WorkingDir string `json:"WorkingDir"`
	} `json:"Config"`
}

type DockerOrchestrator struct {
	soloCtx           *context.CliContext
	eventManager      events_types.Manager
	dockerCommandPath string
	projectDirectory  string
	composeFile       string
}

func NewDockerOrchestrator(
	soloCtx *context.CliContext,
	eventManager events_types.Manager,
	dockerCommandPath string,
) container_types.Orchestrator {
	return &DockerOrchestrator{
		soloCtx:           soloCtx,
		eventManager:      eventManager,
		dockerCommandPath: dockerCommandPath,
		projectDirectory:  soloCtx.Project.GetDirectory(),
		composeFile:       soloCtx.Project.GetGeneratedComposeFilePath(),
	}
}

func (t *DockerOrchestrator) runComposeCommandWithProgress(arguments ...string) error {
	// nolint:gosec
	composeCmd := exec.Command(t.dockerCommandPath, append([]string{"compose"}, arguments...)...)

	stderr, err := composeCmd.StderrPipe()
	if err != nil {
		return err
	}

	defer stderr.Close()

	if err := composeCmd.Start(); err != nil {
		return err
	}

	// Stream the progress events
	eventStreamReader := progress.NewProgressEventPublisher(t.soloCtx, t.eventManager, t.soloCtx.Project.Name(), stderr)
	go func(eventStreamReader *progress.ProgressEventStreamer) {
		eventStreamReader.PublishStreamedProgressEvents()
	}(eventStreamReader)

	if err := composeCmd.Wait(); err != nil {
		return err
	}

	return nil
}

func (t *DockerOrchestrator) ComposeUp(serviceNames []string) error {
	arguments := append([]string{
		"--progress", "json",
		"-f", t.composeFile,
		"--project-directory", t.projectDirectory,
		"up", "-d",
	}, serviceNames...)

	return t.runComposeCommandWithProgress(arguments...)
}

func (t *DockerOrchestrator) ComposeStop(serviceNames []string) error {
	arguments := append([]string{
		"--progress", "json",
		"-f", t.composeFile,
		"--project-directory", t.projectDirectory,
		"stop",
	}, serviceNames...)

	return t.runComposeCommandWithProgress(arguments...)
}

func (t *DockerOrchestrator) ComposeDown(serviceNames []string) error {
	arguments := append([]string{
		"--progress", "json",
		"-f", t.composeFile,
		"--project-directory", t.projectDirectory,
		"down", "-v",
	}, serviceNames...)

	return t.runComposeCommandWithProgress(arguments...)
}

func (t *DockerOrchestrator) ComposeForkAndExecute(
	serviceName string,
	index int,
	command string,
	arguments []string,
	workingDirectory string,
) error {
	// Verify service is running
	status, err := t.ServicesStatus([]string{serviceName})
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	if len(status.RunningServices) != 1 {
		return fmt.Errorf("service %s is not running", serviceName)
	}

	argv := []string{
		t.dockerCommandPath, "compose",
		"-f", t.composeFile,
		"--project-directory", t.projectDirectory,
		"exec",
		"--index", strconv.Itoa(index),
	}

	if workingDirectory != "" {
		argv = append(argv, "--workdir", workingDirectory)
	}

	argv = append(argv, serviceName, command)
	argv = append(argv, arguments...)

	// nolint:gosec
	return syscall.Exec(t.dockerCommandPath, argv, os.Environ())
}

func (t *DockerOrchestrator) ForkAndExecute(
	containerName string,
	command string,
	arguments []string,
	workingDirectory string,
) error {
	argv := []string{
		t.dockerCommandPath,
		"exec", "-it",
	}

	if workingDirectory != "" {
		argv = append(argv, "--workdir", workingDirectory)
	}

	argv = append(argv, containerName, command)
	argv = append(argv, arguments...)

	// nolint:gosec
	return syscall.Exec(t.dockerCommandPath, argv, append(os.Environ(), "DOCKER_CLI_HINTS=false"))
}

func (t *DockerOrchestrator) StartCommand(containerName string, command []string) error {
	// nolint:gosec
	containerCmd := exec.Command(t.dockerCommandPath, append([]string{
		"exec", "-d", containerName,
	}, command...)...)

	if err := containerCmd.Run(); err != nil {
		return err
	}

	return nil
}

func (t *DockerOrchestrator) RunCommand(containerName string, command []string) (string, error) {
	// nolint:gosec
	containerCmd := exec.Command(t.dockerCommandPath, append([]string{
		"exec", "-t", containerName,
	}, command...)...)

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	containerCmd.Stdout = &stdoutBuf
	containerCmd.Stderr = &stderrBuf

	if err := containerCmd.Run(); err != nil {
		// Include stderr in error message for better debugging
		return "", fmt.Errorf("failed to execute command %v in container %s: %w (stderr: %s)",
			command, containerName, err, stderrBuf.String())
	}

	return stdoutBuf.String(), nil
}

func (t *DockerOrchestrator) ServicesStatus(serviceNames []string) (*container_types.ServiceStatus, error) {
	arguments := []string{
		"compose",
		"--profile", strings.Join(t.soloCtx.Project.Profiles(), ","),
		"-f", t.composeFile,
		"--project-directory", t.projectDirectory,
		"ps",
		"--format", "json",
		"--all",
	}

	if len(serviceNames) > 0 {
		arguments = append(arguments, serviceNames...)
	}

	// nolint:gosec
	composeCmd := exec.Command(t.dockerCommandPath, arguments...)

	var stdoutBuf bytes.Buffer
	composeCmd.Stdout = &stdoutBuf

	if err := composeCmd.Run(); err != nil {
		return nil, err
	}

	var runningServiceNames []string
	runningServiceNameMap := make(map[string]bool)

	var stoppedServiceNames []string
	stoppedServiceNameMap := make(map[string]bool)

	var exitedServiceNames []string
	exitedServiceNameMap := make(map[string]bool)

	var absentServiceNames []string
	var notRunningServiceNames []string

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

		serviceStatus := ComposeServiceStatus{}
		if err := json.Unmarshal(line, &serviceStatus); err != nil {
			return nil, err
		}

		// Filter for running/stopped services
		if serviceStatus.State == "running" && !runningServiceNameMap[serviceStatus.Service] {
			runningServiceNameMap[serviceStatus.Service] = true
			runningServiceNames = append(runningServiceNames, serviceStatus.Service)
		} else if serviceStatus.State == "stopped" && !stoppedServiceNameMap[serviceStatus.Service] {
			stoppedServiceNameMap[serviceStatus.Service] = true
			stoppedServiceNames = append(stoppedServiceNames, serviceStatus.Service)
			notRunningServiceNames = append(notRunningServiceNames, serviceStatus.Service)
		} else if serviceStatus.State == "exited" && !exitedServiceNameMap[serviceStatus.Service] {
			exitedServiceNameMap[serviceStatus.Service] = true
			exitedServiceNames = append(exitedServiceNames, serviceStatus.Service)
			notRunningServiceNames = append(notRunningServiceNames, serviceStatus.Service)
		}
	}

	// Add services with no container
	if serviceNames == nil {
		serviceNames = t.soloCtx.Project.Services().ServiceNames()
	}

	for _, service := range serviceNames {
		if !runningServiceNameMap[service] && !stoppedServiceNameMap[service] {
			// Service is neither running nor stopped
			absentServiceNames = append(absentServiceNames, service)
			notRunningServiceNames = append(notRunningServiceNames, service)
		}
	}

	return &container_types.ServiceStatus{
		RunningServices:    runningServiceNames,
		StoppedServices:    stoppedServiceNames,
		ExitedServices:     exitedServiceNames,
		AbsentServices:     absentServiceNames,
		NotRunningServices: notRunningServiceNames,
	}, nil
}

func (t *DockerOrchestrator) GetHostGatewayHostname() string {
	return "host.docker.internal"
}

func (t *DockerOrchestrator) ExportComposeConfiguration(config *domain.Config, project project_types.Project) ([]byte, error) {
	// Reload project with all profiles enabled
	project, err := project.ReloadWithAllProfilesEnabled()
	if err != nil {
		return nil, fmt.Errorf("failed to reload project with all services: %w", err)
	}

	soloEntrypoint := path.Join(project.GetStateDirectoryRoot(), "solo-entrypoint")

	allServicesDataPath := project.GetAllServicesStateDirectory()
	_, err = os.Stat(allServicesDataPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(allServicesDataPath, 0750); err != nil {
				return nil, fmt.Errorf("failed to create all services data directory: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to create all services data directory: %w", err)
		}
	}

	compose := project.GetCompose()

	for index, service := range compose.Services {
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
					return nil, fmt.Errorf("failed to create service data directory: %w", err)
				}
			} else {
				return nil, fmt.Errorf("failed to create service data directory: %w", err)
			}
		}

		serviceDataPath := project.GetServiceMountDirectory(service.Name)
		_, err = os.Stat(serviceDataPath)
		if err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(serviceDataPath, 0750); err != nil {
					return nil, fmt.Errorf("failed to create service data directory: %w", err)
				}
			} else {
				return nil, fmt.Errorf("failed to create service data directory: %w", err)
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

		compose.Services[index] = service
	}

	return compose.MarshalYAML()
}

func (t *DockerOrchestrator) ResolveContainerNameFromMetadata(md metadata.MD) (string, string, error) {
	containerNames := md.Get("hostname")

	if len(containerNames) == 0 {
		return "", "", fmt.Errorf("unable to resolve container name")
	}

	return t.resolveContainerNameFromIDOrName(containerNames[0])
}

func (t *DockerOrchestrator) ResolveImageWorkingDirectory(serviceName string) (string, error) {
	// nolint:gosec
	composeCmd := exec.Command(t.dockerCommandPath,
		"compose",
		"-f", t.composeFile,
		"--project-directory", t.projectDirectory,
		"images",
		"-q",
		serviceName,
	)

	var stdoutBuf bytes.Buffer
	composeCmd.Stdout = &stdoutBuf

	if err := composeCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to lookup image from service %s: %w", serviceName, err)
	}

	imageNameOrID := strings.TrimSpace(stdoutBuf.String())
	inspect, err := t.dockerInspect("image", imageNameOrID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect image %s: %w", imageNameOrID, err)
	}

	workingDirectory := "/"
	if inspect.Config.WorkingDir != "" {
		workingDirectory = inspect.Config.WorkingDir
	}

	return workingDirectory, nil
}

func (t *DockerOrchestrator) ResolveContainerNameFromServiceName(serviceName string, index int) (string, string, error) {
	service := t.soloCtx.Project.Services().GetService(serviceName).GetConfig()
	replicas := 1

	if service.Deploy != nil && service.Deploy.Replicas != nil {
		replicas = *service.Deploy.Replicas
	}

	containerName := ""
	if len(service.ContainerName) > 0 && replicas == 1 {
		containerName = service.ContainerName
	} else {
		containerName = fmt.Sprintf("%s-%s-%d", t.soloCtx.Project.Name(), serviceName, index)
	}

	return t.resolveContainerNameFromIDOrName(containerName)
}

func (t *DockerOrchestrator) resolveContainerNameFromIDOrName(containerNameOrID string) (string, string, error) {
	inspect, err := t.dockerInspect("container", containerNameOrID)
	if err != nil {
		return "", "", fmt.Errorf("failed to inspect container %s: %w", containerNameOrID, err)
	}

	projectName := t.soloCtx.Project.Name()

	fullContainerName := inspect.Name
	fullContainerName = strings.TrimLeft(fullContainerName, "/")
	containerName := strings.TrimPrefix(fullContainerName, projectName+"-")

	return fullContainerName, containerName, nil
}

func (t *DockerOrchestrator) dockerInspect(artifactName string, nameOrID string) (*DockerInspect, error) {
	// nolint:gosec
	composeCmd := exec.Command(t.dockerCommandPath, "inspect",
		"--format", "{{ json . }}",
		"--type", artifactName,
		nameOrID,
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

	return &inspect, nil
}
