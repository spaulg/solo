package progress

import (
	progress2 "github.com/spaulg/solo/internal/pkg/shared/domain/container/progress"
)

type ComposeProgressEvent struct {
	ContextID         string
	Action            progress2.ComposeProgressAction
	EntityType        progress2.ProgressEntityTypeName
	FullEntityName    string
	ProjectEntityName string
	Status            progress2.ComposeProgressStatus
}
