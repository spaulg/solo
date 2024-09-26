package project

import (
	"github.com/spaulg/solo/internal/pkg/config"
	"github.com/spaulg/solo/internal/pkg/project_file"
	"github.com/spaulg/solo/internal/pkg/schema"
)

type Project struct {
	Config      *config.Config
	ProjectFile *project_file.ProjectFile
	ComposeFile string
	Project     *schema.Config
}
