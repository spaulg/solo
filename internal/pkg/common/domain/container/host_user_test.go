package container

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestHostUserTestSuite(t *testing.T) {
	suite.Run(t, new(HostUserTestSuite))
}

type HostUserTestSuite struct {
	suite.Suite
}

func (t *HostUserTestSuite) TestMaskSoloEnvironmentRemovesSoloVariables() {
	environ := []string{
		"PATH=/usr/bin:/bin",
		"SOLO_HOST_UID=1000",
		"HOME=/root",
		"SOLO_HOST_GID=1000",
	}

	t.Equal([]string{"PATH=/usr/bin:/bin", "HOME=/root"}, MaskSoloEnvironment(environ))
}

func (t *HostUserTestSuite) TestMaskSoloEnvironmentDoesNotModifyInput() {
	environ := []string{"SOLO_HOST_UID=1000", "PATH=/usr/bin"}

	_ = MaskSoloEnvironment(environ)

	t.Equal([]string{"SOLO_HOST_UID=1000", "PATH=/usr/bin"}, environ)
}

func (t *HostUserTestSuite) TestMaskSoloEnvironmentMatchesExactKeyOnly() {
	environ := []string{
		"SOLO_HOST_UID_EXTRA=1",
		"MY_SOLO_HOST_UID=2",
		"SOLO_HOST_UID=1000=x",
		"SOLO_HOST_GID",
	}

	t.Equal([]string{"SOLO_HOST_UID_EXTRA=1", "MY_SOLO_HOST_UID=2"}, MaskSoloEnvironment(environ))
}

func (t *HostUserTestSuite) TestMaskSoloEnvironmentEmpty() {
	t.Empty(MaskSoloEnvironment(nil))
	t.Empty(MaskSoloEnvironment([]string{}))
}
