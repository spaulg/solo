package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
)

type WorkflowFactory interface {
	Make(
		project project_types.Project,
		orchestrator container_types.Orchestrator,
		service string,
		workflowName workflowcommon.WorkflowName,
	) (Orchestrator, error)
}
