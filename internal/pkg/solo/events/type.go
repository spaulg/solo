package events

type EventType int

const (
	CommandProgress EventType = iota
	CommandFinished
	WorkflowFinished
)

func (t EventType) String() string {
	return [...]string{"CommandProgress"}[t]
}
