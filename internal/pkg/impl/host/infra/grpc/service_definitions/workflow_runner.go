package service_definitions

import (
	"github.com/spaulg/solo/internal/pkg/impl/host/app/wms/workflow"
)

type WorkflowRunner interface {
	RunWorkflow(workflowSession workflow.Session) error
}
