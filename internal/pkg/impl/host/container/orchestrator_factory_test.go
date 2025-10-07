package container

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/project"
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/events"
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
	loadedConfig := &config_types.Config{
		Orchestration: config_types.OrchestrationConfig{
			SearchOrder: []string{"docker"},
			Orchestrators: map[string]config_types.OrchestratorConfig{
				"docker": {
					Binary: "docker",
				},
			},
		},
	}

	projectFilePath := test.GetTestDataFilePath("container/solo.yml")
	loadedProject, err := project.NewProject(projectFilePath, loadedConfig, []string{})
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
	loadedConfig := &config_types.Config{
		Orchestration: config_types.OrchestrationConfig{
			SearchOrder: []string{},
		},
	}

	projectFilePath := test.GetTestDataFilePath("container/solo.yml")
	loadedProject, err := project.NewProject(projectFilePath, loadedConfig, []string{})
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
