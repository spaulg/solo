package project

import (
	compose_types "github.com/spaulg/solo/internal/pkg/types/host/domain/project/compose"
)

func NewTools() compose_types.Tools {
	return make(compose_types.Tools)
}
