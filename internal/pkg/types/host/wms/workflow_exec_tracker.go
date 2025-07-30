package wms

import commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/wms"

type WorkflowExecTracker interface {
	MarkExecuted(serviceName string, workflowName commonworkflow.WorkflowName) (bool, error)
	Save() error
}
