package progress

import progresscommon "github.com/spaulg/solo/internal/pkg/impl/common/container/progress"

type ComposeProgressEvent struct {
	ContextId         string
	Action            progresscommon.ComposeProgressAction
	EntityType        progresscommon.ProgressEntityTypeName
	FullEntityName    string
	ProjectEntityName string
	Status            progresscommon.ComposeProgressStatus
}
