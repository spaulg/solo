package wms

import (
	"github.com/spaulg/solo/internal/pkg/impl/common/app/cmd"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/wms/wf"
)

type Step struct {
	id               string
	name             string
	command          string
	arguments        []string
	workingDirectory string
	shell            string
}

func NewStep(id string, name string, command string, workingDirectory string, shell string) *Step {
	command, arguments := cmd.SplitCommand(shell, command)

	return &Step{
		id:               id,
		name:             name,
		command:          command,
		arguments:        arguments,
		workingDirectory: workingDirectory,
		shell:            shell,
	}
}

func (t *Step) GetID() string {
	return t.id
}

func (t *Step) GetName() string {
	return t.name
}

func (t *Step) GetCommand() string {
	return t.command
}

func (t *Step) GetArguments() []string {
	return t.arguments
}

func (t *Step) GetShell() string {
	return t.shell
}

func (t *Step) GetWorkingDirectory() string {
	return t.workingDirectory
}

func (t *Step) Trigger(
	start wf.StepTriggerFunc,
	progress wf.StepProgressFunc,
	complete wf.StepCompleteFunc,
) error {
	// Start step
	if err := start(); err != nil {
		return err
	}

	// Cycle progress
	for {
		exitCode, err := progress()

		if err != nil {
			return err
		}

		if exitCode != nil {
			// Report complete and exit
			if err := complete(*exitCode); err != nil {
				return err
			}

			break
		}
	}

	return nil
}
