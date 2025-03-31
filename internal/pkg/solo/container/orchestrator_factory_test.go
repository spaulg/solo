package container

import (
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"github.com/spaulg/solo/test"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MockEventManager struct{}

func (m *MockEventManager) Subscribe(eventSubscriber events.Subscriber)   {}
func (m *MockEventManager) Unsubscribe(eventSubscriber events.Subscriber) {}
func (m *MockEventManager) Publish(data events.Event)                     {}
func (m *MockEventManager) Wait()                                         {}

type OrchestratorFactoryTestSuite struct {
	suite.Suite
}

func TestOrchestratorFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(OrchestratorFactoryTestSuite))
}

func (t *OrchestratorFactoryTestSuite) TestDefaultOrchestratorFactory_Build() {
	loadedConfig := &config.Config{
		Orchestrator: "docker",
	}

	projectFilePath := test.GetTestDataFilePath("container/solo.yml")
	loadedProject, err := project.NewProject(projectFilePath, loadedConfig)
	t.NoError(err)

	soloCtx := &context.CliContext{
		Config:  loadedConfig,
		Project: loadedProject,
	}
	eventManager := &MockEventManager{}

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
