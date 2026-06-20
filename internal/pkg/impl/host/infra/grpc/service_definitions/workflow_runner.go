package service_definitions

import (
	"github.com/spaulg/solo/internal/pkg/impl/host/app/wms/wf"
)

type WorkflowRunner interface {
	RunWorkflow(workflowSession wf.Session) error
}
