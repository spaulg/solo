package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	context_types "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/infra/container"
)

type WorkflowFactory interface {
	Make(
		soloCtx *context_types.CliContext,
		orchestrator container_types.Orchestrator,
		service string,
		workflowName workflowcommon.WorkflowName,
	) (Workflow, error)
}
