package wms

import (
	"github.com/stretchr/testify/mock"
	"iter"

	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type MockOrchestrator struct {
	mock.Mock
}

func (m *MockOrchestrator) StepIterator() iter.Seq[wms_types.Step] {
	args := m.Called()

	if s, ok := args.Get(0).(func(yield func(wms_types.Step) bool)); ok {
		return s
	} else {
		return nil
	}
}
