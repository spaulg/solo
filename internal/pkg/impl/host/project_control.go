package host

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/spaulg/solo/internal/pkg/impl/common/cmd"
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/events"
	grpc_types "github.com/spaulg/solo/internal/pkg/types/host/grpc"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type ProjectControl struct {
	soloCtx              *context.CliContext
	workflowManager      events_types.Manager
	orchestratorFactory  container_types.OrchestratorFactory
	grpcServerFactory    grpc_types.ServerFactory
	workflowGuardFactory wms_types.WorkflowGuardFactory
}

func NewProjectControl(
	soloCtx *context.CliContext,
	workflowManager events_types.Manager,
	orchestratorFactory container_types.OrchestratorFactory,
	grpcServerFactory grpc_types.ServerFactory,
	workflowGuardFactory wms_types.WorkflowGuardFactory,
) *ProjectControl {
	return &ProjectControl{
		soloCtx:              soloCtx,
		workflowManager:      workflowManager,
		orchestratorFactory:  orchestratorFactory,
		grpcServerFactory:    grpcServerFactory,
		workflowGuardFactory: workflowGuardFactory,
	}
}

func (t *ProjectControl) Start() error {
	orchestrator, err := t.orchestratorFactory.Build()
	if err != nil {
		return fmt.Errorf("failed to build orchestrator: %w", err)
	}

	// Write compose file
	if exists, _ := t.composeFileExists(); !exists {
		t.soloCtx.Logger.Info("Generating compose file")
		composeYml, _ := orchestrator.ExportComposeConfiguration(t.soloCtx.Config, t.soloCtx.Project)
		if err := t.exportComposeFile(composeYml); err != nil {
			return err
		}
	}

	serviceStatus, err := orchestrator.ServicesStatus(nil)
	if err != nil {
		return fmt.Errorf("failed to check service status: %w", err)
	}

	if len(serviceStatus.NotRunningServices) == 0 {
		t.soloCtx.Logger.Info("All required services already running")
		return nil
	}

	// Start GRPC services
	grpcServer, err := t.grpcServerFactory.Build(
		orchestrator,
		t.soloCtx.Project,
		t.soloCtx.Config.GrpcServerPort,
	)

	if err != nil {
		return fmt.Errorf("failed to build GRPC server: %w", err)
	}

	if err := grpcServer.Start(); err != nil {
		return fmt.Errorf("failed to start GRPC server: %w", err)
	}

	defer grpcServer.Stop()

	if err := t.copyEntrypointToState(); err != nil {
		return fmt.Errorf("failed to copy entrypoint to state directory: %w", err)
	}

	// Populate a list of container names that will be started
	containerNames, err := t.soloCtx.Project.Services().ContainerNames(serviceStatus.NotRunningServices)
	if err != nil {
		return fmt.Errorf("failed to convert service names to container names: %w", err)
	}

	workflowNames := []workflowcommon.WorkflowName{
		workflowcommon.FirstPreStartContainer,
		workflowcommon.FirstPreStartService,
		workflowcommon.PreStartContainer,
		workflowcommon.PreStartService,
		workflowcommon.PostStartContainer,
		workflowcommon.PostStartService,
		workflowcommon.FirstPostStartContainer,
		workflowcommon.FirstPostStartService,
	}

	guard := t.workflowGuardFactory.Build(workflowNames, containerNames)
	t.workflowManager.Subscribe(guard)
	defer t.workflowManager.Unsubscribe(guard)

	// Start compose services
	if err := orchestrator.ComposeUp(t.soloCtx.Project.Services().ServiceNames()); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	if err := guard.Wait(func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error {
		if err := guardCallback(workflowcommon.FirstPreStartContainer); err != nil {
			return err
		}

		if err := guardCallback(workflowcommon.FirstPreStartService); err != nil {
			return err
		}

		if err := guardCallback(workflowcommon.PreStartContainer); err != nil {
			return err
		}

		if err := guardCallback(workflowcommon.PreStartService); err != nil {
			return err
		}

		// Exec post start commands
		postStartEvents := []workflowcommon.WorkflowName{
			workflowcommon.PostStartContainer,
			workflowcommon.PostStartService,
			workflowcommon.FirstPostStartContainer,
			workflowcommon.FirstPostStartService,
		}

		for _, event := range postStartEvents {
			postStartCommand := []string{
				t.soloCtx.Config.Entrypoint.ContainerEntrypointPath,
				"trigger-event",
				event.String(),
			}

			if err := orchestrator.StartCommand(container, postStartCommand); err != nil {
				return fmt.Errorf("error running compose: %w", err)
			}

			if err := guardCallback(event); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("error waiting for services to complete workflows: %w", err)
	}

	// Wait for all events to be delivered
	t.soloCtx.Logger.Debug("Waiting for all remaining events to be delivered")
	t.workflowManager.Wait()

	t.soloCtx.Logger.Debug("Finished starting all services successfully")
	return nil
}

func (t *ProjectControl) Stop() error {
	if exists, err := t.composeFileExists(); !exists || err != nil {
		return err
	}

	// Build orchestrator
	orchestrator, err := t.orchestratorFactory.Build()
	if err != nil {
		return fmt.Errorf("failed to build orchestrator: %w", err)
	}

	// Build workflow service map
	serviceStatus, err := orchestrator.ServicesStatus(nil)
	if err != nil {
		return fmt.Errorf("failed to check service status: %w", err)
	}

	if len(serviceStatus.RunningServices) > 0 {
		grpcServer, err := t.grpcServerFactory.Build(
			orchestrator,
			t.soloCtx.Project,
			t.soloCtx.Config.GrpcServerPort,
		)

		if err != nil {
			return fmt.Errorf("failed to build GRPC server: %w", err)
		}

		// Start GRPC services
		if err := grpcServer.Start(); err != nil {
			return fmt.Errorf("failed to start GRPC server: %w", err)
		}

		defer grpcServer.Stop()

		// Populate a list of container names that will be stopped
		servicesToStop := serviceStatus.RunningServices
		containerNames, err := t.soloCtx.Project.Services().ContainerNames(servicesToStop)
		if err != nil {
			return fmt.Errorf("failed to convert service names to container names: %w", err)
		}

		workflowNames := []workflowcommon.WorkflowName{
			workflowcommon.PreStopContainer,
		}

		guard := t.workflowGuardFactory.Build(workflowNames, containerNames)
		t.workflowManager.Subscribe(guard)
		defer t.workflowManager.Unsubscribe(guard)

		if err := guard.Wait(func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error {
			// Exec pre stop commands
			preStopCommand := []string{
				t.soloCtx.Config.Entrypoint.ContainerEntrypointPath,
				"trigger-event",
				workflowcommon.PreStopContainer.String(),
			}

			if err := orchestrator.StartCommand(container, preStopCommand); err != nil {
				return fmt.Errorf("error running compose: %w", err)
			}

			return nil
		}); err != nil {
			return fmt.Errorf("error waiting for services to complete workflows: %w", err)
		}
	}

	if err := orchestrator.ComposeStop(t.soloCtx.Project.Services().ExclusiveServiceNames()); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	// Wait for all events to be delivered
	t.workflowManager.Wait()

	return nil
}

func (t *ProjectControl) Destroy() error {
	if exists, err := t.composeFileExists(); !exists || err != nil {
		return nil
	}

	// Build orchestrator
	orchestrator, err := t.orchestratorFactory.Build()
	if err != nil {
		return fmt.Errorf("failed to build orchestrator: %w", err)
	}

	// Build workflow service map
	serviceStatus, err := orchestrator.ServicesStatus(nil)
	if err != nil {
		return fmt.Errorf("failed to check service status: %w", err)
	}

	if len(serviceStatus.RunningServices) > 0 {
		grpcServer, err := t.grpcServerFactory.Build(
			orchestrator,
			t.soloCtx.Project,
			t.soloCtx.Config.GrpcServerPort,
		)

		if err != nil {
			return fmt.Errorf("failed to build GRPC server: %w", err)
		}

		// Start GRPC services
		if err := grpcServer.Start(); err != nil {
			return fmt.Errorf("failed to start GRPC server: %w", err)
		}

		defer grpcServer.Stop()

		// Populate a list of container names that will be destroyed
		servicesToDestroy := serviceStatus.RunningServices
		containerNames, err := t.soloCtx.Project.Services().ContainerNames(servicesToDestroy)
		if err != nil {
			return fmt.Errorf("failed to convert service names to container names: %w", err)
		}

		workflowNames := []workflowcommon.WorkflowName{
			workflowcommon.PreDestroyContainer,
		}

		guard := t.workflowGuardFactory.Build(workflowNames, containerNames)
		t.workflowManager.Subscribe(guard)
		defer t.workflowManager.Unsubscribe(guard)

		if err := guard.Wait(func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error {
			// Exec pre destroy commands
			preDestroyCommand := []string{
				t.soloCtx.Config.Entrypoint.ContainerEntrypointPath,
				"trigger-event",
				workflowcommon.PreDestroyContainer.String(),
			}

			if err := orchestrator.StartCommand(container, preDestroyCommand); err != nil {
				return fmt.Errorf("error running compose: %w", err)
			}

			return nil
		}); err != nil {
			return fmt.Errorf("error waiting for services to complete workflows: %w", err)
		}
	}

	if err := orchestrator.ComposeDown(t.soloCtx.Project.Services().ExclusiveServiceNames()); err != nil {
		return fmt.Errorf("failed to destroy services: %w", err)
	}

	// Wait for all events to be delivered
	t.workflowManager.Wait()

	return nil
}

func (t *ProjectControl) Clean(purgeStateDirectory bool) error {
	var purgeDirectoryList []string

	if purgeStateDirectory {
		// Purge the entire state directory
		purgeDirectoryList = []string{t.soloCtx.Project.GetStateDirectoryRoot()}
	} else {
		// Purge only certain sub folders
		purgeDirectoryList = []string{
			t.soloCtx.Project.GetGeneratedComposeFilePath(),
			t.soloCtx.Project.GetAllServicesStateDirectory(),
			t.soloCtx.Project.GetServiceStateDirectoryRoot(),
		}
	}

	for _, purgeDirectory := range purgeDirectoryList {
		if err := os.RemoveAll(purgeDirectory); err != nil {
			return fmt.Errorf("failed to remove state directory %s: %w", purgeDirectory, err)
		}
	}

	return nil
}

func (t *ProjectControl) Rebuild() error {
	// Build orchestrator
	orchestrator, err := t.orchestratorFactory.Build()
	if err != nil {
		return fmt.Errorf("failed to build orchestrator: %w", err)
	}

	// Build workflow service map
	serviceStatus, err := orchestrator.ServicesStatus(nil)
	if err != nil {
		return fmt.Errorf("failed to check service status: %w", err)
	}

	profiles, err := t.soloCtx.Project.Services().ProfilesOfServices(serviceStatus.RunningServices)
	if err != nil {
		return fmt.Errorf("failed to get profiles of services: %w", err)
	}

	if err := t.Destroy(); err != nil {
		return err
	}

	if err := t.Clean(false); err != nil {
		return err
	}

	if err := t.soloCtx.Project.ReloadWithProfiles(profiles); err != nil {
		return err
	}

	return t.Start()
}

func (t *ProjectControl) ExecuteTool(name string, args []string) error {
	// Build orchestrator
	orchestrator, err := t.orchestratorFactory.Build()
	if err != nil {
		return fmt.Errorf("failed to build orchestrator: %w", err)
	}

	tools := t.soloCtx.Project.Tools()
	toolConfig, ok := tools[name]

	if !ok {
		return fmt.Errorf("tool %s not found in project configuration", name)
	}

	// Validate service exists in the project
	if !t.soloCtx.Project.Services().HasService(toolConfig.Service) {
		return fmt.Errorf("service %s not found in project configuration", toolConfig.Service)
	}

	// Parse the initial command and args for a full path or
	// shell and split into arguments
	command, arguments := cmd.SplitCommand(toolConfig.Command + " " + strings.Join(args, " "))
	workingDirectory := ""

	// If a static working directory is specified, use it
	if toolConfig.WorkingDirectory != "" {
		workingDirectory = toolConfig.WorkingDirectory
	}

	return orchestrator.ComposeForkAndExecute(toolConfig.Service, 1, command, arguments, workingDirectory)
}

func (t *ProjectControl) ExecuteShell(shell string, index int, serviceName string) error {
	// Build orchestrator
	orchestrator, err := t.orchestratorFactory.Build()
	if err != nil {
		return fmt.Errorf("failed to build orchestrator: %w", err)
	}

	// Validate service exists in the project
	if !t.soloCtx.Project.Services().HasService(serviceName) {
		return fmt.Errorf("service %s not found in project configuration", serviceName)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	workingDirectory := t.soloCtx.Project.Services().GetService(serviceName).ResolveContainerWorkingDirectory(cwd)

	// Get the desired container name
	containerName, err := orchestrator.ResolveContainerNameFromServiceName(serviceName, index)
	if err != nil {
		return fmt.Errorf("failed to resolve container name for service %s: %w", serviceName, err)
	}

	if shell == "" {
		// List the shells in the container
		catShellsCommand := []string{t.soloCtx.Config.Entrypoint.ContainerEntrypointPath, "cat-shells"}

		output, err := orchestrator.RunCommand(containerName, catShellsCommand)
		if err != nil {
			return err
		}

		// Select a shell to use
		var shellList []string
		if err := json.Unmarshal([]byte(output), &shellList); err != nil {
			return err
		}

		if len(shellList) > 0 {
			shellmap := make(map[string][]string)

			for _, fullShellPath := range shellList {
				shellPath, shellFile := path.Split(fullShellPath)
				if _, ok := shellmap[shellFile]; !ok {
					shellmap[shellFile] = make([]string, 0)
				}

				shellmap[shellFile] = append(shellmap[shellFile], shellPath)
			}

			for _, priorityShell := range t.soloCtx.Config.ShellPriority {
				if shellList, ok := shellmap[priorityShell]; ok && len(shellList) > 0 {
					shell = path.Join(shellList[len(shellList)-1], priorityShell)
					break
				}
			}

			// If a shell from the preferred list could not be
			// selected, take the first one from the list
			if shell == "" {
				shell = shellList[0]
			}
		} else {
			shell = t.soloCtx.Config.DefaultShell
		}
	}

	return orchestrator.ForkAndExecute(containerName, shell, nil, workingDirectory)
}

func (t *ProjectControl) exportComposeFile(composeYml []byte) error {
	composeDirectory := path.Dir(t.soloCtx.Project.GetGeneratedComposeFilePath())
	if _, err := os.Stat(composeDirectory); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to check .solo directory existence: %w", err)
		}

		if err := os.MkdirAll(composeDirectory, 0755); err != nil {
			return fmt.Errorf("failed to create .solo directory: %w", err)
		}
	}

	if err := os.WriteFile(t.soloCtx.Project.GetGeneratedComposeFilePath(), composeYml, 0640); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	return nil
}

func (t *ProjectControl) composeFileExists() (bool, error) {
	if _, err := os.Stat(t.soloCtx.Project.GetGeneratedComposeFilePath()); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		} else {
			return false, fmt.Errorf("error looking for compose file: %w", err)
		}
	}

	return true, nil
}

func (t *ProjectControl) copyEntrypointToState() error {
	src := t.soloCtx.Config.Entrypoint.HostEntrypointPath
	dst := path.Join(t.soloCtx.Project.GetStateDirectoryRoot(), "solo-entrypoint")

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}

	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	if err := os.Chmod(dst, 0755); err != nil {
		return err
	}

	return nil
}
