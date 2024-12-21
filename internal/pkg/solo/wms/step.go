package wms

type StepProgress struct {
	ExitCode *uint8
	Stdout   *string
	Stderr   *string
}

type StepTriggerLambda func() error
type StepProgressLambda func() (*StepProgress, error)
type StepCompleteLambda func() error

type Step interface {
	Trigger(trigger StepTriggerLambda, progress StepProgressLambda, complete StepCompleteLambda) error
	GetCommand() string
	GetCommandArguments() []string
	GetWorkingDirectory() *string
}

type DefaultStep struct {
	command          string
	workingDirectory *string
}

func NewStep(command string, workingDirectory string) Step {
	return &DefaultStep{
		command:          command,
		workingDirectory: &workingDirectory,
	}
}

func (t *DefaultStep) GetCommand() string {
	return t.command
}

func (t *DefaultStep) GetCommandArguments() []string {
	return []string{}
}

func (t *DefaultStep) GetWorkingDirectory() *string {
	return t.workingDirectory
}

func (t *DefaultStep) Trigger(start StepTriggerLambda, progress StepProgressLambda, complete StepCompleteLambda) error {
	// Start step
	if err := start(); err != nil {
		return err
	}

	// Cycle progress
	for {
		progressStatus, err := progress()

		if err != nil {
			return err
		}

		if progressStatus.ExitCode != nil {
			// Report complete and exit
			if err := complete(); err != nil {
				return err
			}

			break
		}
	}

	return nil
}
