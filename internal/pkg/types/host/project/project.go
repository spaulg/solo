package project

import (
	"time"

	"github.com/compose-spec/compose-go/v2/types"
)

type Project interface {
	ResolveStateDirectory(relativePath string) string
	GetAllServicesStateDirectory() string
	GetServiceStateDirectoryRoot() string
	GetServiceStateDirectory(serviceName string) string
	GetServiceLogDirectory(serviceName string) string
	GetServiceMountDirectory(serviceName string) string
	GetStateDirectoryRoot() string
	GetDirectory() string
	GetFilePath() string
	GetServiceWorkflow(serviceName string, eventName string) ServiceWorkflowConfig
	GetGeneratedComposeFilePath() string
	GetMaxWorkflowTimeout(eventName string) time.Duration
	ContainerNames(serviceNames []string) ([]string, error)
	ProfilesOfServices(serviceNames []string) ([]string, error)
	Services() types.Services
	ServiceNames() []string
	ExclusiveServiceNames() []string
	MarshalYAML() ([]byte, error)
	Name() string
	ReloadWithAllProfilesEnabled() (Project, error)
}
