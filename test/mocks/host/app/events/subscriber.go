package events

import (
	"github.com/stretchr/testify/mock"

	events_types "github.com/spaulg/solo/internal/pkg/types/host/app/events"
)

type MockSubscriber struct {
	mock.Mock
}

func (m *MockSubscriber) Publish(event events_types.Event) {
	m.Called(event)
}
