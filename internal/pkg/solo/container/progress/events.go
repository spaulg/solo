package progress

import progresscommon "github.com/spaulg/solo/internal/pkg/common/container/progress"

type ComposeProgressEvent struct {
	ContextId         string
	Action            progresscommon.ComposeProgressAction
	EntityType        progresscommon.ProgressEntityTypeName
	FullEntityName    string
	ProjectEntityName string
	Status            progresscommon.ComposeProgressStatus
}
