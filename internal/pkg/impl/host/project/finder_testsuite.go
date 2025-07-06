package project

import (
	"path"

	"github.com/stretchr/testify/suite"

	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	"github.com/spaulg/solo/test"
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

	project, err := FindProject(startPath, t.config, []string{})

	t.NoError(err)
	t.NotNil(project)

	t.Equal(expectedProjectPath, project.GetFilePath())
}

func (t *FinderTestSuite) TestProjectFileNotFoundBeforeFsRoot() {
	project, err := FindProject(t.T().TempDir(), t.config, []string{})

	t.Error(err)
	t.ErrorContains(err, "filesystem root reached, project file not found")

	t.Nil(project)
}
