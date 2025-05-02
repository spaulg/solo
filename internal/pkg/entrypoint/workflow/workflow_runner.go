package workflow

import (
	commonworkflow "github.com/spaulg/solo/internal/pkg/common/wms"
	"io"
)

type WorkflowRunner interface {
	io.Closer
	Execute(workflowName commonworkflow.WorkflowName)
}
