package config

import (
	_ "github.com/spaulg/solo/test"
	asserter "github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigLoading(t *testing.T) {
	assert := asserter.New(t)
	config, err := NewConfig()

	assert.Nil(err, "Failed to load config without error: %v", err)
	assert.Equal(DefaultHostEntrypoint, config.Entrypoint.HostEntrypointPath, "Entrypoint does not match default")
	assert.Equal(DefaultStateDirectoryName, config.StateDirectoryName, "StateDirectoryName does not match default")

	if err := config.AddConfigPath("test/data/config"); err != nil {
		assert.Fail("failed to add config path: %v", err)
	}

	assert.Equal("/opt/bin/solo-custom-entrypoint.sh", config.Entrypoint.HostEntrypointPath, "Entrypoint %s does not match overridden config")
	assert.Equal("/opt/solo", config.StateDirectoryName, "StateDirectoryName %s does not match overridden config")
}

func TestConfigPathNotFound(t *testing.T) {
	assert := asserter.New(t)
	config, err := NewConfig()

	assert.Nil(err, "Failed to load config without error: %v", err)

	assert.Equal(DefaultHostEntrypoint, config.Entrypoint.HostEntrypointPath, "Entrypoint does not match default")
	assert.Equal(DefaultStateDirectoryName, config.StateDirectoryName, "StateDirectoryName does not match default")

	if err := config.AddConfigPath("test/data/config/notfound"); err != nil {
		assert.Fail("failed to add config path: %v", err)
	}

	assert.Equal(DefaultHostEntrypoint, config.Entrypoint.HostEntrypointPath, "Entrypoint does not match default")
	assert.Equal(DefaultStateDirectoryName, config.StateDirectoryName, "StateDirectoryName does not match default")
}
