package wms

import "strings"

type StepTriggerLambda func() error
type StepProgressLambda func() (*uint8, error)
type StepCompleteLambda func(exitCode uint8) error

type Step interface {
	Trigger(trigger StepTriggerLambda, progress StepProgressLambda, complete StepCompleteLambda) error
	GetId() string
	GetName() string
	GetCommand() string
	GetArguments() []string
	GetWorkingDirectory() string
}

type DefaultStep struct {
	id               string
	name             string
	command          string
	arguments        []string
	workingDirectory string
}

func NewStep(id string, name string, command string, workingDirectory *string) Step {
	command, arguments := extractCommandArgs(command)

	cwd := "/"
	if workingDirectory != nil {
		cwd = *workingDirectory
	}

	return &DefaultStep{
		id:               id,
		name:             name,
		command:          command,
		arguments:        arguments,
		workingDirectory: cwd,
	}
}

func (t *DefaultStep) GetId() string {
	return t.id
}

func (t *DefaultStep) GetName() string {
	return t.name
}

func (t *DefaultStep) GetCommand() string {
	return t.command
}

func (t *DefaultStep) GetArguments() []string {
	return t.arguments
}

func (t *DefaultStep) GetWorkingDirectory() string {
	return t.workingDirectory
}

func (t *DefaultStep) Trigger(start StepTriggerLambda, progress StepProgressLambda, complete StepCompleteLambda) error {
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
