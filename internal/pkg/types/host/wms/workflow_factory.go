package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	context_types "github.com/spaulg/solo/internal/pkg/impl/host/context"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
)

type WorkflowFactory interface {
	Make(
		soloCtx *context_types.CliContext,
		orchestrator container_types.Orchestrator,
		service string,
		workflowName workflowcommon.WorkflowName,
	) (Orchestrator, error)
}
