package container

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	domain_config "github.com/spaulg/solo/internal/pkg/impl/host/domain/config"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/app/events"
)

func TestOrchestratorFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(OrchestratorFactoryTestSuite))
}

type OrchestratorFactoryTestSuite struct {
	suite.Suite

	mockEventManager *events.MockEventManager
}

func (t *OrchestratorFactoryTestSuite) SetupTest() {
	t.mockEventManager = &events.MockEventManager{}
}

func (t *OrchestratorFactoryTestSuite) TestOrchestratorFactorySuccess() {
	loadedConfig := &domain.Config{
		Orchestration: domain_config.OrchestrationConfig{
			SearchOrder: []string{"docker"},
			Orchestrators: map[string]domain_config.OrchestratorConfig{
				"docker": {
					Binary: "docker",
				},
			},
		},
	}

	projectFilePath := test.GetTestDataFilePath("container/solo.yml")
	loadedProject, err := domain.NewProject(projectFilePath, loadedConfig, []string{})
	t.NoError(err)

	soloCtx := &context.CliContext{
		Config:  loadedConfig,
		Project: loadedProject,
	}

	factory := NewOrchestratorFactory(soloCtx, t.mockEventManager)
	t.NotNil(factory)

	orchestrator, err := factory.Build()
	t.NotNil(orchestrator)
	t.NoError(err)

}

func (t *OrchestratorFactoryTestSuite) TestOrchestratorFactoryFailure() {
	loadedConfig := &domain.Config{
		Orchestration: domain_config.OrchestrationConfig{
			SearchOrder: []string{},
		},
	}

	projectFilePath := test.GetTestDataFilePath("container/solo.yml")
	loadedProject, err := domain.NewProject(projectFilePath, loadedConfig, []string{})
	t.NoError(err)

	soloCtx := &context.CliContext{
		Config:  loadedConfig,
		Project: loadedProject,
	}

	factory := NewOrchestratorFactory(soloCtx, t.mockEventManager)
	orchestrator, err := factory.Build()
	t.Nil(orchestrator)
	t.Error(err)
}
