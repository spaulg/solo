package interceptors

import (
	"context"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/container"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const ContainerNameContextValueName = "ContainerName"

type ContainerName string

type ContainerNameInterceptor struct {
	orchestrator container.Orchestrator
}

func NewContainerNameInterceptor(orchestrator container.Orchestrator) *ContainerNameInterceptor {
	return &ContainerNameInterceptor{
		orchestrator: orchestrator,
	}
}

func (t *ContainerNameInterceptor) ContainerNameUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to load metadata from incoming context")
	}

	containerName, err := t.orchestrator.ResolveContainerNameFromMetadata(md)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, ContainerName(ContainerNameContextValueName), *containerName)
	return handler(ctx, req)
}

func (t *ContainerNameInterceptor) ContainerNameStreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := ss.Context()
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return fmt.Errorf("failed to load metadata from incoming context")
	}

	containerName, err := t.orchestrator.ResolveContainerNameFromMetadata(md)
	if err != nil {
		return err
	}

	streamWrapper := NewServerStreamWrapper(ss, context.WithValue(ctx, ContainerName(ContainerNameContextValueName), *containerName))
	return handler(srv, streamWrapper)
}
