package config

import (
	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/test"
)

type ConfigTestSuite struct {
	suite.Suite
}

func (t *ConfigTestSuite) SetupTest() {
	test.ChWorkingDirectory()
}

func (t *ConfigTestSuite) TestConfigLoading() {
	config, err := NewConfigReader()

	t.Nil(err, "Failed to load config without error: %v", err)
	t.Equal(DefaultHostEntrypoint, config.GetConfig().Entrypoint.HostEntrypointPath, "Entrypoint does not match default")
	t.Equal(DefaultStateDirectoryName, config.GetConfig().StateDirectoryName, "StateDirectoryName does not match default")

	err = config.AddConfigPath("test/data/config")
	t.NoError(err)

	t.Equal("/opt/bin/solo-custom-entrypoint.sh", config.GetConfig().Entrypoint.HostEntrypointPath, "Entrypoint %s does not match overridden config")
	t.Equal("/opt/solo", config.GetConfig().StateDirectoryName, "StateDirectoryName %s does not match overridden config")
}

func (t *ConfigTestSuite) TestConfigPathNotFound() {
	config, err := NewConfigReader()

	t.Nil(err, "Failed to load config without error: %v", err)

	t.Equal(DefaultHostEntrypoint, config.GetConfig().Entrypoint.HostEntrypointPath, "Entrypoint does not match default")
	t.Equal(DefaultStateDirectoryName, config.GetConfig().StateDirectoryName, "StateDirectoryName does not match default")

	err = config.AddConfigPath("test/data/config/notfound")
	t.NoError(err)

	t.Equal(DefaultHostEntrypoint, config.GetConfig().Entrypoint.HostEntrypointPath, "Entrypoint does not match default")
	t.Equal(DefaultStateDirectoryName, config.GetConfig().StateDirectoryName, "StateDirectoryName does not match default")
}
