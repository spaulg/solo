package workflow

type StepTriggerFunc func() error
type StepProgressFunc func() (*uint8, error)
type StepCompleteFunc func(exitCode uint8) error

type Step interface {
	Trigger(trigger StepTriggerFunc, progress StepProgressFunc, complete StepCompleteFunc) error
	GetID() string
	GetName() string
	GetCommand() string
	GetArguments() []string
	GetShell() string
	GetWorkingDirectory() string
}
