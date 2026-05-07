package app

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/types/host/app/wms"
)

type WorkflowGuardFactory interface {
	Build(workflowNames []workflowcommon.WorkflowName, containerNames []string) wms.WorkflowGuard
}
