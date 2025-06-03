package solo

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/certificate"
	"github.com/spaulg/solo/internal/pkg/solo/container"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/grpc"
	"github.com/spaulg/solo/internal/pkg/solo/grpc/service_definitions"
	"github.com/spaulg/solo/internal/pkg/solo/subscribers"
	"github.com/spaulg/solo/internal/pkg/solo/wms"
)

func ProjectControlFactory(soloCtx *context.CliContext) (*ProjectControl, error) {
	// Provisioning grpc service
	eventManager := events.GetEventManagerInstance()
	eventManager.Subscribe(subscribers.NewLogWriterEventSubscriber(soloCtx))

	workflowFactory := wms.NewWorkflowFactory()
	workflowService := service_definitions.NewWorkflowService(soloCtx, eventManager, workflowFactory)

	// Container orchestrator factory
	orchestratorFactory := container.NewOrchestratorFactory(soloCtx, eventManager)

	// GRPC server factory
	certificateAuthority, err := certificate.NewCertificateAuthority()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize certificate authority: %w", err)
	}

	grpcServerFactory := grpc.NewMutualTLSServerFactory(certificateAuthority, workflowService)

	// Project control
	projectControl := NewProjectControl(soloCtx, eventManager, orchestratorFactory, grpcServerFactory)

	return projectControl, nil
}
