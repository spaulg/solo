package app

import (
	"fmt"

	"github.com/spaulg/solo/internal/pkg/host/app/audit"
	"github.com/spaulg/solo/internal/pkg/host/app/context"
	"github.com/spaulg/solo/internal/pkg/host/app/events"
	wms2 "github.com/spaulg/solo/internal/pkg/host/app/wms"
	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/internal/pkg/host/infra/certificate"
	"github.com/spaulg/solo/internal/pkg/host/infra/container"
	"github.com/spaulg/solo/internal/pkg/host/infra/grpc"
	repository2 "github.com/spaulg/solo/internal/pkg/host/infra/repository"
)

func ProjectControlFactory(soloCtx *context.CliContext) (*ProjectControl, error) {
	execEventRepository := repository2.NewJSONFileRepository[*domain2.ExecutionEvent]()
	workflowLogMetaRepository := repository2.NewJSONFileRepository[domain2.WorkflowLogMeta]()
	workflowStepLogMetaRepository := repository2.NewJSONFileRepository[*domain2.WorkflowStepLogMeta]()
	logWriter := repository2.NewAppendFileStore()

	auditor := audit.NewStateDirectoryAuditor(
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

	workflowFactory := wms2.NewWorkflowFactory()
	grpcServerFactory := grpc.NewMutualTLSServerFactory(soloCtx, eventManager, workflowFactory, certificateAuthority)

	workflowGuardFactory := wms2.NewWorkflowGuardFactory(soloCtx)

	// Project control
	projectControl := NewProjectControl(soloCtx, eventManager, orchestratorFactory, grpcServerFactory, workflowGuardFactory, auditor)

	return projectControl, nil
}
