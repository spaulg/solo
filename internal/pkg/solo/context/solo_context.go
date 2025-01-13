package context

import (
	"errors"
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"log/slog"
	"os"
)

type SoloContext struct {
	Project        *project.Project
	Config         *config.Config
	ProjectLoadErr error
	ConfigLoadErr  error
	Logger         *slog.Logger
	lockFile       *os.File
}

func (t *SoloContext) TryLock() error {
	var err error
	lockFile := t.Project.ResolveStateDirectory("lock")
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
