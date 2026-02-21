package audit

import (
	events_types "github.com/spaulg/solo/internal/pkg/types/host/app/events"
)

type Auditor interface {
	events_types.Subscriber

	RecordExecutionEvent(callback func() error) error
}
