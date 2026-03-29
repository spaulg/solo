package progress

import (
	progresscommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/container/progress"
)

type ComposeProgressEvent struct {
	ContextID         string
	Action            progresscommon.ComposeProgressAction
	EntityType        progresscommon.ProgressEntityTypeName
	FullEntityName    string
	ProjectEntityName string
	Status            progresscommon.ComposeProgressStatus
}
