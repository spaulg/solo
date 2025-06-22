package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
)

type WorkflowFactory interface {
	Make(project project_types.Project, service string, workflowName workflowcommon.WorkflowName) Orchestrator
}
