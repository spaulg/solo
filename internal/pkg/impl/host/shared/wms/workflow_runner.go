package wms

type WorkflowRunner interface {
	RunWorkflow(workflowSession WorkflowSession) (bool, error)
}
