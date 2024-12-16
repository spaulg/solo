package solo

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/certificate"
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/event"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/grpc"
	"github.com/spaulg/solo/internal/pkg/solo/grpc/service_definitions"
	"github.com/spaulg/solo/internal/pkg/solo/orchestrator"
	"github.com/spaulg/solo/internal/pkg/solo/project"
)

func ProjectControlFactory(config *config.Config, project *project.Project) (*ProjectControl, error) {
	// Provisioning grpc service
	eventStream := event.NewStream[events.ProvisioningEvent]()
	provisionerService := service_definitions.NewProvisionerService(eventStream)

	// Container orchestrator
	containerOrchestrator := orchestrator.OrchestratorFactory(config)

	// GRPC server factory
	certificateAuthority, err := certificate.NewCertificateAuthority()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize certificate authority: %w", err)
	}

	grpcServerFactory := grpc.NewMutualTLSServerFactory(certificateAuthority, provisionerService)

	// Project control
	projectControl := NewProjectControl(config, project, containerOrchestrator, grpcServerFactory, eventStream)
	return projectControl, nil
}
