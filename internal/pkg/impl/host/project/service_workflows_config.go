package project

import (
	compose_types "github.com/spaulg/solo/internal/pkg/types/host/project/compose"
)

func NewServiceWorkflows() compose_types.ServiceWorkflows {
	return make(compose_types.ServiceWorkflows)
}
