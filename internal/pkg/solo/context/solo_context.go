package context

import (
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"log/slog"
)

type SoloContext struct {
	Project        *project.Project
	Config         *config.Config
	ProjectLoadErr error
	ConfigLoadErr  error
	Logger         *slog.Logger
}
