package wms

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
