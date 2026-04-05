package compose

import "github.com/compose-spec/compose-go/v2/types"

type ServiceConfig interface {
	GetServiceWorkflow(eventName string) ServiceWorkflowConfig
	GetConfig() types.ServiceConfig
	ResolveContainerWorkingDirectory(cwd string) string
}
