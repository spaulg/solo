package interceptors

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
)

const FirstPreStartContainerCompleteMetadataKey = "first_pre_start_container_complete"
const FirstPostStartContainerCompleteMetadataKey = "first_post_start_container_complete"
const FirstPreStartContainerCompleteContextValueName = "FirstPreStartComplete"
const FirstPostStartContainerCompleteContextValueName = "FirstPostStartComplete"

type FirstPreStartComplete string
type FirstPostStartComplete string

type FirstPreStartCompleteInterceptor struct {
	orchestrator container_types.Orchestrator
}

func NewFirstPreStartCompleteInterceptor(orchestrator container_types.Orchestrator) *FirstPreStartCompleteInterceptor {
	return &FirstPreStartCompleteInterceptor{
		orchestrator: orchestrator,
	}
}

func (t *FirstPreStartCompleteInterceptor) FirstPreStartCompleteUnaryInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to load metadata from incoming context")
	}

	firstPreStartComplete := md.Get(FirstPreStartContainerCompleteMetadataKey)
	if len(firstPreStartComplete) > 0 {
		ctx = context.WithValue(ctx, FirstPreStartComplete(FirstPreStartContainerCompleteContextValueName), firstPreStartComplete[0])
	}

	firstPostStartComplete := md.Get(FirstPostStartContainerCompleteMetadataKey)
	if len(firstPostStartComplete) > 0 {
		ctx = context.WithValue(ctx, FirstPostStartComplete(FirstPostStartContainerCompleteContextValueName), firstPostStartComplete[0])
	}

	return handler(ctx, req)
}

func (t *FirstPreStartCompleteInterceptor) FirstPreStartCompleteStreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := ss.Context()
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return fmt.Errorf("failed to load metadata from incoming context")
	}

	firstPreStartComplete := md.Get(FirstPreStartContainerCompleteMetadataKey)
	if len(firstPreStartComplete) > 0 {
		ctx = context.WithValue(ctx, FirstPreStartComplete(FirstPreStartContainerCompleteContextValueName), firstPreStartComplete[0])
	}

	firstPostStartComplete := md.Get(FirstPostStartContainerCompleteMetadataKey)
	if len(firstPostStartComplete) > 0 {
		ctx = context.WithValue(ctx, FirstPostStartComplete(FirstPostStartContainerCompleteContextValueName), firstPostStartComplete[0])
	}

	streamWrapper := NewServerStreamWrapper(ss, ctx)
	return handler(srv, streamWrapper)
}
