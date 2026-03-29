package workflow

import (
	"io"

	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
)

type WorkflowRunner interface {
	io.Closer
	Execute(workflowName commonworkflow.WorkflowName) error
}
