package project

import (
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
)

func NewServiceWorkflows() project_types.ServiceWorkflows {
	return make(project_types.ServiceWorkflows)
}
