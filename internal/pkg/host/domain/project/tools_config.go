package project

import (
	compose_types "github.com/spaulg/solo/internal/pkg/shared/domain/project/compose"
)

func NewTools() compose_types.Tools {
	return make(compose_types.Tools)
}
