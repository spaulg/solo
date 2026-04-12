package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	context_types "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
)

type WorkflowFactory interface {
	Make(
		soloCtx *context_types.CliContext,
		service string,
		serviceWorkingDirectory string,
		workflowName workflowcommon.WorkflowName,
	) (Workflow, error)
}
