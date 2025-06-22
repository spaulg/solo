package container

import (
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/project"
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/events"
	"github.com/stretchr/testify/suite"
)

type OrchestratorFactoryTestSuite struct {
	suite.Suite
}

func (t *OrchestratorFactoryTestSuite) TestDefaultOrchestratorFactory_Build() {
	loadedConfig := &config_types.Config{
		Orchestrator: "docker",
	}

	projectFilePath := test.GetTestDataFilePath("container/solo.yml")
	loadedProject, err := project.NewProject(projectFilePath, loadedConfig)
	t.NoError(err)

	soloCtx := &context.CliContext{
		Config:  loadedConfig,
		Project: loadedProject,
	}
	eventManager := &events.MockEventManager{}

	factory := NewOrchestratorFactory(soloCtx, eventManager)
	t.NotNil(factory)

	orchestrator, err := factory.Build()
	t.NotNil(orchestrator)
	t.NoError(err)

	soloCtx.Config.Orchestrator = "unsupported"
	orchestrator, err = factory.Build()
	t.Nil(orchestrator)
	t.Error(err)
}
