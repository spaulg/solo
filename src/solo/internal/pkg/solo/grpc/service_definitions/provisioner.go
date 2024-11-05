package service_definitions

import (
	"context"
	"github.com/spaulg/solo/shared/pkg/solo/grpc/services"
	"google.golang.org/grpc"
)

type ProvisionerServerImpl struct {
	services.UnimplementedProvisionerServer
}

func (ProvisionerServerImpl) GetProvisionerSteps(context.Context, *services.GetProvisionerStepsRequest) (*services.GetProvisionerStepsResponse, error) {
	return &services.GetProvisionerStepsResponse{}, nil
}

func (ProvisionerServerImpl) PublishCommandOutput(grpc.ClientStreamingServer[services.PublishCommandOutputRequest, services.PublishCommandOutputResponse]) error {
	return nil
}

func (ProvisionerServerImpl) PublishCommandResult(context.Context, *services.PublishCommandResultRequest) (*services.PublishCommandResultResponse, error) {
	return &services.PublishCommandResultResponse{}, nil
}

func (ProvisionerServerImpl) NotifyProvisionerComplete(context.Context, *services.NotifyProvisionerCompleteRequest) (*services.NotifyProvisionerCompleteResponse, error) {
	return &services.NotifyProvisionerCompleteResponse{}, nil
}
