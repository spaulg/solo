package context

import (
	"errors"
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

const lockFileName = "lock"

type SoloContext struct {
	Project        *project.Project
	Config         *config.Config
	ProjectLoadErr error
	ConfigLoadErr  error
	Logger         *slog.Logger
	lockFile       *os.File
}

func (t *SoloContext) ProtectWithLock(impl func(*cobra.Command, []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := t.TryLock(); err != nil {
			return err
		}

		if err := impl(cmd, args); err != nil {
			return err
		}

		if err := t.Unlock(); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}

		return nil
	}
}

func (t *SoloContext) TryLock() error {
	var err error
	lockFile := t.Project.ResolveStateDirectory(lockFileName)
	t.lockFile, err = os.OpenFile(lockFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return errors.New("lock file exists")
	}

	return err
}

func (t *SoloContext) Unlock() error {
	lockFile := t.lockFile.Name()

	if err := t.lockFile.Close(); err != nil {
		return err
	}

	return os.Remove(lockFile)
}
