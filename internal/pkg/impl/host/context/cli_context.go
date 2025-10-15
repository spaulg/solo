package context

import (
	"fmt"
	"log/slog"
	"os"
	"path"
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

	Project        project_types.Project
	ProjectLoadErr error

	Config        *config_types.Config
	ConfigLoadErr error

	Logger *slog.Logger

	lockFile *flock.Flock

	CommandPath string
	CommandArgs []string

	TriggerDateTime time.Time
}

func LoadCliContext() (*CliContext, error) {
	loadedConfigReader, configLoadErr := config.NewConfigReader()

	context := &CliContext{
		configReader:    loadedConfigReader,
		ConfigLoadErr:   configLoadErr,
		Logger:          slog.New(logging.NewBlackHoleHandler()),
		TriggerDateTime: time.Now(),
	}

	if configLoadErr == nil {
		context.Config = context.configReader.GetConfig()

		loadedProject, projectLoadErr := project.FindProject("./", context.Config, []string{})
		if projectLoadErr == nil {
			context.ConfigLoadErr = context.configReader.AddConfigPath(loadedProject.GetDirectory())
			context.Config = context.configReader.GetConfig()
		}

		context.Project = loadedProject
		context.ProjectLoadErr = projectLoadErr

		// If logging is enabled override the default logger
		if context.Config.Logging.Enabled && context.Project != nil {
			stateDirectory := path.Join(context.Project.GetStateDirectoryRoot(), "cli", "logs")
			if err := os.MkdirAll(stateDirectory, 0755); err != nil {
				return nil, fmt.Errorf("failed to create log directory: %v", err)
			}

			logFileName := path.Join(stateDirectory, time.Now().Format("2006-01-02.log"))

			builder := logging.NewLogHandlerBuilder()
			handler, err := builder.
				WithLogFilePath(logFileName).
				WithLogLevel(context.Config.Logging.Level).
				WithLogHandlerName(context.Config.Logging.Handler).
				Build()

			if err != nil {
				return nil, err
			}

			context.Logger = slog.New(handler)
		}
	}

	return context, nil
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
