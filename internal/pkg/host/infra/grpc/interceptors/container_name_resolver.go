package interceptors

import (
	"google.golang.org/grpc/metadata"
)

type ContainerNameResolver interface {
	ResolveContainerNameFromMetadata(md metadata.MD) (string, string, error)
}
