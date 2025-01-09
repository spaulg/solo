package wms

type StepTriggerLambda func() error
type StepProgressLambda func() (*uint8, error)
type StepCompleteLambda func(exitCode uint8) error

type Step interface {
	Trigger(trigger StepTriggerLambda, progress StepProgressLambda, complete StepCompleteLambda) error
	GetName() string
	GetCommand() string
	GetCommandArguments() []string
	GetWorkingDirectory() *string
}

type DefaultStep struct {
	name             string
	command          string
	workingDirectory *string
}

func NewStep(name string, command string, workingDirectory string) Step {
	return &DefaultStep{
		name:             name,
		command:          command,
		workingDirectory: &workingDirectory,
	}
}

func (t *DefaultStep) GetName() string {
	return t.name
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
