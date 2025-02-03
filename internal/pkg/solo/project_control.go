package solo

import (
	"errors"
	"fmt"
	workflowcommon "github.com/spaulg/solo/internal/pkg/common/wms"
	"github.com/spaulg/solo/internal/pkg/solo/container"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/grpc"
	"io"
	"os"
	"path"
)

type ProjectControl struct {
	soloCtx             *context.CliContext
	workflowManager     events.Manager
	orchestratorFactory container.OrchestratorFactory
	grpcServerFactory   grpc.ServerFactory
}

func NewProjectControl(
	soloCtx *context.CliContext,
	workflowManager events.Manager,
	orchestratorFactory container.OrchestratorFactory,
	grpcServerFactory grpc.ServerFactory,
) *ProjectControl {
	return &ProjectControl{
		soloCtx:             soloCtx,
		workflowManager:     workflowManager,
		orchestratorFactory: orchestratorFactory,
		grpcServerFactory:   grpcServerFactory,
	}
}

func (t *ProjectControl) Start() error {
	orchestrator := t.orchestratorFactory.Build(t.soloCtx)

	grpcServer, err := t.grpcServerFactory.Build(
		orchestrator.GetHostGatewayHostname(),
		t.soloCtx.Config.GrpcServerPort,
		t.soloCtx.Project,
	)

	if err != nil {
		return err
	}

	// Start GRPC services
	if err := grpcServer.Start(); err != nil {
		return err
	}

	defer grpcServer.Stop()

	// Build workflow service map
	serviceNames := t.soloCtx.Project.ServiceNames()
	workflowMap, err := t.buildWorkflowServiceMap(serviceNames, []workflowcommon.Name{
		workflowcommon.FirstPreStart,
		workflowcommon.PreStart,
		workflowcommon.PostStart,
	})

	if err != nil {
		return err
	}

	// Register workflow guard
	guard := NewProjectWorkflowGuard(t.soloCtx, workflowMap)
	guard.Subscribe(t.workflowManager)
	defer t.workflowManager.Unsubscribe(guard)

	if err := t.copyEntrypointToState(); err != nil {
		return err
	}

	// Write compose file
	if exists, _ := t.composeFileExists(); !exists {
		composeYml, _ := orchestrator.ExportComposeConfiguration(t.soloCtx.Config, t.soloCtx.Project)
		if err := t.exportComposeFile(composeYml); err != nil {
			return err
		}
	}

	// Start compose services
	if err := orchestrator.Up(); err != nil {
		return fmt.Errorf("error running compose: %v", err)
	}

	if err := guard.WaitForCompletion(workflowcommon.FirstPreStart); err != nil {
		return err
	}

	if err := guard.WaitForCompletion(workflowcommon.PreStart); err != nil {
		return err
	}

	// Exec post start commands
	postStartCommand := []string{t.soloCtx.Config.Entrypoint.ContainerEntrypointPath, "trigger-event", "post_start"}

	if err := orchestrator.Execute(serviceNames, postStartCommand); err != nil {
		return fmt.Errorf("error running compose: %v", err)
	}

	if err := guard.WaitForCompletion(workflowcommon.PostStart); err != nil {
		return err
	}

	return nil
}

func (t *ProjectControl) Stop() error {
	if exists, err := t.composeFileExists(); !exists || err != nil {
		return err
	}

	orchestrator := t.orchestratorFactory.Build(t.soloCtx)

	// Build workflow service map
	serviceNames, err := orchestrator.RunningServices()
	if err != nil {
		return err
	}

	if len(serviceNames) > 0 {
		grpcServer, err := t.grpcServerFactory.Build(
			orchestrator.GetHostGatewayHostname(),
			t.soloCtx.Config.GrpcServerPort,
			t.soloCtx.Project,
		)

		if err != nil {
			return err
		}

		// Start GRPC services
		if err := grpcServer.Start(); err != nil {
			return err
		}

		defer grpcServer.Stop()

		workflowMap, err := t.buildWorkflowServiceMap(serviceNames, []workflowcommon.Name{
			workflowcommon.PreStop,
		})

		if err != nil {
			return err
		}

		// Register workflow guard
		guard := NewProjectWorkflowGuard(t.soloCtx, workflowMap)
		guard.Subscribe(t.workflowManager)
		defer t.workflowManager.Unsubscribe(guard)

		// Exec pre stop commands
		preStopCommand := []string{t.soloCtx.Config.Entrypoint.ContainerEntrypointPath, "trigger-event", "pre_stop"}

		if err := orchestrator.Execute(serviceNames, preStopCommand); err != nil {
			return fmt.Errorf("error running compose: %v", err)
		}

		if err := guard.WaitForCompletion(workflowcommon.PreStop); err != nil {
			return err
		}
	}

	if err := orchestrator.Down(); err != nil {
		return fmt.Errorf("error running compose: %v", err)
	}

	return nil
}

func (t *ProjectControl) Destroy() error {
	if exists, err := t.composeFileExists(); !exists || err != nil {
		return nil
	}

	orchestrator := t.orchestratorFactory.Build(t.soloCtx)

	// Build workflow service map
	serviceNames, err := orchestrator.RunningServices()
	if err != nil {
		return err
	}

	if len(serviceNames) > 0 {
		grpcServer, err := t.grpcServerFactory.Build(
			orchestrator.GetHostGatewayHostname(),
			t.soloCtx.Config.GrpcServerPort,
			t.soloCtx.Project,
		)

		if err != nil {
			return err
		}

		// Start GRPC services
		if err := grpcServer.Start(); err != nil {
			return err
		}

		defer grpcServer.Stop()

		workflowMap, err := t.buildWorkflowServiceMap(serviceNames, []workflowcommon.Name{
			workflowcommon.PreDestroy,
		})

		if err != nil {
			return err
		}

		// Register workflow guard
		guard := NewProjectWorkflowGuard(t.soloCtx, workflowMap)
		guard.Subscribe(t.workflowManager)
		defer t.workflowManager.Unsubscribe(guard)

		// Exec pre destroy commands
		preDestroyCommand := []string{t.soloCtx.Config.Entrypoint.ContainerEntrypointPath, "trigger-event", "pre_destroy"}

		if err := orchestrator.Execute(serviceNames, preDestroyCommand); err != nil {
			return fmt.Errorf("error running compose: %v", err)
		}

		if err := guard.WaitForCompletion(workflowcommon.PreDestroy); err != nil {
			return err
		}
	}

	if err := orchestrator.Destroy(); err != nil {
		return fmt.Errorf("error running compose: %v", err)
	}

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

func (t *ProjectControl) buildWorkflowServiceMap(serviceNames []string, workflowNames []workflowcommon.Name) (WorkflowServiceMap, error) {
	workflowMap := make(WorkflowServiceMap)

	for _, workflowName := range workflowNames {
		if workflowName == workflowcommon.FirstPreStart {
			servicesToBuild := t.soloCtx.Project.ServicesPendingFirstPreStartWorkflow()

			if len(servicesToBuild) > 0 {
				workflowMap[workflowName] = servicesToBuild
			}
		} else {
			workflowMap[workflowName] = serviceNames
		}
	}

	return workflowMap, nil
}

func (t *ProjectControl) copyEntrypointToState() error {
	src := t.soloCtx.Config.Entrypoint.HostEntrypointPath
	dst := path.Join(t.soloCtx.Project.GetStateDirectoryRoot(), "solo-entrypoint")

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	sourceFile.Close()
	destFile.Close()

	if err := os.Chmod(dst, 0755); err != nil {
		return err
	}

	return nil
}
