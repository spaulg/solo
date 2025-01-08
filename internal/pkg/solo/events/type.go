package events

type EventType int

const (
	WorkflowStarted EventType = iota
	WorkflowStepStarted
	WorkflowStepOutput
	WorkflowStepComplete
	WorkflowComplete
)

func (t EventType) String() string {
	return [...]string{
		"WorkflowStarted",
		"WorkflowStepStarted",
		"WorkflowStepOutput",
		"WorkflowStepComplete",
		"WorkflowComplete",
	}[t]
}
