package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spaulg/solo/internal/pkg/impl/common/app/cmd"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/infra/container"
)

const maxOutputPreviewLen = 200

type ProjectTooling struct {
	soloCtx             *context.CliContext
	orchestratorFactory container_types.OrchestratorFactory
}

func NewProjectTooling(
	soloCtx *context.CliContext,
	orchestratorFactory container_types.OrchestratorFactory,
) *ProjectTooling {
	return &ProjectTooling{
		soloCtx:             soloCtx,
		orchestratorFactory: orchestratorFactory,
	}
}

func (t *ProjectTooling) ExecuteTool(name string, args []string) error {
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

	shell := t.soloCtx.Config.Shell.DefaultShell
	if toolConfig.Shell != nil {
		shell = *toolConfig.Shell
	}

	// Parse the initial command and args for a full path or
	// shell and split into arguments
	command, arguments := cmd.SplitCommand(shell, toolConfig.Command+" "+strings.Join(args, " "))
	workingDirectory := ""

	// If a static working directory is specified, use it
	if toolConfig.WorkingDirectory != "" {
		workingDirectory = toolConfig.WorkingDirectory
	}

	return orchestrator.ComposeForkAndExecute(toolConfig.Service, 1, command, arguments, workingDirectory)
}

func (t *ProjectTooling) ExecuteShell(shell string, index int, serviceName string) error {
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
	fullContainerName, _, err := orchestrator.ResolveContainerNameFromServiceName(serviceName, index)
	if err != nil {
		return fmt.Errorf("failed to resolve container name for service %s: %w", serviceName, err)
	}

	if shell == "" {
		// List the shells in the container
		catShellsCommand := []string{t.soloCtx.Config.Entrypoint.ContainerEntrypointPath, "cat-shells"}

		output, err := orchestrator.RunCommand(fullContainerName, catShellsCommand)
		if err != nil {
			return err
		}

		// Select a shell to use
		var shellList []string
		if err := json.Unmarshal([]byte(output), &shellList); err != nil {
			outputPreview := output

			if len(outputPreview) > maxOutputPreviewLen {
				outputPreview = outputPreview[:maxOutputPreviewLen] + "..."
			}

			return fmt.Errorf("failed to parse shell list JSON from cat-shells for container %s: %w; output preview: %q",
				fullContainerName, err, outputPreview)
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

			for _, priorityShell := range t.soloCtx.Config.Shell.ShellPriority {
				if shellPaths, ok := shellmap[priorityShell]; ok && len(shellPaths) > 0 {
					shell = path.Join(shellPaths[len(shellPaths)-1], priorityShell)
					break
				}
			}

			// If a shell from the preferred list could not be
			// selected, take the first one from the list
			if shell == "" {
				shell = shellList[0]
			}
		} else {
			shell = t.soloCtx.Config.Shell.DefaultShell
		}
	}

	return orchestrator.ForkAndExecute(fullContainerName, shell, nil, workingDirectory)
}
