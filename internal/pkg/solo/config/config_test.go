package config

import (
	_ "github.com/spaulg/solo/test"
	"testing"
)

func TestConfigLoading(t *testing.T) {
	config, err := NewConfig()
	if err != nil {
		t.Fatalf("Failed to load config without error: %v", err)
	}

	if config.Entrypoint != DefaultEntrypoint {
		t.Fatal("Entrypoint does not match default")
	}

	if config.StateDirectoryName != DefaultStateDirectoryName {
		t.Fatal("StateDirectoryName does not match default")
	}

	if err := config.AddConfigPath("test/data/config"); err != nil {
		t.Fatalf("failed to add config path: %v", err)
	}

	if config.Entrypoint != "/opt/bin/solo-custom-entrypoint.sh" {
		t.Fatalf("Entrypoint %s does not match overridden config %s", config.Entrypoint, "/opt/bin/solo-custom-entrypoint.sh")
	}

	if config.StateDirectoryName != "/opt/solo" {
		t.Fatalf("StateDirectoryName %s does not match overridden config %s", config.StateDirectoryName, "/opt/solo")
	}
}

func TestConfigPathNotFound(t *testing.T) {
	config, err := NewConfig()
	if err != nil {
		t.Fatalf("Failed to load config without error: %v", err)
	}

	if config.Entrypoint != DefaultEntrypoint {
		t.Fatal("Entrypoint does not match default")
	}

	if config.StateDirectoryName != DefaultStateDirectoryName {
		t.Fatal("StateDirectoryName does not match default")
	}

	if err := config.AddConfigPath("test/data/config/notfound"); err != nil {
		t.Fatalf("failed to add config path: %v", err)
	}

	if config.Entrypoint != DefaultEntrypoint {
		t.Fatal("Entrypoint does not match default")
	}

	if config.StateDirectoryName != DefaultStateDirectoryName {
		t.Fatal("StateDirectoryName does not match default")
	}
}
