package interceptors

import (
	"context"
	"fmt"

	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const FirstPreStartCompleteMetadataKey = "first_pre_start_complete"
const FirstPreStartCompleteContextValueName = "FirstPreStartComplete"

type FirstPreStartComplete string

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
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to load metadata from incoming context")
	}

	firstPreStartComplete := md.Get(FirstPreStartCompleteMetadataKey)
	if len(firstPreStartComplete) > 0 {
		ctx = context.WithValue(ctx, FirstPreStartComplete(FirstPreStartCompleteContextValueName), firstPreStartComplete[0])
	}

	return handler(ctx, req)
}

func (t *FirstPreStartCompleteInterceptor) FirstPreStartCompleteStreamInterceptor(
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

	firstPreStartComplete := md.Get(FirstPreStartCompleteMetadataKey)
	if len(firstPreStartComplete) > 0 {
		ctx = context.WithValue(ctx, FirstPreStartComplete(FirstPreStartCompleteContextValueName), firstPreStartComplete[0])
	}

	streamWrapper := NewServerStreamWrapper(ss, ctx)
	return handler(srv, streamWrapper)
}
