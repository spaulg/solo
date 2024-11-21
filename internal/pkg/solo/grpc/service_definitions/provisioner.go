package service_definitions

import (
	"context"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/shared/grpc/services"
	"github.com/spaulg/solo/internal/pkg/solo/event"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"google.golang.org/grpc"
	"log"
)

type ProvisionerServerImpl struct {
	services.UnimplementedProvisionerServer
	eventStream event.Stream[events.ProvisioningEvent]
}

func NewProvisionerService(eventStream event.Stream[events.ProvisioningEvent]) *ProvisionerServerImpl {
	return &ProvisionerServerImpl{
		eventStream: eventStream,
	}
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

func (t ProvisionerServerImpl) NotifyProvisionerComplete(ctx context.Context, request *services.NotifyProvisionerCompleteRequest) (*services.NotifyProvisionerCompleteResponse, error) {
	// Extract service name
	serviceName, ok := ctx.Value("ServiceName").(string)
	if !ok {
		log.Println("Service name not found")
		return nil, fmt.Errorf("unauthorized")
	}

	fmt.Printf("compose service name: %s\n", serviceName)

	t.eventStream.Push(&events.ProvisioningEvent{
		EventType: events.Finished,
		Service:   serviceName,
		Status:    0,
	})

	return &services.NotifyProvisionerCompleteResponse{}, nil
}
