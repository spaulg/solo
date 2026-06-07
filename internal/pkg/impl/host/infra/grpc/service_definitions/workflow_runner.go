package service_definitions

import (
	"github.com/spaulg/solo/internal/pkg/impl/host/shared/wms"
)

type WorkflowRunner interface {
	RunWorkflow(workflowSession wms.WorkflowSession) error
}
