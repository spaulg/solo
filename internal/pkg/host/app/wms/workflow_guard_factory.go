package wms

import (
	"log/slog"

	workflowcommon "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/host/app/wms/wf"
	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
)

type WorkflowGuardFactory struct {
	logger  *slog.Logger
	config  *domain2.Config
	project domain2.Project
}

func NewWorkflowGuardFactory(logger *slog.Logger, config *domain2.Config, project domain2.Project) *WorkflowGuardFactory {
	return &WorkflowGuardFactory{
		logger:  logger,
		config:  config,
		project: project,
	}
}

func (t *WorkflowGuardFactory) Build(workflowNames []workflowcommon.WorkflowName, containerNames []string) wf.Guard {
	return NewWorkflowGuard(t.logger, t.config, t.project, workflowNames, containerNames)
}
