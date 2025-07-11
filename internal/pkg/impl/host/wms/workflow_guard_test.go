package wms

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	solo_context "github.com/spaulg/solo/internal/pkg/impl/host/context"
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"

	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/logging"
	"github.com/spaulg/solo/test/mocks/host/project"
)

func TestWorkflowGuardTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowGuardTestSuite))
}

type WorkflowGuardTestSuite struct {
	suite.Suite

	soloCtx        *solo_context.CliContext
	mockProject    *project.MockProject
	mockLogHandler *logging.MockHandler
}

func (t *WorkflowGuardTestSuite) SetupTest() {
	t.mockProject = &project.MockProject{}
	t.mockProject.On("GetMaxWorkflowTimeout", "first_pre_start").Return(30 * time.Second)

	t.mockLogHandler = &logging.MockHandler{}
	t.mockLogHandler.On("Enabled", context.Background(), mock.Anything).Return(true)

	t.soloCtx = &solo_context.CliContext{
		Project: t.mockProject,
		Logger:  slog.New(t.mockLogHandler),
		Config: &config_types.Config{
			Entrypoint: config_types.Entrypoint{
				HostEntrypointPath: test.GetTestDataFilePath("entrypoint.sh"),
			},
			GrpcServerPort: 0,
		},
	}
}

func (t *WorkflowGuardTestSuite) TestWorkflowCompleteOrSkippedEvents() {
	t.mockLogHandler.On("Handle", context.Background(), mock.Anything).Return(nil)

	guard := NewWorkflowGuard(
		t.soloCtx,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStart},
		[]string{"container1", "container2"},
	)

	guard.Publish(&wms_types.WorkflowCompleteEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   "test-service",
			ContainerName: "container1",
			WorkflowName:  workflowcommon.FirstPreStart,
		},
		Successful: true,
	})

	guard.Publish(&wms_types.WorkflowSkippedEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   "test-service",
			ContainerName: "container2",
			WorkflowName:  workflowcommon.FirstPreStart,
		},
		Successful: true,
	})

	err := guard.Wait(func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error {
		return guardCallback(workflowcommon.FirstPreStart)
	})

	t.NoError(err)
	t.mockLogHandler.AssertExpectations(t.T())
}

func (t *WorkflowGuardTestSuite) TestWorkflowEventWithUnrecognisedEventType() {
	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Received unsupported event; ignoring"
	})).Return(nil)

	guard := NewWorkflowGuard(
		t.soloCtx,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStart},
		[]string{"container1"},
	)

	guard.Publish(&wms_types.WorkflowErrorEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   "test-service",
			ContainerName: "container1",
			WorkflowName:  workflowcommon.PreStart,
		},
	})

	t.mockLogHandler.AssertExpectations(t.T())
}

func (t *WorkflowGuardTestSuite) TestWorkflowEventWithUnrecognisedWorkflow() {
	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Received event completed for workflow pre_start for container container1"
	})).Return(nil)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Workflow pre_start not registered for workflow guard"
	})).Return(nil)

	guard := NewWorkflowGuard(
		t.soloCtx,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStart},
		[]string{"container1"},
	)

	guard.Publish(&wms_types.WorkflowCompleteEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   "test-service",
			ContainerName: "container1",
			WorkflowName:  workflowcommon.PreStart,
		},
		Successful: true,
	})

	t.mockLogHandler.AssertExpectations(t.T())
}

func (t *WorkflowGuardTestSuite) TestWorkflowEventWithUnrecognisedContainer() {
	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Received event completed for workflow first_pre_start for container container2"
	})).Return(nil)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Container container2 not registered for workflow guard or channel already closed"
	})).Return(nil)

	guard := NewWorkflowGuard(
		t.soloCtx,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStart},
		[]string{"container1"},
	)

	guard.Publish(&wms_types.WorkflowCompleteEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   "test-service",
			ContainerName: "container2",
			WorkflowName:  workflowcommon.FirstPreStart,
		},
		Successful: true,
	})

	t.mockLogHandler.AssertExpectations(t.T())
}

func (t *WorkflowGuardTestSuite) TestWaitWithUnrecognisedWorkflow() {
	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Cannot wait for workflow pre_start to complete as this is not mapped"
	})).Return(nil)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Error waiting for container container1: unrecognised workflow pre_start"
	})).Return(nil)

	t.mockLogHandler.On("Handle", context.Background(), mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Encountered 1 errors while waiting for containers: [unrecognised workflow pre_start]"
	})).Return(nil)

	guard := NewWorkflowGuard(
		t.soloCtx,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStart},
		[]string{"container1"},
	)

	err := guard.Wait(func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error {
		return guardCallback(workflowcommon.PreStart)
	})

	t.ErrorContains(err, "encountered errors while waiting for containers")
	t.ErrorContains(err, "[unrecognised workflow pre_start]")
}
