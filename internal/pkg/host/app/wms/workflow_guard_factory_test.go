package wms

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	workflowcommon "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/host/domain"
	config2 "github.com/spaulg/solo/internal/pkg/host/domain/config"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/domain/compose"
	"github.com/spaulg/solo/test/mocks/logging"
)

func TestWorkflowGuardFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowGuardFactoryTestSuite))
}

type WorkflowGuardFactoryTestSuite struct {
	suite.Suite

	mockLogger     *slog.Logger
	mockConfig     *domain.Config
	mockProject    *compose.MockProject
	mockLogHandler *logging.MockHandler
}

func (t *WorkflowGuardFactoryTestSuite) SetupTest() {
	t.mockProject = &compose.MockProject{}
	t.mockProject.On("GetMaxWorkflowTimeout", "first_pre_start_container").Return(30 * time.Second)

	t.mockLogHandler = &logging.MockHandler{}
	t.mockLogger = slog.New(t.mockLogHandler)

	t.mockConfig = &domain.Config{
		Entrypoint: config2.EntrypointConfig{
			HostEntrypointPath: test.GetTestDataFilePath("entrypoint.sh"),
		},
		Workflow: config2.WorkflowConfig{
			Grpc: config2.GrpcConfig{
				ServerPort: 0,
			},
		},
	}
}

func (t *WorkflowGuardFactoryTestSuite) TestBuild() {
	workflowNames := []workflowcommon.WorkflowName{workflowcommon.FirstPreStartContainer, workflowcommon.PreStartContainer, workflowcommon.PostStartContainer}
	containerNames := []string{"container1", "container2"}

	workflowGuardFactory := NewWorkflowGuardFactory(t.mockLogger, t.mockConfig, t.mockProject)
	workflowGuard := workflowGuardFactory.Build(workflowNames, containerNames)

	t.NotNil(workflowGuard)
}
