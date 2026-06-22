package service_definitions

import (
	"github.com/spaulg/solo/internal/pkg/host/app/wms/wf"
)

type WorkflowRunner interface {
	RunWorkflow(workflowSession wf.Session) error
}
