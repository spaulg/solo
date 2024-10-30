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
	Config       *config.Config
	Project      *project.Project
	ComposeFile  string
	Orchestrator orchestrator.Orchestrator
}

func NewProjectControl(config *config.Config, projectFile *project.Project) *ProjectControl {
	return &ProjectControl{
		Config:       config,
		Project:      projectFile,
		ComposeFile:  path.Join(projectFile.Directory, ".solo", "docker-compose.yml"),
		Orchestrator: orchestrator.BuildOrchestrator(),
	}
}

func (p *ProjectControl) DumpComposeConfig() error {
	composeYml, err := p.Orchestrator.ExportComposeConfiguration(p.Config, p.Project)
	if err != nil {
		return err
	}

	fmt.Println(string(composeYml))
	return nil
}

func (p *ProjectControl) Start() error {
	// Start GRPC services
	grpcServer := grpc.NewServer(
		p.Orchestrator.GetHostGatewayHostname(),
		p.Config.GrpcServerPort,
		p.Project.GetAllServicesStateDirectory(),
	)

	if err := grpcServer.Start(); err != nil {
		return err
	}

	defer grpcServer.Stop()

	// Write compose file
	composeYml, _ := p.Orchestrator.ExportComposeConfiguration(p.Config, p.Project)
	if err := p.exportComposeFile(composeYml); err != nil {
		return err
	}

	// Start compose services
	if err := p.Orchestrator.Up(p.Project.Directory, p.ComposeFile); err != nil {
		return fmt.Errorf("error running composeCmd: %v", err)
	}

	// todo: refactor in to timer - context object ??
	fmt.Println("Sleeping...")
	time.Sleep(30 * time.Second)

	// todo: Exec post start commands (via docker exec)
	// todo: wait delay period for all containers to checkin for post start commands provisioning

	return nil
}

func (p *ProjectControl) Stop() error {
	if err := p.composeFileExists(); err != nil {
		return err
	}

	// todo: Exec pre stop commands

	if err := p.Orchestrator.Down(p.Project.Directory, p.ComposeFile); err != nil {
		return fmt.Errorf("error running compose: %v", err)
	}

	return nil
}

func (p *ProjectControl) Destroy() error {
	if err := p.composeFileExists(); err != nil {
		return err
	}

	// todo: Exec pre stop commands

	if err := p.Orchestrator.Destroy(p.Project.Directory, p.ComposeFile); err != nil {
		return fmt.Errorf("error running compose: %v", err)
	}

	if err := os.Remove(p.ComposeFile); err != nil {
		return fmt.Errorf("failed to remove compose file: %v", err)
	}

	return nil
}

func (p *ProjectControl) exportComposeFile(composeYml []byte) error {
	composeDirectory := path.Dir(p.ComposeFile)
	if _, err := os.Stat(composeDirectory); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to check .solo directory existence: %v", err)
		}

		if err := os.MkdirAll(composeDirectory, 0755); err != nil {
			return fmt.Errorf("failed to create .solo directory: %v", err)
		}
	}

	if err := os.WriteFile(p.ComposeFile, composeYml, 0640); err != nil {
		return fmt.Errorf("failed to write compose file: %v", err)
	}

	return nil
}

func (p *ProjectControl) composeFileExists() error {
	if _, err := os.Stat(p.ComposeFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("compose file not found")
		} else {
			return fmt.Errorf("error looking for compose file: %v", err)
		}
	}

	return nil
}
