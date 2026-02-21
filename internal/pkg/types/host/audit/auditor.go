package audit

import events_types "github.com/spaulg/solo/internal/pkg/types/host/events"

type Auditor interface {
	events_types.Subscriber

	RecordExecutionEvent(callback func() error) error
}
