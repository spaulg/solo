package container

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain/compose"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain/config"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/app/events"
	"github.com/spaulg/solo/test/mocks/logging"
)

func TestOrchestratorFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(OrchestratorFactoryTestSuite))
}

type OrchestratorFactoryTestSuite struct {
	suite.Suite

	mockEventManager *events.MockEventManager
	mockLogHandler   *logging.MockHandler
	mockLogger       *slog.Logger
}

func (t *OrchestratorFactoryTestSuite) SetupTest() {
	t.mockEventManager = &events.MockEventManager{}
	t.mockLogHandler = &logging.MockHandler{}
	t.mockLogger = slog.New(t.mockLogHandler)
}

func (t *OrchestratorFactoryTestSuite) TestOrchestratorFactorySuccess() {
	loadedConfig := &domain.Config{
		Orchestration: config.OrchestrationConfig{
			SearchOrder: []string{"docker"},
			Orchestrators: map[string]config.OrchestratorConfig{
				"docker": {
					Binary: "docker",
				},
			},
		},
	}

	projectFilePath := test.GetTestDataFilePath("container/solo.yml")
	loadedProject, err := compose.NewProject(projectFilePath, loadedConfig, []string{})
	t.NoError(err)

	factory := NewOrchestratorFactory(t.mockLogger, loadedConfig, loadedProject, t.mockEventManager)
	t.NotNil(factory)

	orchestrator, err := factory.Build()
	t.NotNil(orchestrator)
	t.NoError(err)

}

func (t *OrchestratorFactoryTestSuite) TestOrchestratorFactoryFailure() {
	loadedConfig := &domain.Config{
		Orchestration: config.OrchestrationConfig{
			SearchOrder: []string{},
		},
	}

	projectFilePath := test.GetTestDataFilePath("container/solo.yml")
	loadedProject, err := compose.NewProject(projectFilePath, loadedConfig, []string{})
	t.NoError(err)

	factory := NewOrchestratorFactory(t.mockLogger, loadedConfig, loadedProject, t.mockEventManager)
	orchestrator, err := factory.Build()
	t.Nil(orchestrator)
	t.Error(err)
}
