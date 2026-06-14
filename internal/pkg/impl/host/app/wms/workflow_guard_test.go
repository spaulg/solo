package wms

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/wms/wf"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	domain_config_types "github.com/spaulg/solo/internal/pkg/impl/host/domain/config"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/domain/compose"
	"github.com/spaulg/solo/test/mocks/logging"
)

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
		Entrypoint: domain_config_types.EntrypointConfig{
			HostEntrypointPath: test.GetTestDataFilePath("entrypoint.sh"),
		},
		Workflow: domain_config_types.WorkflowConfig{
			Grpc: domain_config_types.GrpcConfig{
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
		return record.Message == "Received event completed for workflow pre_start_container for container test_container1"
	})).Return(nil)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Workflow pre_start_container not registered for workflow guard"
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
		return record.Message == "Received event completed for workflow first_pre_start_container for container test_container2"
	})).Return(nil)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Container test_container2 not registered for workflow guard or channel already closed"
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
		return record.Message == "Cannot wait for workflow pre_start_container to complete as this is not mapped"
	})).Return(nil)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Error waiting for container test_container1: unrecognised workflow pre_start_container"
	})).Return(nil)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Encountered 1 errors while waiting for containers: [unrecognised workflow pre_start_container]"
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
