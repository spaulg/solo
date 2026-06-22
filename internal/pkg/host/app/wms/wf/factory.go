package wf

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
)

type Factory interface {
	Make(
		config *domain2.Config,
		project domain2.Project,
		service string,
		serviceWorkingDirectory string,
		workflowName workflowcommon.WorkflowName,
	) (Definition, error)
}
