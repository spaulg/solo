package interceptors

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	workflowcommon "github.com/spaulg/solo/internal/pkg/shared/domain/wms"
	container_types "github.com/spaulg/solo/internal/pkg/shared/infra/container"
)

type FirstContainerComplete workflowcommon.WorkflowName

type FirstContainerCompleteInterceptor struct {
	orchestrator container_types.Orchestrator
}

func NewFirstContainerCompleteInterceptor(orchestrator container_types.Orchestrator) *FirstContainerCompleteInterceptor {
	return &FirstContainerCompleteInterceptor{
		orchestrator: orchestrator,
	}
}

func (t *FirstContainerCompleteInterceptor) FirstContainerCompleteUnaryInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to load metadata from incoming context")
	}

	for _, workflow := range workflowcommon.WorkflowNames {
		if workflow.IsFirstContainerWorkflow() {
			firstContainerComplete := md.Get(workflow.String() + "_complete")
			if len(firstContainerComplete) > 0 {
				ctx = context.WithValue(ctx, FirstContainerComplete(workflow), firstContainerComplete[0])
			}
		}
	}

	return handler(ctx, req)
}

func (t *FirstContainerCompleteInterceptor) FirstContainerCompleteStreamInterceptor(
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

	for _, workflow := range workflowcommon.WorkflowNames {
		if workflow.IsFirstContainerWorkflow() {
			firstContainerComplete := md.Get(workflow.String() + "_complete")
			if len(firstContainerComplete) > 0 {
				ctx = context.WithValue(ctx, FirstContainerComplete(workflow), firstContainerComplete[0])
			}
		}
	}

	streamWrapper := NewServerStreamWrapper(ctx, ss)
	return handler(srv, streamWrapper)
}
