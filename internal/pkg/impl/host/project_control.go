package host

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

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

	serviceStatus, err := orchestrator.ServicesStatus()
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
	containerNames, err := t.soloCtx.Project.ContainerNames(serviceStatus.NotRunningServices)
	if err != nil {
		return fmt.Errorf("failed to convert service names to container names: %w", err)
	}

	workflowNames := []workflowcommon.WorkflowName{
		workflowcommon.FirstPreStart,
		workflowcommon.PreStart,
		workflowcommon.PostStart,
	}

	guard := t.workflowGuardFactory.Build(workflowNames, containerNames)
	t.workflowManager.Subscribe(guard)
	defer t.workflowManager.Unsubscribe(guard)

	// Start compose services
	if err := orchestrator.ComposeUp(); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	if err := guard.Wait(func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error {
		if err := guardCallback(workflowcommon.FirstPreStart); err != nil {
			return err
		}

		if err := guardCallback(workflowcommon.PreStart); err != nil {
			return err
		}

		// Exec post start commands
		postStartCommand := []string{t.soloCtx.Config.Entrypoint.ContainerEntrypointPath, "trigger-event", "post_start"}

		if err := orchestrator.Execute(container, postStartCommand); err != nil {
			return fmt.Errorf("error running compose: %v", err)
		}

		if err := guardCallback(workflowcommon.PostStart); err != nil {
			return err
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
	serviceStatus, err := orchestrator.ServicesStatus()
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
		containerNames, err := t.soloCtx.Project.ContainerNames(servicesToStop)
		if err != nil {
			return fmt.Errorf("failed to convert service names to container names: %w", err)
		}

		workflowNames := []workflowcommon.WorkflowName{
			workflowcommon.PreStop,
		}

		guard := t.workflowGuardFactory.Build(workflowNames, containerNames)
		t.workflowManager.Subscribe(guard)
		defer t.workflowManager.Unsubscribe(guard)

		if err := guard.Wait(func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error {
			// Exec pre stop commands
			preStopCommand := []string{t.soloCtx.Config.Entrypoint.ContainerEntrypointPath, "trigger-event", "pre_stop"}

			if err := orchestrator.Execute(container, preStopCommand); err != nil {
				return fmt.Errorf("error running compose: %v", err)
			}

			return nil
		}); err != nil {
			return fmt.Errorf("error waiting for services to complete workflows: %w", err)
		}
	}

	if err := orchestrator.ComposeStop(); err != nil {
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
	serviceStatus, err := orchestrator.ServicesStatus()
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
		containerNames, err := t.soloCtx.Project.ContainerNames(servicesToDestroy)
		if err != nil {
			return fmt.Errorf("failed to convert service names to container names: %w", err)
		}

		workflowNames := []workflowcommon.WorkflowName{
			workflowcommon.PreDestroy,
		}

		guard := t.workflowGuardFactory.Build(workflowNames, containerNames)
		t.workflowManager.Subscribe(guard)
		defer t.workflowManager.Unsubscribe(guard)

		if err := guard.Wait(func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error {
			// Exec pre destroy commands
			preDestroyCommand := []string{t.soloCtx.Config.Entrypoint.ContainerEntrypointPath, "trigger-event", "pre_destroy"}

			if err := orchestrator.Execute(container, preDestroyCommand); err != nil {
				return fmt.Errorf("error running compose: %v", err)
			}

			return nil
		}); err != nil {
			return fmt.Errorf("error waiting for services to complete workflows: %w", err)
		}
	}

	if err := orchestrator.ComposeDown(); err != nil {
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
			return fmt.Errorf("failed to remove state directory %s: %v", purgeDirectory, err)
		}
	}

	return nil
}

func (t *ProjectControl) exportComposeFile(composeYml []byte) error {
	composeDirectory := path.Dir(t.soloCtx.Project.GetGeneratedComposeFilePath())
	if _, err := os.Stat(composeDirectory); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to check .solo directory existence: %v", err)
		}

		if err := os.MkdirAll(composeDirectory, 0755); err != nil {
			return fmt.Errorf("failed to create .solo directory: %v", err)
		}
	}

	if err := os.WriteFile(t.soloCtx.Project.GetGeneratedComposeFilePath(), composeYml, 0640); err != nil {
		return fmt.Errorf("failed to write compose file: %v", err)
	}

	return nil
}

func (t *ProjectControl) composeFileExists() (bool, error) {
	if _, err := os.Stat(t.soloCtx.Project.GetGeneratedComposeFilePath()); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		} else {
			return false, fmt.Errorf("error looking for compose file: %v", err)
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
