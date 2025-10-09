package logging

type WorkflowExecutionLogMeta interface {
	MarkComplete(res error)
	Persist() error

	GetCommandPath() string
	GetCommandArgs() []string
	GetError() string
	GetComplete() bool
}
