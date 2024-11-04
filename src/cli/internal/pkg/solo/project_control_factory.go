package solo

import (
	"github.com/spaulg/solo/cli/internal/pkg/solo/config"
	"github.com/spaulg/solo/cli/internal/pkg/solo/grpc"
	"github.com/spaulg/solo/cli/internal/pkg/solo/orchestrator"
	"github.com/spaulg/solo/cli/internal/pkg/solo/project"
)

func ProjectControlFactory(config *config.Config, project *project.Project) *ProjectControl {
	// Container orchestrator
	containerOrchestrator := orchestrator.OrchestratorFactory(config)

	// GRPC server
	grpcServer := grpc.NewServer(
		containerOrchestrator.GetHostGatewayHostname(),
		config.GrpcServerPort,
		project.GetAllServicesStateDirectory(),
	)

	// Project control
	projectControl := NewProjectControl(config, project, containerOrchestrator, grpcServer)
	return projectControl
}
