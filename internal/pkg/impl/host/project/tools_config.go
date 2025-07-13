package project

import (
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
)

func NewTools() project_types.Tools {
	return make(project_types.Tools, 0)
}
