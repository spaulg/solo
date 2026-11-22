package wms

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	workflowcommon "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/host/app/wms/wf"
	"github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/internal/pkg/host/domain/config"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/domain/compose"
	"github.com/spaulg/solo/test/mocks/logging"
)

// recordMatches reports whether the record has the given message and
// every expected attribute key with the expected string value.
func recordMatches(record slog.Record, message string, attrs map[string]string) bool {
	if record.Message != message {
		return false
	}

	found := 0
	record.Attrs(func(attr slog.Attr) bool {
		if expected, ok := attrs[attr.Key]; ok && attr.Value.String() == expected {
			found++
		}

		return true
	})

	return found == len(attrs)
}

func TestWorkflowGuardTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowGuardTestSuite))
}

type WorkflowGuardTestSuite struct {
	suite.Suite

	mockLogger     *slog.Logger
	mockConfig     *domain.Config
	mockProject    *compose.MockProject
	mockLogHandler *logging.MockHandler
}

func (t *WorkflowGuardTestSuite) SetupTest() {
	t.mockProject = &compose.MockProject{}
	t.mockProject.On("GetMaxWorkflowTimeout", "first_pre_start_container").Return(30 * time.Second)

	t.mockLogHandler = &logging.MockHandler{}
	t.mockLogger = slog.New(t.mockLogHandler)

	t.mockConfig = &domain.Config{
		Entrypoint: config.EntrypointConfig{
			HostEntrypointPath: test.GetTestDataFilePath("entrypoint.sh"),
		},
		Workflow: config.WorkflowConfig{
			Grpc: config.GrpcConfig{
				ServerPort: 0,
			},
		},
	}
}

func (t *WorkflowGuardTestSuite) TestWorkflowCompleteOrSkippedEvents() {
	t.mockLogHandler.On("Enabled", context.Background(), mock.Anything).Return(true)
	t.mockLogHandler.On("Handle", context.Background(), mock.Anything).Return(nil)

	guard := NewWorkflowGuard(
		t.mockLogger,
		t.mockConfig,
		t.mockProject,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStartContainer},
		[]string{"test_container1", "test_container2"},
	)

	guard.Publish(&wf.CompleteEvent{
		BaseWorkflowEvent: wf.BaseWorkflowEvent{
			ServiceName:       "test-service",
			ContainerName:     "container1",
			FullContainerName: "test_container1",
			WorkflowName:      workflowcommon.FirstPreStartContainer,
		},
		Successful: true,
	})

	guard.Publish(&wf.SkippedEvent{
		BaseWorkflowEvent: wf.BaseWorkflowEvent{
			ServiceName:       "test-service",
			ContainerName:     "container2",
			FullContainerName: "test_container2",
			WorkflowName:      workflowcommon.FirstPreStartContainer,
		},
		Successful: true,
	})

	err := guard.Wait(func(_ string, guardCallback func(name workflowcommon.WorkflowName) error) error {
		return guardCallback(workflowcommon.FirstPreStartContainer)
	})

	t.NoError(err)
	t.mockLogHandler.AssertExpectations(t.T())
}

func (t *WorkflowGuardTestSuite) TestWorkflowEventWithUnrecognisedEventType() {
	guard := NewWorkflowGuard(
		t.mockLogger,
		t.mockConfig,
		t.mockProject,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStartContainer},
		[]string{"test_container1"},
	)

	guard.Publish(&wf.ErrorEvent{
		BaseWorkflowEvent: wf.BaseWorkflowEvent{
			ServiceName:       "test-service",
			ContainerName:     "container1",
			FullContainerName: "test_container1",
			WorkflowName:      workflowcommon.PreStartContainer,
		},
	})

	t.mockLogHandler.AssertExpectations(t.T())
}

func (t *WorkflowGuardTestSuite) TestWorkflowEventWithUnrecognisedWorkflow() {
	t.mockLogHandler.On("Enabled", context.Background(), mock.Anything).Return(true)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return recordMatches(record, "Received event completed", map[string]string{"workflow": "pre_start_container", "container": "test_container1"})
	})).Return(nil)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return recordMatches(record, "Workflow not registered for workflow guard", map[string]string{"workflow": "pre_start_container"})
	})).Return(nil)

	guard := NewWorkflowGuard(
		t.mockLogger,
		t.mockConfig,
		t.mockProject,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStartContainer},
		[]string{"test_container1"},
	)

	guard.Publish(&wf.CompleteEvent{
		BaseWorkflowEvent: wf.BaseWorkflowEvent{
			ServiceName:       "test-service",
			ContainerName:     "container1",
			FullContainerName: "test_container1",
			WorkflowName:      workflowcommon.PreStartContainer,
		},
		Successful: true,
	})

	t.mockLogHandler.AssertExpectations(t.T())
}

func (t *WorkflowGuardTestSuite) TestWorkflowEventWithUnrecognisedContainer() {
	t.mockLogHandler.On("Enabled", context.Background(), mock.Anything).Return(true)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return recordMatches(record, "Received event completed", map[string]string{"workflow": "first_pre_start_container", "container": "test_container2"})
	})).Return(nil)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return recordMatches(record, "Container not registered for workflow guard or channel already closed", map[string]string{"container": "test_container2"})
	})).Return(nil)

	guard := NewWorkflowGuard(
		t.mockLogger,
		t.mockConfig,
		t.mockProject,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStartContainer},
		[]string{"test_container1"},
	)

	guard.Publish(&wf.CompleteEvent{
		BaseWorkflowEvent: wf.BaseWorkflowEvent{
			ServiceName:       "test-service",
			ContainerName:     "container2",
			FullContainerName: "test_container2",
			WorkflowName:      workflowcommon.FirstPreStartContainer,
		},
		Successful: true,
	})

	t.mockLogHandler.AssertExpectations(t.T())
}

func (t *WorkflowGuardTestSuite) TestWaitWithUnrecognisedWorkflow() {
	t.mockLogHandler.On("Enabled", context.Background(), mock.Anything).Return(true)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return recordMatches(record, "Cannot wait for workflow to complete as this is not mapped", map[string]string{"workflow": "pre_start_container"})
	})).Return(nil)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return recordMatches(record, "Error waiting for container", map[string]string{"container": "test_container1", "error": "unrecognised workflow pre_start_container"})
	})).Return(nil)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return recordMatches(record, "Encountered errors while waiting for containers", map[string]string{"count": "1", "errors": "[unrecognised workflow pre_start_container]"})
	})).Return(nil)

	guard := NewWorkflowGuard(
		t.mockLogger,
		t.mockConfig,
		t.mockProject,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStartContainer},
		[]string{"test_container1"},
	)

	err := guard.Wait(func(_ string, guardCallback func(name workflowcommon.WorkflowName) error) error {
		return guardCallback(workflowcommon.PreStartContainer)
	})

	t.ErrorContains(err, "encountered errors while waiting for containers")
	t.ErrorContains(err, "[unrecognised workflow pre_start_container]")
}
