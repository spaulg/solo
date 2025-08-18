package wms

import (
	"github.com/spaulg/solo/internal/pkg/impl/common/cmd"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type Step struct {
	id               string
	name             string
	command          string
	arguments        []string
	workingDirectory string
}

func NewStep(id string, name string, command string, workingDirectory string) wms_types.Step {
	command, arguments := cmd.SplitCommand(command)

	return &Step{
		id:               id,
		name:             name,
		command:          command,
		arguments:        arguments,
		workingDirectory: workingDirectory,
	}
}

func (t *Step) GetId() string {
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

func (t *Step) GetWorkingDirectory() string {
	return t.workingDirectory
}

func (t *Step) Trigger(
	start wms_types.StepTriggerLambda,
	progress wms_types.StepProgressLambda,
	complete wms_types.StepCompleteLambda,
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
