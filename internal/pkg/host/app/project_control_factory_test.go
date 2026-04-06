package app

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/internal/pkg/host/app/context"
	"github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/test/mocks/host/domain/project"
	"github.com/spaulg/solo/test/mocks/logging"
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

	tmpDir := t.T().TempDir()
	t.mockProject.On("GetStateDirectoryRoot").Return(tmpDir)

	t.mockLogHandler = &logging.MockHandler{}
	t.mockLogHandler.On("Enabled").Return(true)

	t.soloCtx = &context.CliContext{
		Project: t.mockProject,
		Logger:  slog.New(t.mockLogHandler),
		Config:  &domain.Config{},
	}
}

func (t *ProjectControlFactoryTestSuite) TestBuild() {
	projectControl, err := ProjectControlFactory(t.soloCtx)

	t.Nil(err)
	t.NotNil(projectControl)
}
