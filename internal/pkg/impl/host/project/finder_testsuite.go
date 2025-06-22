package project

import (
	"path"

	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	"github.com/spaulg/solo/test"
	"github.com/stretchr/testify/suite"
)

type FinderTestSuite struct {
	suite.Suite

	config *config_types.Config
}

func (t *FinderTestSuite) SetupTest() {
	t.config = &config_types.Config{}
}

func (t *FinderTestSuite) TestFindProjectFile() {
	startPath := test.GetTestDataFilePath("project/foo/bar/baz")
	expectedProjectPath := path.Join(path.Dir(path.Dir(path.Dir(startPath))), DefaultProjectFileName)

	project, err := FindProject(startPath, t.config)

	t.NoError(err)
	t.NotNil(project)

	t.Equal(expectedProjectPath, project.GetFilePath())
}

func (t *FinderTestSuite) TestProjectFileNotFoundBeforeFsRoot() {
	project, err := FindProject(t.T().TempDir(), t.config)

	t.Error(err)
	t.ErrorContains(err, "filesystem root reached, project file not found")

	t.Nil(project)
}
