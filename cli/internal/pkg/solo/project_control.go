package solo

import (
	"errors"
	"fmt"
	"github.com/spaulg/solo/cli/internal/pkg/solo/config"
	"github.com/spaulg/solo/cli/internal/pkg/solo/orchestrator"
	"os"
	"path"
)

type ProjectControl struct {
	Config       *config.Config
	Project      *Project
	ComposeFile  string
	Orchestrator orchestrator.Orchestrator
}

func NewProjectControl(config *config.Config, projectFile *Project) *ProjectControl {
	return &ProjectControl{
		Config:       config,
		Project:      projectFile,
		ComposeFile:  path.Join(projectFile.Directory, ".solo", "docker-compose.yml"),
		Orchestrator: orchestrator.BuildOrchestrator(),
	}
}

func (p *ProjectControl) DumpComposeConfig() {
	composeYml, _ := p.Orchestrator.ExportComposeConfiguration(p.Config, p.Project.FilePath)
	fmt.Println(string(composeYml))
}

func (p *ProjectControl) Start() {
	// Write compose file
	composeYml, _ := p.Orchestrator.ExportComposeConfiguration(p.Config, p.Project.FilePath)
	fmt.Println("composeYml exported")

	p.exportComposeFile(composeYml)

	// todo: launch provisioning grpc server
	//fmt.Println("Launching GRPC service...")
	//grpc_server := NewGrpcServer()
	//go grpc_server.Listen()

	fmt.Println("Starting orchestrator")
	if err := p.Orchestrator.Up(p.Project.Directory, p.ComposeFile); err != nil {
		fmt.Println(fmt.Errorf("error running composeCmd: %v", err))
		os.Exit(1)
	}

	fmt.Println("done")

	//fmt.Println("Sleeping...")
	//time.Sleep(30 * time.Second)

	// todo: wait for confirmation that all containers have completed provisioning
	// todo: wait delay period for final containers to start
	// todo: Exec post start commands (via docker exec)
	// todo: wait delay period for all containers to checkin for post start commands provisioning
}

func (p *ProjectControl) Stop() {
	p.composeFileExists()

	// todo: Exec pre stop commands

	if err := p.Orchestrator.Down(p.Project.Directory, p.ComposeFile); err != nil {
		fmt.Println(fmt.Errorf("error running compose: %v", err))
	}
}

func (p *ProjectControl) Destroy() {
	p.composeFileExists()

	// todo: Exec pre stop commands

	if err := p.Orchestrator.Destroy(p.Project.Directory, p.ComposeFile); err != nil {
		fmt.Println(fmt.Errorf("error running compose: %v", err))
	}

	if err := os.Remove(p.ComposeFile); err != nil {
		fmt.Println(fmt.Errorf("failed to remove compose file: %v", err))
		os.Exit(1)
	}
}

func (p *ProjectControl) exportComposeFile(composeYml []byte) {
	composeDirectory := path.Dir(p.ComposeFile)
	if _, err := os.Stat(composeDirectory); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Println(fmt.Errorf("failed to check .solo directory existence: %v", err))
			os.Exit(1)
		}

		if err := os.MkdirAll(composeDirectory, 0755); err != nil {
			fmt.Println(fmt.Errorf("failed to create .solo directory: %v", err))
			os.Exit(1)
		}
	}

	if err := os.WriteFile(p.ComposeFile, composeYml, 0640); err != nil {
		fmt.Println(fmt.Errorf("failed to write compose file: %v", err))
		os.Exit(1)
	}
}

func (p *ProjectControl) composeFileExists() {
	if _, err := os.Stat(p.ComposeFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("compose file not found")
			os.Exit(1)
		} else {
			fmt.Println(fmt.Errorf("error running composeCmd: %v", err))
			os.Exit(1)
		}
	}
}
