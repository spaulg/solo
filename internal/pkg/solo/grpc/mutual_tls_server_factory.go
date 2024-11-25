package grpc

import (
	"github.com/spaulg/solo/internal/pkg/solo/grpc/credentials"
	"github.com/spaulg/solo/internal/pkg/solo/grpc/service_definitions"
	"github.com/spaulg/solo/internal/pkg/solo/project"
)

type MutualTLSServerFactory struct {
	hostname           string
	port               uint16
	stateDirectory     string
	credentialsBuilder credentials.Builder
	provisionerServer  *service_definitions.ProvisionerServerImpl
}

func NewMutualTLSServerFactory(
	hostname string,
	port uint16,
	stateDirectory string,
	credentialsBuilder credentials.Builder,
	provisionerServer *service_definitions.ProvisionerServerImpl,
) ServerFactory {
	return &MutualTLSServerFactory{
		hostname:           hostname,
		port:               port,
		stateDirectory:     stateDirectory,
		credentialsBuilder: credentialsBuilder,
		provisionerServer:  provisionerServer,
	}
}

func (t *MutualTLSServerFactory) Build(project *project.Project) (Server, error) {

	// todo: Refactor credentials builder to move logic in to the new factory
	// todo: Make the factory build peer certificates for each service and store in the services state directory

	transportCredentials, err := t.credentialsBuilder.Build()
	if err != nil {
		return nil, err
	}

	return NewAsynchronousServer(
		t.hostname,
		t.port,
		t.stateDirectory,
		transportCredentials,
		t.provisionerServer,
	), nil
}
