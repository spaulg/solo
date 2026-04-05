package app

import (
	events_types "github.com/spaulg/solo/internal/pkg/shared/app/events"
)

type Auditor interface {
	events_types.Subscriber

	RecordExecutionEvent(callback func() error) error
}
