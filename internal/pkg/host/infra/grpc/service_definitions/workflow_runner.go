package service_definitions

import (
	"github.com/spaulg/solo/internal/pkg/host/infra/grpc/service_definitions/wfsession"
)

type WorkflowRunner interface {
	RunWorkflow(workflowSession wfsession.Session) error
}
