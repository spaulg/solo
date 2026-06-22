package domain

import (
	"testing"

	"github.com/stretchr/testify/suite"

	commonworkflow "github.com/spaulg/solo/internal/pkg/common/domain/wms"
)

func TestWorkflowExecTraceTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowExecTraceTestSuite))
}

type WorkflowExecTraceTestSuite struct {
	suite.Suite
}

func (t *WorkflowExecTraceTestSuite) TestNewWorkflowExecTrace() {
	workflowExecTrace := NewWorkflowExecTrace()
	t.NotNil(workflowExecTrace)

	var loaded bool

	loaded = workflowExecTrace.MarkExecuted("service", commonworkflow.FirstPreStartService)
	t.Equal([]string{"service:first_pre_start_service"}, workflowExecTrace.Get())
	t.True(loaded)

	loaded = workflowExecTrace.MarkExecuted("service", commonworkflow.PreStartService)
	t.Equal([]string{"service:first_pre_start_service", "service:pre_start_service"}, workflowExecTrace.Get())
	t.True(loaded)

	loaded = workflowExecTrace.MarkExecuted("service", commonworkflow.FirstPreStartService)
	t.Equal([]string{"service:first_pre_start_service", "service:pre_start_service"}, workflowExecTrace.Get())
	t.False(loaded)

	workflowExecTrace.Clear([]string{"service"}, []commonworkflow.WorkflowName{commonworkflow.FirstPreStartService})
	t.Equal([]string{"service:pre_start_service"}, workflowExecTrace.Get())
}
