package config

import (
	"testing"

	"github.com/stretchr/testify/suite"

	config_domain "github.com/spaulg/solo/internal/pkg/impl/host/domain/config"

	"github.com/spaulg/solo/test"
)

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

type ConfigTestSuite struct {
	suite.Suite
}

func (t *ConfigTestSuite) SetupTest() {
	test.ChWorkingDirectory()
}

func (t *ConfigTestSuite) TestConfigLoading() {
	config, err := NewConfigReader()

	t.Nil(err, "Failed to load config without error: %v", err)
	t.Equal(config_domain.DefaultHostEntrypoint, config.GetConfig().Entrypoint.HostEntrypointPath, "Entrypoint does not match default")
	t.Equal(config_domain.DefaultStateDirectoryName, config.GetConfig().StateDirectoryName, "StateDirectoryName does not match default")

	err = config.AddConfigPath("test/data/config")
	t.NoError(err)

	t.Equal("/opt/bin/solo-custom-entrypoint.sh", config.GetConfig().Entrypoint.HostEntrypointPath, "Entrypoint %s does not match overridden config")
	t.Equal("/opt/solo", config.GetConfig().StateDirectoryName, "StateDirectoryName %s does not match overridden config")
}

func (t *ConfigTestSuite) TestConfigPathNotFound() {
	config, err := NewConfigReader()

	t.Nil(err, "Failed to load config without error: %v", err)

	t.Equal(config_domain.DefaultHostEntrypoint, config.GetConfig().Entrypoint.HostEntrypointPath, "Entrypoint does not match default")
	t.Equal(config_domain.DefaultStateDirectoryName, config.GetConfig().StateDirectoryName, "StateDirectoryName does not match default")

	err = config.AddConfigPath("test/data/config/notfound")
	t.NoError(err)

	t.Equal(config_domain.DefaultHostEntrypoint, config.GetConfig().Entrypoint.HostEntrypointPath, "Entrypoint does not match default")
	t.Equal(config_domain.DefaultStateDirectoryName, config.GetConfig().StateDirectoryName, "StateDirectoryName does not match default")
}
