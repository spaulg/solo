package events

import workflowcommon "github.com/spaulg/solo/internal/pkg/common/wms"

type Event interface{}

type BaseEvent struct {
	ServiceName  string
	WorkflowName workflowcommon.Name
}
