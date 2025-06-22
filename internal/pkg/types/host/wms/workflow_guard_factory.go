package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
)

type WorkflowGuardFactory interface {
	Build(workflowNames []workflowcommon.WorkflowName, containerNames []string) WorkflowGuard
}
