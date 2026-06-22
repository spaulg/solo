package grpc

import (
	"github.com/spaulg/solo/internal/pkg/host/infra/grpc/interceptors"
	"github.com/spaulg/solo/internal/pkg/host/infra/grpc/service_definitions"
)

type ContainerResolver interface {
	interceptors.ContainerNameResolver
	service_definitions.ContainerImageWorkingDirectoryResolver

	GetHostGatewayHostname() string
}
