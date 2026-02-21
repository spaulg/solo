package wms

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	solo_context "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	domain_config_types "github.com/spaulg/solo/internal/pkg/impl/host/domain/config"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/app/wms"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/domain/project"
	"github.com/spaulg/solo/test/mocks/logging"
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
	t.mockProject.On("GetMaxWorkflowTimeout", "first_pre_start_container").Return(30 * time.Second)

	t.mockLogHandler = &logging.MockHandler{}

	t.soloCtx = &solo_context.CliContext{
		Project: t.mockProject,
		Logger:  slog.New(t.mockLogHandler),
		Config: &domain.Config{
			Entrypoint: domain_config_types.EntrypointConfig{
				HostEntrypointPath: test.GetTestDataFilePath("entrypoint.sh"),
			},
			Workflow: domain_config_types.WorkflowConfig{
				Grpc: domain_config_types.GrpcConfig{
					ServerPort: 0,
				},
			},
		},
	}
}

func (t *WorkflowGuardTestSuite) TestWorkflowCompleteOrSkippedEvents() {
	t.mockLogHandler.On("Enabled", context.Background(), mock.Anything).Return(true)
	t.mockLogHandler.On("Handle", context.Background(), mock.Anything).Return(nil)

	guard := NewWorkflowGuard(
		t.soloCtx,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStartContainer},
		[]string{"test_container1", "test_container2"},
	)

	guard.Publish(&wms_types.WorkflowCompleteEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:       "test-service",
			ContainerName:     "container1",
			FullContainerName: "test_container1",
			WorkflowName:      workflowcommon.FirstPreStartContainer,
		},
		Successful: true,
	})

	guard.Publish(&wms_types.WorkflowSkippedEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
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
		t.soloCtx,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStartContainer},
		[]string{"test_container1"},
	)

	guard.Publish(&wms_types.WorkflowErrorEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
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
		t.soloCtx,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStartContainer},
		[]string{"test_container1"},
	)

	guard.Publish(&wms_types.WorkflowCompleteEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
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
		t.soloCtx,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStartContainer},
		[]string{"test_container1"},
	)

	guard.Publish(&wms_types.WorkflowCompleteEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
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
		t.soloCtx,
		[]workflowcommon.WorkflowName{workflowcommon.FirstPreStartContainer},
		[]string{"test_container1"},
	)

	err := guard.Wait(func(_ string, guardCallback func(name workflowcommon.WorkflowName) error) error {
		return guardCallback(workflowcommon.PreStartContainer)
	})

	t.ErrorContains(err, "encountered errors while waiting for containers")
	t.ErrorContains(err, "[unrecognised workflow pre_start_container]")
}
