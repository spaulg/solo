package compose

import (
	"github.com/compose-spec/compose-go/v2/types"

	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
	compose_types "github.com/spaulg/solo/internal/pkg/types/host/project/compose"
)

type ServiceConfig struct {
	serviceConfig types.ServiceConfig
}

func NewServiceConfig(serviceConfig types.ServiceConfig) compose_types.ServiceConfig {
	return &ServiceConfig{
		serviceConfig: serviceConfig,
	}
}

func (t *ServiceConfig) GetServiceWorkflow(eventName string) compose_types.ServiceWorkflowConfig {
	serviceWorkflows := t.serviceConfig.Extensions[project_types.ServiceWorkflowExtensionName].(compose_types.ServiceWorkflows)
	return serviceWorkflows[eventName]
}

func (t *ServiceConfig) GetConfig() types.ServiceConfig {
	return t.serviceConfig
}
