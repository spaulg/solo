package context

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/common/logging"
	"github.com/spaulg/solo/internal/pkg/impl/host/config"
	"github.com/spaulg/solo/internal/pkg/impl/host/project"
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
)

const lockFileName = "locking_file"

type CliContext struct {
	configReader config_types.ConfigReader

	Project         project_types.Project
	Config          *config_types.Config
	ProjectLoadErr  error
	ConfigLoadErr   error
	Logger          *slog.Logger
	lockFile        *flock.Flock
	TriggerDateTime time.Time
	Profiles        []string
}

func LoadCliContext() *CliContext {
	loadedConfigReader, configLoadErr := config.NewConfigReader()

	context := &CliContext{
		configReader:    loadedConfigReader,
		ConfigLoadErr:   configLoadErr,
		Logger:          slog.New(logging.NewBlackHoleHandler()),
		TriggerDateTime: time.Now(),
	}

	if configLoadErr == nil {
		context.Config = context.configReader.GetConfig()
	}

	return context
}

func (t *CliContext) ReloadProject() {
	loadedProject, projectLoadErr := project.FindProject("./", t.Config, t.Profiles)
	if projectLoadErr == nil {
		t.ConfigLoadErr = t.configReader.AddConfigPath(loadedProject.GetDirectory())
		t.Config = t.configReader.GetConfig()
	}

	t.Project = loadedProject
	t.ProjectLoadErr = projectLoadErr
}

func (t *CliContext) ProtectWithLock(impl func(*cobra.Command, []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := t.TryLock(); err != nil {
			return err
		}

		return impl(cmd, args)
	}
}

func (t *CliContext) TryLock() error {
	// Create the lock file if it does not already exist
	lockFileName := t.Project.ResolveStateDirectory(lockFileName)

	if err := os.MkdirAll(filepath.Dir(lockFileName), 0700); err != nil {
		return err
	}

	lockFile, err := os.OpenFile(lockFileName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	} else {
		_ = lockFile.Close()
	}

	// Lock the locking file with an optimistic exclusive write lock
	t.lockFile = flock.New(lockFileName)
	locked, err := t.lockFile.TryLock()
	if err != nil {
		return err
	}

	if !locked {
		return fmt.Errorf("could not acquire lock")
	}

	return nil
}
