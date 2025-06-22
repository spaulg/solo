package wms

import (
	"strings"

	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type Step struct {
	id               string
	name             string
	command          string
	arguments        []string
	workingDirectory string
}

func NewStep(id string, name string, command string, workingDirectory *string) wms_types.Step {
	command, arguments := extractCommandArgs(command)

	cwd := "/"
	if workingDirectory != nil {
		cwd = *workingDirectory
	}

	return &Step{
		id:               id,
		name:             name,
		command:          command,
		arguments:        arguments,
		workingDirectory: cwd,
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

func (t *Step) Trigger(start wms_types.StepTriggerLambda, progress wms_types.StepProgressLambda, complete wms_types.StepCompleteLambda) error {
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

func extractCommandArgs(command string) (string, []string) {
	if []rune(command)[0] == '/' {
		// Exec format
		return extractExecCommandArgs(command)
	} else {
		// Shell format
		return extractShellCommandArgs(command)
	}
}

func extractExecCommandArgs(command string) (string, []string) {
	var extracted []string
	var current strings.Builder
	escaped := false
	singleQuoted := false
	doubleQuoted := false

	for _, char := range command {
		if char == '\\' && !escaped && !singleQuoted && !doubleQuoted {
			escaped = true
		} else if char == '"' && !escaped && !singleQuoted {
			if doubleQuoted {
				doubleQuoted = false

				extracted = append(extracted, current.String())
				current.Reset()
			} else {
				doubleQuoted = true
			}
		} else if char == '\'' && !escaped && !doubleQuoted {
			if singleQuoted {
				singleQuoted = false

				extracted = append(extracted, current.String())
				current.Reset()
			} else {
				singleQuoted = true
			}
		} else if char == ' ' && !escaped && !singleQuoted && !doubleQuoted {
			if current.Len() > 0 {
				extracted = append(extracted, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(char)
			escaped = false
		}
	}

	if current.Len() > 0 {
		extracted = append(extracted, current.String())
	}

	return extracted[0], extracted[1:]
}

func extractShellCommandArgs(command string) (string, []string) {
	return "/bin/sh", []string{"-c", command}
}
