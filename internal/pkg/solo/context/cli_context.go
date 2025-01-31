package context

import (
	"fmt"
	"github.com/gofrs/flock"
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"path/filepath"
)

const lockFileName = "locking_file"

type CliContext struct {
	Project        *project.Project
	Config         *config.Config
	ProjectLoadErr error
	ConfigLoadErr  error
	Logger         *slog.Logger
	lockFile       *flock.Flock
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
