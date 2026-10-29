package domain

type ContainerStepMap map[string][]string

func NewWorkflowContainerStepMap() ContainerStepMap {
	return make(ContainerStepMap)
}

func (t ContainerStepMap) AppendStep(containerName string, stepID string) {
	t[containerName] = append(t[containerName], stepID)
}
