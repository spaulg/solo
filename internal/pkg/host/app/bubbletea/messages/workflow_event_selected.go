package messages

type WorkflowEventSelected struct {
	ExecutionEventName string
	ContainerName      string
	WorkflowName       string
}
