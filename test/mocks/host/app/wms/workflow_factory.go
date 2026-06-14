package wms

import (
	"github.com/stretchr/testify/mock"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/wms/wf"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
)

type MockWorkflowFactory struct {
	mock.Mock
}

func (m *MockWorkflowFactory) Make(
	config *domain.Config,
	project domain.Project,
	service string,
	serviceWorkingDirectory string,
	workflowName workflowcommon.WorkflowName,
) (wf.Definition, error) {
	args := m.Called(config, project, service, serviceWorkingDirectory, workflowName)
	if o, ok := args.Get(0).(wf.Definition); ok {
		return o, args.Error(1)
	}

	return nil, args.Error(1)
}
