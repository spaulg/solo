package host

import (
	"fmt"

	"github.com/spaulg/solo/internal/pkg/impl/host/audit"
	"github.com/spaulg/solo/internal/pkg/impl/host/certificate"
	"github.com/spaulg/solo/internal/pkg/impl/host/container"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/events"
	"github.com/spaulg/solo/internal/pkg/impl/host/grpc"
	"github.com/spaulg/solo/internal/pkg/impl/host/wms"
)

func ProjectControlFactory(soloCtx *context.CliContext) (*ProjectControl, error) {
	workflowLogWriter := audit.NewWorkflowLogWriter(soloCtx)

	// Event manager
	eventManager := events.GetEventManagerInstance()
	eventManager.Subscribe(workflowLogWriter)

	// Container orchestrator factory
	orchestratorFactory := container.NewOrchestratorFactory(soloCtx, eventManager)

	// GRPC server factory
	certificateAuthority, err := certificate.NewCertificateAuthority()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize certificate authority: %w", err)
	}

	workflowFactory := wms.NewWorkflowFactory()
	grpcServerFactory := grpc.NewMutualTLSServerFactory(soloCtx, eventManager, workflowFactory, certificateAuthority)

	workflowGuardFactory := wms.NewWorkflowGuardFactory(soloCtx)

	// Project control
	projectControl := NewProjectControl(soloCtx, eventManager, orchestratorFactory, grpcServerFactory, workflowGuardFactory, workflowLogWriter)

	return projectControl, nil
}
