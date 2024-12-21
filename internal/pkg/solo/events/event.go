package events

import workflowcommon "github.com/spaulg/solo/internal/pkg/common/wms"

type Event struct {
	ServiceName  string
	WorkflowName workflowcommon.Name
}
