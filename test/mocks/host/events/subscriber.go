package events

import (
	events_types "github.com/spaulg/solo/internal/pkg/types/host/events"
	"github.com/stretchr/testify/mock"
)

type MockSubscriber struct {
	mock.Mock
}

func (m *MockSubscriber) Publish(event events_types.Event) {
	m.Called(event)
}
