package app

import (
	"github.com/spaulg/solo/internal/pkg/impl/host/app/event_manager/events"
)

type Auditor interface {
	events.Subscriber

	RecordExecutionEvent(callback func() error) error
}
