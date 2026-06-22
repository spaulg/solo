package domain

type WorkflowLogMeta map[string][]string

func NewWorkflowLogMeta() WorkflowLogMeta {
	return make(WorkflowLogMeta)
}

func (t WorkflowLogMeta) AppendStep(containerName string, stepID string) {
	t[containerName] = append(t[containerName], stepID)
}
