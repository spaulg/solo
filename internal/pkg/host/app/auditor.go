package app

import (
	"github.com/spaulg/solo/internal/pkg/host/domain/events"
)

type Auditor interface {
	events.Subscriber

	RecordExecutionEvent(callback func() error) error
}
