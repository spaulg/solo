package project

import (
	"errors"
	"fmt"
	"github.com/spaulg/solo/cli/internal/pkg/compose_exporter"
	"github.com/spaulg/solo/cli/internal/pkg/config"
	"github.com/spaulg/solo/cli/internal/pkg/project_file"
	"os"
	"os/exec"
	"path"
)

func LoadProject(config *config.Config, projectFile *project_file.ProjectFile) *Project {
	projectConfig, err := projectFile.Marshall()
	if err != nil {
		fmt.Println(fmt.Errorf("failed to read project file: %v", err))
		os.Exit(1)
	}

	return &Project{
		Config:      config,
		ProjectFile: projectFile,
		ComposeFile: path.Join(projectFile.Directory, ".solo", "docker-compose.yml"),
		Project:     projectConfig,
	}
}

func (p Project) DumpComposeConfig() {
	composeYml, _ := compose_exporter.ExportComposeConfiguration(p.Config, p.ProjectFile)
	fmt.Println(string(composeYml))
}

func (p Project) Start() {
	// Write compose file
	composeYml, _ := compose_exporter.ExportComposeConfiguration(p.Config, p.ProjectFile)

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

	// todo: Extract yaml steps file

	composeCmd := exec.Command("/usr/local/bin/docker", "compose",
		"-f", p.ComposeFile,
		"--project-directory", p.ProjectFile.Directory,
		"up", "-d")

	if err := composeCmd.Run(); err != nil {
		fmt.Println(fmt.Errorf("error running composeCmd: %v", err))
		os.Exit(1)
	}

	// todo: wait for lock file / health check from service to notify startup succeeded
	//		 could this be done via a compose event stream?

	// todo: Exec post start commands (via docker exec)
	//	     or could I make the entrypoint an agent that accepts remote connections over a named pipe
	//		 that I could feed instruction to
	//		 or start a grpc service in this command, pass the port used to the guest and wait for a connection
	//		 then feed back instruction
	//		 would this be a long running process? would it be passed via env var
	//		 or a file that can be updated
	//		 would the guest run the agent forever or just initially and then exit?
	//		 would stop/destroy commands restart the agent and reconnect for instruction to receive events again
	//		 what if the container is already stopped when this happens?
}

func (p Project) Stop() {
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

func (p Project) Destroy() {
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
