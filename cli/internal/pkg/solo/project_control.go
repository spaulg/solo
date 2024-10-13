package solo

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
)

func NewProjectControl(config *Config, projectFile *Project) *ProjectControl {
	projectConfig, err := projectFile.Marshall()
	if err != nil {
		fmt.Println(fmt.Errorf("failed to read project file: %v", err))
		os.Exit(1)
	}

	return &ProjectControl{
		Config:      config,
		ProjectFile: projectFile,
		ComposeFile: path.Join(projectFile.Directory, ".solo", "docker-compose.yml"),
		Project:     projectConfig,
	}
}

func (p ProjectControl) DumpComposeConfig() {
	composeYml, _ := ExportComposeConfiguration(p.Config, p.ProjectFile)
	fmt.Println(string(composeYml))
}

func (p ProjectControl) Start() {
	// Write compose file
	composeYml, _ := ExportComposeConfiguration(p.Config, p.ProjectFile)

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

	// todo: launch provisioning grpc server
	//fmt.Println("Launching GRPC service...")
	//grpc_server := NewGrpcServer()
	//go grpc_server.Listen()

	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", p.ComposeFile,
		"--project-directory", p.ProjectFile.Directory,
		"up", "-d")

	if err := composeCmd.Run(); err != nil {
		fmt.Println(fmt.Errorf("error running composeCmd: %v", err))
		os.Exit(1)
	}

	//fmt.Println("Sleeping...")
	//time.Sleep(30 * time.Second)

	// todo: wait for confirmation that all containers have completed provisioning
	// todo: wait delay period for final containers to start
	// todo: Exec post start commands (via docker exec)
	// todo: wait delay period for all containers to checkin for post start commands provisioning
}

func (p ProjectControl) Stop() {
	if _, err := os.Stat(p.ComposeFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("compose file not found")
			os.Exit(1)
		} else {
			fmt.Println(fmt.Errorf("error running composeCmd: %v", err))
			os.Exit(1)
		}
	}

	// todo: Exec pre stop commands

	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", p.ComposeFile,
		"--project-directory", p.ProjectFile.Directory,
		"stop")

	if err := composeCmd.Run(); err != nil {
		fmt.Println(fmt.Errorf("error running compose: %v", err))
	}
}

func (p ProjectControl) Destroy() {
	if _, err := os.Stat(p.ComposeFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("compose file not found")
			os.Exit(1)
		} else {
			fmt.Println(fmt.Errorf("error running composeCmd: %v", err))
			os.Exit(1)
		}
	}

	// todo: Exec pre stop commands

	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", p.ComposeFile,
		"--project-directory", p.ProjectFile.Directory,
		"down", "-v")

	if err := composeCmd.Run(); err != nil {
		fmt.Println(fmt.Errorf("error running compose: %v", err))
	}

	if err := os.Remove(p.ComposeFile); err != nil {
		fmt.Println(fmt.Errorf("failed to remove compose file: %v", err))
		os.Exit(1)
	}
}
