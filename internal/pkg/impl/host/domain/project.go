package domain

import (
	"iter"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
)

type Project interface {
	GetCompose() *types.Project
	Profiles() []string
	ReloadWithProfiles(profiles []string) error
	ResolveStateDirectory(relativePath string) string
	GetAllServicesStateDirectory() string
	GetServiceStateDirectoryRoot() string
	GetServiceStateDirectory(serviceName string) string
	GetServiceLogDirectory(serviceName string) string
	GetServiceMountDirectory(serviceName string) string
	GetStateDirectoryRoot() string
	GetDirectory() string
	GetFilePath() string
	GetGeneratedComposeFilePath() string
	GetMaxWorkflowTimeout(eventName string) time.Duration
	Name() string
	Services() Services
	Tools() Tools
}

type Services interface {
	GetService(serviceName string) ServiceConfig
	ServiceConfigIterator() iter.Seq2[string, ServiceConfig]
	HasService(serviceName string) bool
	ServiceNames() []string
	ExclusiveServiceNames() []string
	ContainerNames(serviceNames []string) ([]string, error)
	ProfilesOfServices(serviceNames []string) ([]string, error)
}

type ServiceConfig interface {
	GetServiceWorkflow(eventName string) ServiceWorkflowConfig
	GetConfig() types.ServiceConfig
	ResolveContainerWorkingDirectory(cwd string) string
}

type ServiceWorkflowConfig interface {
	Timeout() types.Duration
	Shell() *string
	Steps() []WorkflowStep
}

type WorkflowStep interface {
	Name() string
	Run() string
	Shell() *string
	WorkingDirectory() *string
}

type Tools map[string]ToolConfig

type ToolConfig interface {
	Description() string
	Command() string
	Service() string
	WorkingDirectory() string
	Shell() *string
}
