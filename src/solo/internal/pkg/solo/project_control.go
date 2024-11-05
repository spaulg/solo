package solo

import (
	"errors"
	"fmt"
	"github.com/spaulg/solo/cli/internal/pkg/solo/config"
	"github.com/spaulg/solo/cli/internal/pkg/solo/grpc"
	"github.com/spaulg/solo/cli/internal/pkg/solo/orchestrator"
	"github.com/spaulg/solo/cli/internal/pkg/solo/project"
	"os"
	"path"
	"time"
)

type ProjectControl struct {
	config       *config.Config
	project      *project.Project
	composeFile  string
	orchestrator orchestrator.Orchestrator
	grpcServer   grpc.Server
}

func NewProjectControl(
	config *config.Config,
	project *project.Project,
	orchestrator orchestrator.Orchestrator,
	grpcServer grpc.Server,
) *ProjectControl {
	return &ProjectControl{
		config:       config,
		project:      project,
		composeFile:  project.ResolveStateDirectory("docker-compose.yml"),
		orchestrator: orchestrator,
		grpcServer:   grpcServer,
	}
}

func (t *ProjectControl) Start() error {
	// Start GRPC services
	if err := t.grpcServer.Start(); err != nil {
		return err
	}

	defer t.grpcServer.Stop()

	// Write compose file
	composeYml, _ := t.orchestrator.ExportComposeConfiguration(t.config, t.project)
	if err := t.exportComposeFile(composeYml); err != nil {
		return err
	}

	// Start compose services
	if err := t.orchestrator.Up(t.project.GetDirectory(), t.composeFile); err != nil {
		return fmt.Errorf("error running composeCmd: %v", err)
	}

	// todo: refactor in to timer - context object ??
	fmt.Println("Sleeping...")
	time.Sleep(30 * time.Second)

	// todo: Exec post start commands (via docker exec)
	// todo: wait delay period for all containers to checkin for post start commands provisioning

	return nil
}

func (t *ProjectControl) Stop() error {
	if err := t.composeFileExists(); err != nil {
		return err
	}

	// todo: Exec pre stop commands

	if err := t.orchestrator.Down(t.project.GetDirectory(), t.composeFile); err != nil {
		return fmt.Errorf("error running compose: %v", err)
	}

	return nil
}

func (t *ProjectControl) Destroy() error {
	if err := t.composeFileExists(); err != nil {
		return err
	}

	// todo: Exec pre stop commands

	if err := t.orchestrator.Destroy(t.project.GetDirectory(), t.composeFile); err != nil {
		return fmt.Errorf("error running compose: %v", err)
	}

	if err := os.Remove(t.composeFile); err != nil {
		return fmt.Errorf("failed to remove compose file: %v", err)
	}

	return nil
}

func (t *ProjectControl) exportComposeFile(composeYml []byte) error {
	composeDirectory := path.Dir(t.composeFile)
	if _, err := os.Stat(composeDirectory); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to check .solo directory existence: %v", err)
		}

		if err := os.MkdirAll(composeDirectory, 0755); err != nil {
			return fmt.Errorf("failed to create .solo directory: %v", err)
		}
	}

	if err := os.WriteFile(t.composeFile, composeYml, 0640); err != nil {
		return fmt.Errorf("failed to write compose file: %v", err)
	}

	return nil
}

func (t *ProjectControl) composeFileExists() error {
	if _, err := os.Stat(t.composeFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("compose file not found")
		} else {
			return fmt.Errorf("error looking for compose file: %v", err)
		}
	}

	return nil
}
