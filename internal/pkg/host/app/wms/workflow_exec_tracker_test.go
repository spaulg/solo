package wms

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	commonworkflow "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/test/mocks/host/infra/repository"
)

func TestWorkflowExecTrackerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowExecTrackerTestSuite))
}

type WorkflowExecTrackerTestSuite struct {
	suite.Suite

	mockRepository *repository.MockJSONFileRepository[*domain.WorkflowExecTrace]
	trackingFile   string
}

func (t *WorkflowExecTrackerTestSuite) SetupTest() {
	t.trackingFile = "workflow_exec_tracker.json"
	t.mockRepository = &repository.MockJSONFileRepository[*domain.WorkflowExecTrace]{}
}

func (t *WorkflowExecTrackerTestSuite) TestExecutionTracking() {
	var workflowList []string

	workflowExecTrace := domain.NewWorkflowExecTrace()
	t.mockRepository.On("Load", t.trackingFile).Return(workflowExecTrace, nil)

	t.mockRepository.On("Save", t.trackingFile, mock.AnythingOfType("*domain.WorkflowExecTrace")).Run(func(args mock.Arguments) {
		filePath := args.Get(0).(string)
		t.Equal(filePath, t.trackingFile)

		trace := args.Get(1).(*domain.WorkflowExecTrace)
		t.Equal(workflowList, trace.Get())
	}).Return(nil)

	executionTracker, err := NewWorkflowExecTracker(t.trackingFile, t.mockRepository)
	t.NoError(err)

	workflowList = []string{"service:first_pre_start_service"}
	loaded, err := executionTracker.MarkExecuted("service", commonworkflow.FirstPreStartService)
	t.NoError(err)
	t.True(loaded)

	workflowList = []string{"service:first_pre_start_service", "service:pre_start_service"}
	loaded2, err := executionTracker.MarkExecuted("service", commonworkflow.PreStartService)
	t.NoError(err)
	t.True(loaded2)

	workflowList = []string{"service:pre_start_service"}
	err = executionTracker.Clear([]string{"service"}, []commonworkflow.WorkflowName{commonworkflow.FirstPreStartService})
	t.NoError(err)

	t.mockRepository.AssertExpectations(t.T())
}

func (t *WorkflowExecTrackerTestSuite) TestSaveReturnsError() {
	workflowExecTrace := domain.NewWorkflowExecTrace()
	t.mockRepository.On("Load", t.trackingFile).Return(workflowExecTrace, nil)

	t.mockRepository.On("Save", t.trackingFile, mock.AnythingOfType("*domain.WorkflowExecTrace")).Return(errors.New("mock save error"))

	executionTracker, err := NewWorkflowExecTracker(t.trackingFile, t.mockRepository)
	t.NoError(err)

	loaded, err := executionTracker.MarkExecuted("service", commonworkflow.FirstPreStartService)
	t.ErrorContains(err, "mock save error")
	t.True(loaded)

	err = executionTracker.Clear([]string{"service"}, []commonworkflow.WorkflowName{commonworkflow.FirstPreStartService})
	t.ErrorContains(err, "mock save error")

	t.mockRepository.AssertExpectations(t.T())
}

func (t *WorkflowExecTrackerTestSuite) TestLoadReturnsError() {
	t.mockRepository.On("Load", t.trackingFile).Return(nil, errors.New("mock load error"))

	executionTracker, err := NewWorkflowExecTracker(t.trackingFile, t.mockRepository)
	t.ErrorContains(err, "mock load error")
	t.Nil(executionTracker)
}
