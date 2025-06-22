package config

import (
	"github.com/spaulg/solo/test"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func (suite *ConfigTestSuite) SetupTest() {
	test.ChWorkingDirectory()
}

func (suite *ConfigTestSuite) TestConfigLoading() {
	config, err := NewConfigReader()

	suite.Nil(err, "Failed to load config without error: %v", err)
	suite.Equal(DefaultHostEntrypoint, config.GetConfig().Entrypoint.HostEntrypointPath, "Entrypoint does not match default")
	suite.Equal(DefaultStateDirectoryName, config.GetConfig().StateDirectoryName, "StateDirectoryName does not match default")

	if err := config.AddConfigPath("test/data/config"); err != nil {
		suite.Fail("failed to add config path: %v", err)
	}

	suite.Equal("/opt/bin/solo-custom-entrypoint.sh", config.GetConfig().Entrypoint.HostEntrypointPath, "Entrypoint %s does not match overridden config")
	suite.Equal("/opt/solo", config.GetConfig().StateDirectoryName, "StateDirectoryName %s does not match overridden config")
}

func (suite *ConfigTestSuite) TestConfigPathNotFound() {
	config, err := NewConfigReader()

	suite.Nil(err, "Failed to load config without error: %v", err)

	suite.Equal(DefaultHostEntrypoint, config.GetConfig().Entrypoint.HostEntrypointPath, "Entrypoint does not match default")
	suite.Equal(DefaultStateDirectoryName, config.GetConfig().StateDirectoryName, "StateDirectoryName does not match default")

	if err := config.AddConfigPath("test/data/config/notfound"); err != nil {
		suite.Fail("failed to add config path: %v", err)
	}

	suite.Equal(DefaultHostEntrypoint, config.GetConfig().Entrypoint.HostEntrypointPath, "Entrypoint does not match default")
	suite.Equal(DefaultStateDirectoryName, config.GetConfig().StateDirectoryName, "StateDirectoryName does not match default")
}
