package wms

import (
	"log/slog"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/wms/wf"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
)

type WorkflowGuardFactory struct {
	logger  *slog.Logger
	config  *domain.Config
	project domain.Project
}

func NewWorkflowGuardFactory(logger *slog.Logger, config *domain.Config, project domain.Project) *WorkflowGuardFactory {
	return &WorkflowGuardFactory{
		logger:  logger,
		config:  config,
		project: project,
	}
}

func (t *WorkflowGuardFactory) Build(workflowNames []workflowcommon.WorkflowName, containerNames []string) wf.Guard {
	return NewWorkflowGuard(t.logger, t.config, t.project, workflowNames, containerNames)
}
