package wms

import (
	context_types "github.com/spaulg/solo/internal/pkg/host/app/context"
	workflowcommon "github.com/spaulg/solo/internal/pkg/shared/domain/wms"
	container_types "github.com/spaulg/solo/internal/pkg/shared/infra/container"
)

type WorkflowFactory interface {
	Make(
		soloCtx *context_types.CliContext,
		orchestrator container_types.Orchestrator,
		service string,
		workflowName workflowcommon.WorkflowName,
	) (Workflow, error)
}
