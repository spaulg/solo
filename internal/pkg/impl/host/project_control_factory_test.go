package host

import (
	"testing"
	"log/slog"

	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	"github.com/spaulg/solo/test/mocks/host/logging"
	"github.com/spaulg/solo/test/mocks/host/project"
)

func TestProjectControlFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectControlFactoryTestSuite))
}

type ProjectControlFactoryTestSuite struct {
	suite.Suite

	soloCtx        *context.CliContext
	mockProject    *project.MockProject
	mockLogHandler *logging.MockHandler
}

func (t *ProjectControlFactoryTestSuite) SetupTest() {
	t.mockProject = &project.MockProject{}

	t.mockLogHandler = &logging.MockHandler{}
	t.mockLogHandler.On("Enabled").Return(true)

	t.soloCtx = &context.CliContext{
		Project: t.mockProject,
		Logger:  slog.New(t.mockLogHandler),
		Config:  &config_types.Config{},
	}
}

func (t *ProjectControlFactoryTestSuite) TestBuild() {
	projectControl, err := ProjectControlFactory(t.soloCtx)

	t.Nil(err)
	t.NotNil(projectControl)
}
