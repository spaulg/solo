package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/shared/domain/wms"
)

type WorkflowGuardFactory interface {
	Build(workflowNames []workflowcommon.WorkflowName, containerNames []string) WorkflowGuard
}
