package service_definitions

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/spaulg/solo/internal/pkg/shared/infra/grpc/services"
)

type RunWorkflowStreamWrapper struct {
	server grpc.BidiStreamingServer[services.RunWorkflowStreamRequest, services.WorkflowStreamResponse]
}

func NewRunWorkflowStreamWrapper(
	server grpc.BidiStreamingServer[services.RunWorkflowStreamRequest, services.WorkflowStreamResponse],
) RunWorkflowStreamWrapper {
	return RunWorkflowStreamWrapper{
		server: server,
	}
}

func (t RunWorkflowStreamWrapper) Recv() (*services.WorkflowStreamRequest, error) {
	req, err := t.server.Recv()
	if err != nil {
		return nil, err
	}

	switch subRequest := (req.Request).(type) {
	case *services.RunWorkflowStreamRequest_StreamRequest:
		return subRequest.StreamRequest, nil
	default:
		return nil, errors.New("not a stream request")
	}
}

func (t RunWorkflowStreamWrapper) Send(response *services.WorkflowStreamResponse) error {
	return t.server.Send(response)
}

func (t RunWorkflowStreamWrapper) SetHeader(md metadata.MD) error {
	return t.server.SetHeader(md)
}

func (t RunWorkflowStreamWrapper) SendHeader(md metadata.MD) error {
	return t.server.SendHeader(md)
}

func (t RunWorkflowStreamWrapper) SetTrailer(md metadata.MD) {
	t.server.SetTrailer(md)
}

func (t RunWorkflowStreamWrapper) Context() context.Context {
	return t.server.Context()
}

func (t RunWorkflowStreamWrapper) SendMsg(m any) error {
	return t.server.SendMsg(m)
}

func (t RunWorkflowStreamWrapper) RecvMsg(m any) error {
	return t.server.RecvMsg(m)
}
