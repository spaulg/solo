package interceptors

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
)

type FirstWorkflowComplete workflowcommon.WorkflowName

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

	for _, workflow := range workflowcommon.WorkflowNames {
		if workflow.IsFirstContainerWorkflow() {
			firstWorkflowComplete := md.Get(workflow.String() + "_complete")
			if len(firstWorkflowComplete) > 0 {
				ctx = context.WithValue(ctx, FirstWorkflowComplete(workflow), firstWorkflowComplete[0])
			}
		}
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

	for _, workflow := range workflowcommon.WorkflowNames {
		if workflow.IsFirstContainerWorkflow() {
			firstWorkflowComplete := md.Get(workflow.String() + "_complete")
			if len(firstWorkflowComplete) > 0 {
				ctx = context.WithValue(ctx, FirstWorkflowComplete(workflow), firstWorkflowComplete[0])
			}
		}
	}

	streamWrapper := NewServerStreamWrapper(ss, ctx)
	return handler(srv, streamWrapper)
}
