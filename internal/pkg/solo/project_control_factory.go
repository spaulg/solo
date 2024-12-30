package solo

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/certificate"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/event"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/grpc"
	"github.com/spaulg/solo/internal/pkg/solo/grpc/service_definitions"
	"github.com/spaulg/solo/internal/pkg/solo/orchestrator"
)

func ProjectControlFactory(soloCtx *context.SoloContext) (*ProjectControl, error) {
	// Provisioning grpc service
	eventStream := event.NewStream[events.ProvisioningEvent]()
	provisionerService := service_definitions.NewProvisionerService(eventStream)

	// Container orchestrator
	containerOrchestrator := orchestrator.OrchestratorFactory(soloCtx)

	// GRPC server factory
	certificateAuthority, err := certificate.NewCertificateAuthority()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize certificate authority: %w", err)
	}

	grpcServerFactory := grpc.NewMutualTLSServerFactory(certificateAuthority, provisionerService)

	// Project control
	projectControl := NewProjectControl(soloCtx, containerOrchestrator, grpcServerFactory, eventStream)

	return projectControl, nil
}
