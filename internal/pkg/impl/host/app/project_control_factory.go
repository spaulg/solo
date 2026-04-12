package app

import (
	"fmt"

	"github.com/spaulg/solo/internal/pkg/impl/host/app/audit"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/events"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/certificate"
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/container"
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/grpc"
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/repository"
)

func ProjectControlFactory(soloCtx *context.CliContext) (*ProjectControl, error) {
	execEventRepository := repository.NewJSONFileRepository[*domain.ExecutionEvent]()
	workflowLogMetaRepository := repository.NewJSONFileRepository[domain.WorkflowLogMeta]()
	workflowStepLogMetaRepository := repository.NewJSONFileRepository[*domain.WorkflowStepLogMeta]()
	logWriter := repository.NewAppendFileStore()

	auditor := audit.NewAuditor(
		soloCtx,
		execEventRepository,
		workflowLogMetaRepository,
		workflowStepLogMetaRepository,
		logWriter,
	)

	// Event manager
	eventManager := events.GetEventManagerInstance()
	eventManager.Subscribe(auditor)

	// Container orchestrator factory
	orchestratorFactory := container.NewOrchestratorFactory(soloCtx, eventManager)

	// GRPC server factory
	certificateAuthority, err := certificate.NewCertificateAuthority()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize certificate authority: %w", err)
	}

	workflowFactory := wms.NewWorkflowFactory()
	workflowRunner := wms.NewWorkflowRunner(soloCtx, eventManager, workflowFactory)

	grpcServerFactory := grpc.NewMutualTLSServerFactory(soloCtx, eventManager, certificateAuthority, workflowRunner)

	workflowGuardFactory := wms.NewWorkflowGuardFactory(soloCtx)

	// Project control
	projectControl := NewProjectControl(soloCtx, eventManager, orchestratorFactory, grpcServerFactory, workflowGuardFactory, auditor)

	return projectControl, nil
}
