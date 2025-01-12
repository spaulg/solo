package solo

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/certificate"
	"github.com/spaulg/solo/internal/pkg/solo/container"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/grpc"
	"github.com/spaulg/solo/internal/pkg/solo/grpc/service_definitions"
	"github.com/spaulg/solo/internal/pkg/solo/logs"
	"github.com/spaulg/solo/internal/pkg/solo/wms"
)

func ProjectControlFactory(soloCtx *context.SoloContext) (*ProjectControl, error) {
	// Provisioning grpc service
	eventManager := events.GetEventManagerInstance()
	eventManager.Subscribe(logs.NewLogWriterEventSubscriber(soloCtx))

	workflowFactory := wms.NewWorkflowFactory()
	workflowService := service_definitions.NewWorkflowService(soloCtx, eventManager, workflowFactory)

	// Container orchestrator
	containerOrchestrator := container.OrchestratorFactory(soloCtx)

	// GRPC server factory
	certificateAuthority, err := certificate.NewCertificateAuthority()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize certificate authority: %w", err)
	}

	grpcServerFactory := grpc.NewMutualTLSServerFactory(certificateAuthority, workflowService)

	// Project control
	projectControl := NewProjectControl(soloCtx, eventManager, containerOrchestrator, grpcServerFactory)

	return projectControl, nil
}
