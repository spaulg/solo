package wms

import (
	"github.com/stretchr/testify/mock"

	workflowcommon "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/host/app/wms/wf"
	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
)

type MockWorkflowFactory struct {
	mock.Mock
}

func (m *MockWorkflowFactory) Make(
	config *domain2.Config,
	project domain2.Project,
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
