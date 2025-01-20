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
	soloCtx           *context.SoloContext
	workflowManager   events.Manager
	orchestrator      container.Orchestrator
	grpcServerFactory grpc.ServerFactory
}

func NewProjectControl(
	soloCtx *context.SoloContext,
	workflowManager events.Manager,
	orchestrator container.Orchestrator,
	grpcServerFactory grpc.ServerFactory,
) *ProjectControl {
	return &ProjectControl{
		soloCtx:           soloCtx,
		workflowManager:   workflowManager,
		orchestrator:      orchestrator,
		grpcServerFactory: grpcServerFactory,
	}
}

func (t *ProjectControl) Start() error {
	grpcServer, err := t.grpcServerFactory.Build(
		t.orchestrator.GetHostGatewayHostname(),
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
	composeYml, _ := t.orchestrator.ExportComposeConfiguration(t.soloCtx.Config, t.soloCtx.Project)
	if err := t.exportComposeFile(composeYml); err != nil {
		return err
	}

	// Start compose services
	if err := t.orchestrator.Up(t.soloCtx.Project.GetDirectory(), t.soloCtx.Project.GetGeneratedComposeFilePath()); err != nil {
		return fmt.Errorf("error running composeCmd: %v", err)
	}

	// todo: only wait for Build if not previously started
	if err := guard.WaitForCompletion(workflowcommon.Build); err != nil {
		return err
	}

	if err := guard.WaitForCompletion(workflowcommon.PreStart); err != nil {
		return err
	}

	// todo: Exec post start commands (via docker exec)
	// todo: wait delay period for all containers to checkin for post start commands provisioning

	return nil
}

func (t *ProjectControl) Stop() error {
	if err := t.composeFileExists(); err != nil {
		return err
	}

	// todo: Exec pre stop commands

	if err := t.orchestrator.Down(t.soloCtx.Project.GetDirectory(), t.soloCtx.Project.GetGeneratedComposeFilePath()); err != nil {
		return fmt.Errorf("error running compose: %v", err)
	}

	return nil
}

func (t *ProjectControl) Destroy() error {
	if err := t.composeFileExists(); err != nil {
		return err
	}

	// todo: Exec pre stop commands

	if err := t.orchestrator.Destroy(t.soloCtx.Project.GetDirectory(), t.soloCtx.Project.GetGeneratedComposeFilePath()); err != nil {
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

func (t *ProjectControl) composeFileExists() error {
	if _, err := os.Stat(t.soloCtx.Project.GetGeneratedComposeFilePath()); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("compose file not found")
		} else {
			return fmt.Errorf("error looking for compose file: %v", err)
		}
	}

	return nil
}

func (t *ProjectControl) buildWorkflowServiceMap(workflowNames []workflowcommon.Name) (WorkflowServiceMap, error) {
	serviceNames := t.soloCtx.Project.ServiceNames()
	workflowMap := make(WorkflowServiceMap)

	for _, workflowName := range workflowNames {
		workflowMap[workflowName] = serviceNames
	}

	return workflowMap, nil
}
