package app

import (
	"fmt"

	"github.com/spaulg/solo/internal/pkg/host/app/audit"
	"github.com/spaulg/solo/internal/pkg/host/app/context"
	"github.com/spaulg/solo/internal/pkg/host/app/event_manager"
	"github.com/spaulg/solo/internal/pkg/host/app/wms"
	"github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/internal/pkg/host/infra/certificate/self_signed"
	"github.com/spaulg/solo/internal/pkg/host/infra/container"
	"github.com/spaulg/solo/internal/pkg/host/infra/grpc"
	"github.com/spaulg/solo/internal/pkg/host/infra/repository"
)

func ProjectControlFactory(soloCtx *context.CliContext) (*ProjectControl, error) {
	execEventRepository := repository.NewJSONFileRepository[*domain.ExecutionEvent]()
	containerStepMapRepository := repository.NewJSONFileRepository[domain.ContainerStepMap]()
	workflowStepLogMetaRepository := repository.NewJSONFileRepository[*domain.WorkflowStepLogMeta]()
	workflowExecTraceRepository := repository.NewJSONFileRepository[*domain.WorkflowExecTrace]()
	logWriter := repository.NewAppendFileStore()

	auditor := audit.NewAuditor(
		soloCtx,
		soloCtx.Logger,
		soloCtx.Config,
		soloCtx.Project,
		execEventRepository,
		containerStepMapRepository,
		workflowStepLogMetaRepository,
		logWriter,
	)

	// Event manager
	eventManager := event_manager.GetEventManagerInstance()
	eventManager.Subscribe(auditor)

	// Container orchestrator factory
	orchestratorFactory := container.NewOrchestratorFactory(soloCtx.Logger, soloCtx.Config, soloCtx.Project, eventManager)

	// GRPC server factory
	certificateAuthority, err := self_signed.NewCertificateAuthority()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize certificate authority: %w", err)
	}

	workflowFactory := wms.NewWorkflowFactory()
	workflowRunner := wms.NewWorkflowRunner(soloCtx.Config, soloCtx.Project, eventManager, workflowFactory)

	grpcServerFactory := grpc.NewMutualTLSServerFactory(soloCtx.Logger, soloCtx.Project, certificateAuthority, workflowRunner)

	workflowGuardFactory := wms.NewWorkflowGuardFactory(soloCtx.Logger, soloCtx.Config, soloCtx.Project)

	// Project control
	projectControl := NewProjectControl(
		soloCtx,
		eventManager,
		orchestratorFactory,
		grpcServerFactory,
		workflowGuardFactory,
		auditor,
		workflowExecTraceRepository,
	)

	return projectControl, nil
}
