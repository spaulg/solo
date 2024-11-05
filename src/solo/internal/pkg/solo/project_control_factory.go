package solo

import (
	"fmt"
	"github.com/spaulg/solo/cli/internal/pkg/solo/config"
	"github.com/spaulg/solo/cli/internal/pkg/solo/grpc"
	"github.com/spaulg/solo/cli/internal/pkg/solo/orchestrator"
	"github.com/spaulg/solo/cli/internal/pkg/solo/project"
)

func ProjectControlFactory(config *config.Config, project *project.Project) (*ProjectControl, error) {
	// Container orchestrator
	containerOrchestrator := orchestrator.OrchestratorFactory(config)

	hostname := containerOrchestrator.GetHostGatewayHostname()
	stateDirectory := project.GetAllServicesStateDirectory()
	port := config.GrpcServerPort

	// Certificate generator
	certificateGenerator, err := grpc.NewCertificateGenerator(hostname, stateDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate generator: %v", err)
	}

	// GRPC server
	grpcServer := grpc.NewServer(
		hostname,
		port,
		stateDirectory,
		certificateGenerator,
	)

	// Project control
	projectControl := NewProjectControl(config, project, containerOrchestrator, grpcServer)
	return projectControl, nil
}
