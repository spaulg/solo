package wms

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/wms"
)

func TestWorkflowExecTrackerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowExecTrackerTestSuite))
}

type WorkflowExecTrackerTestSuite struct {
	suite.Suite
}

func (t *WorkflowExecTrackerTestSuite) TestSaveTracking() {
	trackingFile := t.T().TempDir() + "/workflow_exec_tracker.json"
	executionTracker, err := LoadWorkflowExecTracker(trackingFile)
	t.NoError(err)

	loaded, err := executionTracker.MarkExecuted("service", commonworkflow.FirstPreStartService)
	t.NoError(err)

	t.True(loaded)
	t.FileExists(trackingFile)

	loaded2, err := executionTracker.MarkExecuted("service", commonworkflow.PreStartService)
	t.NoError(err)

	t.True(loaded2)

	data, err := os.ReadFile(trackingFile)
	t.NoError(err)

	t.Contains(string(data), `["service:first_pre_start_service","service:pre_start_service"]`)

	loaded3, err := executionTracker.MarkExecuted("service", commonworkflow.PostStartService)
	t.NoError(err)

	t.True(loaded3)

	data, err = os.ReadFile(trackingFile)
	t.NoError(err)

	t.Contains(string(data), `["service:first_pre_start_service","service:pre_start_service","service:post_start_service"]`)

	loaded4, err := executionTracker.MarkExecuted("service", commonworkflow.PostStartService)
	t.NoError(err)

	t.False(loaded4)

	data, err = os.ReadFile(trackingFile)
	t.NoError(err)

	t.Contains(string(data), `["service:first_pre_start_service","service:pre_start_service","service:post_start_service"]`)
}

func (t *WorkflowExecTrackerTestSuite) TestLoadTracking() {
	trackingFile := t.T().TempDir() + "/workflow_exec_tracker.json"

	err := os.WriteFile(trackingFile, []byte(`["service:first_pre_start_service","service:pre_start_service"]`), 0644)
	t.NoError(err)

	executionTracker, err := LoadWorkflowExecTracker(trackingFile)
	t.NoError(err)

	loaded3, err := executionTracker.MarkExecuted("service", commonworkflow.PostStartService)
	t.NoError(err)

	t.True(loaded3)

	data, err := os.ReadFile(trackingFile)
	t.NoError(err)

	t.Contains(string(data), `["service:first_pre_start_service","service:pre_start_service","service:post_start_service"]`)
}
