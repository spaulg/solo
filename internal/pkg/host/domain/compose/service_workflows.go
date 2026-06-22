package compose

type ServiceWorkflows map[string]ServiceWorkflowConfig

func NewServiceWorkflows() ServiceWorkflows {
	return make(ServiceWorkflows)
}
