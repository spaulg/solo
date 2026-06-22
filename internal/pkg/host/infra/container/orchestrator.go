package container

import (
	"google.golang.org/grpc/metadata"

	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
)

type Orchestrator interface {
	ComposeUp(serviceNames []string) error
	ComposeStop(serviceNames []string) error
	ComposeDown(serviceNames []string) error
	ComposeForkAndExecute(serviceName string, index int, command string, arguments []string, workingDirectory string) error
	ForkAndExecute(containerName string, command string, arguments []string, workingDirectory string) error
	StartCommand(containerName string, command []string) error
	RunCommand(containerName string, command []string) (string, error)
	GetHostGatewayHostname() string
	ServicesStatus(serviceNames []string) (*ServiceStatus, error)
	ExportComposeConfiguration(config *domain2.Config, project domain2.Project) ([]byte, error)
	ResolveContainerNameFromMetadata(md metadata.MD) (string, string, error)
	ResolveContainerNameFromServiceName(serviceName string, index int) (string, string, error)
	ResolveImageWorkingDirectory(serviceName string) (string, error)
}
