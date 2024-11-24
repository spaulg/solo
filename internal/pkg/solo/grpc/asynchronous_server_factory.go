package grpc

import (
	"github.com/spaulg/solo/internal/pkg/solo/grpc/credentials"
	"github.com/spaulg/solo/internal/pkg/solo/grpc/service_definitions"
	"github.com/spaulg/solo/internal/pkg/solo/project"
)

type AsynchronousServerFactory struct {
	hostname           string
	port               uint16
	stateDirectory     string
	credentialsBuilder credentials.Builder
	provisionerServer  *service_definitions.ProvisionerServerImpl
}

func NewAsynchronousServerFactory(
	hostname string,
	port uint16,
	stateDirectory string,
	credentialsBuilder credentials.Builder,
	provisionerServer *service_definitions.ProvisionerServerImpl,
) ServerFactory {
	return &AsynchronousServerFactory{
		hostname:           hostname,
		port:               port,
		stateDirectory:     stateDirectory,
		credentialsBuilder: credentialsBuilder,
		provisionerServer:  provisionerServer,
	}
}

func (t *AsynchronousServerFactory) Build(project *project.Project) Server {

	// todo: Refactor credentials builder to move logic in to the new factory
	// todo: Make the factory build peer certificates for each service and store in the services state directory

	return NewAsynchronousServer(
		t.hostname,
		t.port,
		t.stateDirectory,
		t.credentialsBuilder,
		t.provisionerServer,
	)
}
