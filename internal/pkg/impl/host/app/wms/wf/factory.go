package wf

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
)

type Factory interface {
	Make(
		config *domain.Config,
		project domain.Project,
		service string,
		serviceWorkingDirectory string,
		workflowName workflowcommon.WorkflowName,
	) (Definition, error)
}
