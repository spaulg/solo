package wms

import (
	events_types "github.com/spaulg/solo/internal/pkg/types/host/app/events"
)

type WorkflowLogWriter interface {
	events_types.Subscriber

	RecordEvent(callback func() error) error
}
