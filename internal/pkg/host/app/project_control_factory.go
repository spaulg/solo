package app

import (
	"fmt"

	"github.com/spaulg/solo/internal/pkg/host/app/audit"
	"github.com/spaulg/solo/internal/pkg/host/app/context"
	"github.com/spaulg/solo/internal/pkg/host/app/event_manager"
	wms2 "github.com/spaulg/solo/internal/pkg/host/app/wms"
	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/internal/pkg/host/infra/certificate/self_signed"
	"github.com/spaulg/solo/internal/pkg/host/infra/container"
	"github.com/spaulg/solo/internal/pkg/host/infra/grpc"
	repository2 "github.com/spaulg/solo/internal/pkg/host/infra/repository"
)

func ProjectControlFactory(soloCtx *context.CliContext) (*ProjectControl, error) {
	execEventRepository := repository2.NewJSONFileRepository[*domain2.ExecutionEvent]()
	workflowLogMetaRepository := repository2.NewJSONFileRepository[domain2.WorkflowLogMeta]()
	workflowStepLogMetaRepository := repository2.NewJSONFileRepository[*domain2.WorkflowStepLogMeta]()
	workflowExecTraceRepository := repository2.NewJSONFileRepository[*domain2.WorkflowExecTrace]()
	logWriter := repository2.NewAppendFileStore()

	auditor := audit.NewAuditor(
		soloCtx,
		soloCtx.Logger,
		soloCtx.Config,
		soloCtx.Project,
		execEventRepository,
		workflowLogMetaRepository,
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

	workflowFactory := wms2.NewWorkflowFactory()
	workflowRunner := wms2.NewWorkflowRunner(soloCtx.Config, soloCtx.Project, eventManager, workflowFactory)

	grpcServerFactory := grpc.NewMutualTLSServerFactory(soloCtx.Logger, soloCtx.Project, certificateAuthority, workflowRunner)

	workflowGuardFactory := wms2.NewWorkflowGuardFactory(soloCtx.Logger, soloCtx.Config, soloCtx.Project)

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
