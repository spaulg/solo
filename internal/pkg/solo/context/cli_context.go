package context

import (
	"errors"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

const lockFileName = "lock"

type CliContext struct {
	Project        *project.Project
	Config         *config.Config
	ProjectLoadErr error
	ConfigLoadErr  error
	Logger         *slog.Logger
	lockFile       *os.File
}

func (t *CliContext) ProtectWithLock(impl func(*cobra.Command, []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := t.TryLock(); err != nil {
			return err
		}

		commandErr := impl(cmd, args)

		if err := t.Unlock(); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%v: %v", err, commandErr)
			}
		}

		return commandErr
	}
}

func (t *CliContext) TryLock() error {
	var err error
	lockFile := t.Project.ResolveStateDirectory(lockFileName)
	t.lockFile, err = os.OpenFile(lockFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return errors.New("lock file exists")
	}

	return err
}

func (t *CliContext) Unlock() error {
	lockFile := t.lockFile.Name()

	if err := t.lockFile.Close(); err != nil {
		return err
	}

	return os.Remove(lockFile)
}
