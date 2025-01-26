package solo

import (
	"errors"
	"fmt"
	workflowcommon "github.com/spaulg/solo/internal/pkg/common/wms"
	"github.com/spaulg/solo/internal/pkg/solo/container"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/grpc"
	"os"
	"path"
)

type ProjectControl struct {
	soloCtx             *context.SoloContext
	workflowManager     events.Manager
	orchestratorFactory container.OrchestratorFactory
	grpcServerFactory   grpc.ServerFactory
}

func NewProjectControl(
	soloCtx *context.SoloContext,
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
	workflowMap, err := t.buildWorkflowServiceMap([]workflowcommon.Name{
		workflowcommon.Build,
		workflowcommon.PreStart,
		workflowcommon.PostStart,
	})

	if err != nil {
		return err
	}

	// Register workflow guard
	guard := NewProjectWorkflowGuard(t.soloCtx, workflowMap)
	t.workflowManager.Subscribe(guard)
	defer t.workflowManager.Unsubscribe(guard)

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

	if err := guard.WaitForCompletion(workflowcommon.Build); err != nil {
		return err
	}

	if err := guard.WaitForCompletion(workflowcommon.PreStart); err != nil {
		return err
	}

	// Exec post start commands
	postStartCommand := []string{"/usr/local/sbin/solo", "trigger-event", "post_start"}
	serviceNames := t.soloCtx.Project.ServiceNames()

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

	// todo: Exec pre stop commands
	orchestrator := t.orchestratorFactory.Build(t.soloCtx)
	if err := orchestrator.Down(); err != nil {
		return fmt.Errorf("error running compose: %v", err)
	}

	return nil
}

func (t *ProjectControl) Destroy() error {
	if exists, err := t.composeFileExists(); !exists || err != nil {
		return nil
	}

	// todo: Exec pre stop commands
	orchestrator := t.orchestratorFactory.Build(t.soloCtx)
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

func (t *ProjectControl) buildWorkflowServiceMap(workflowNames []workflowcommon.Name) (WorkflowServiceMap, error) {
	serviceNames := t.soloCtx.Project.ServiceNames()
	workflowMap := make(WorkflowServiceMap)

	for _, workflowName := range workflowNames {
		if workflowName == workflowcommon.Build {
			servicesToBuild := t.soloCtx.Project.ServicesPendingBuildWorkflow()

			if len(servicesToBuild) > 0 {
				workflowMap[workflowName] = servicesToBuild
			}
		} else {
			workflowMap[workflowName] = serviceNames
		}
	}

	return workflowMap, nil
}
