package app

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/host/app/wms/wf"
)

type WorkflowGuardFactory interface {
	Build(workflowNames []workflowcommon.WorkflowName, containerNames []string) wf.Guard
}
