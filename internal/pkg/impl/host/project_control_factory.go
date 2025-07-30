package host

import (
	"fmt"

	"github.com/spaulg/solo/internal/pkg/impl/host/certificate"
	"github.com/spaulg/solo/internal/pkg/impl/host/container"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/events"
	"github.com/spaulg/solo/internal/pkg/impl/host/grpc"
	"github.com/spaulg/solo/internal/pkg/impl/host/grpc/service_definitions"
	"github.com/spaulg/solo/internal/pkg/impl/host/subscribers"
	"github.com/spaulg/solo/internal/pkg/impl/host/wms"
)

func ProjectControlFactory(soloCtx *context.CliContext) (*ProjectControl, error) {
	// Provisioning grpc service
	eventManager := events.GetEventManagerInstance()
	eventManager.Subscribe(subscribers.NewLogWriterEventSubscriber(soloCtx))

	workflowExecutionTracker := service_definitions.LoadWorkflowExecTracker()
	workflowFactory := wms.NewWorkflowFactory()
	workflowService := service_definitions.NewWorkflowService(soloCtx, eventManager, workflowFactory, workflowExecutionTracker)

	// Container orchestrator factory
	orchestratorFactory := container.NewOrchestratorFactory(soloCtx, eventManager)

	// GRPC server factory
	certificateAuthority, err := certificate.NewCertificateAuthority()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize certificate authority: %w", err)
	}

	grpcServerFactory := grpc.NewMutualTLSServerFactory(certificateAuthority, workflowService)

	workflowGuardFactory := wms.NewWorkflowGuardFactory(soloCtx)

	// Project control
	projectControl := NewProjectControl(soloCtx, eventManager, orchestratorFactory, grpcServerFactory, workflowGuardFactory)

	return projectControl, nil
}
