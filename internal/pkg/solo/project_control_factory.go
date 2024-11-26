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

	hostname := containerOrchestrator.GetHostGatewayHostname()
	stateDirectory := project.GetAllServicesStateDirectory()

	// Certificate generator
	certificateGenerator, err := certificate.NewCertificateGenerator(hostname, stateDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate generator: %v", err)
	}

	// GRPC server
	grpcServer := grpc.NewMutualTLSServerFactory(
		certificateGenerator,
		provisionerService,
	)

	// Project control
	projectControl := NewProjectControl(config, project, containerOrchestrator, grpcServer, eventStream)
	return projectControl, nil
}
