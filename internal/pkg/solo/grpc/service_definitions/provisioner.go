package service_definitions

import (
	"context"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/shared/grpc/services"
	"github.com/spaulg/solo/internal/pkg/solo/event"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"strings"
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

func (t ProvisionerServerImpl) NotifyProvisionerComplete(context context.Context, request *services.NotifyProvisionerCompleteRequest) (*services.NotifyProvisionerCompleteResponse, error) {
	peer, ok := peer.FromContext(context)
	if !ok {
		return nil, fmt.Errorf("unexpected peer transport credentials")
	}

	tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, fmt.Errorf("unexpected peer transport credentials")
	}

	if len(tlsInfo.State.PeerCertificates) == 0 {
		return nil, fmt.Errorf("missing peer certificate")
	}

	clientCert := tlsInfo.State.PeerCertificates[0]
	fmt.Printf("compose service name: %s\n", clientCert.Subject.CommonName)

	lastIndex := strings.LastIndex(clientCert.Subject.CommonName, ":")
	if lastIndex == -1 {
		return nil, fmt.Errorf("invalid subject common name")
	}

	serviceName := clientCert.Subject.CommonName[lastIndex+len(":"):]

	t.eventStream.Push(&events.ProvisioningEvent{
		EventType: events.Finished,
		Service:   serviceName,
		Status:    0,
	})

	return &services.NotifyProvisionerCompleteResponse{}, nil
}
