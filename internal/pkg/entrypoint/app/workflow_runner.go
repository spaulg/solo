package app

import (
	"io"

	commonworkflow "github.com/spaulg/solo/internal/pkg/shared/domain/wms"
)

type WorkflowRunner interface {
	io.Closer
	Execute(workflowName commonworkflow.WorkflowName) error
}
