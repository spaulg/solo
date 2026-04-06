package wms

import (
	"iter"

	"github.com/stretchr/testify/mock"

	wms_types "github.com/spaulg/solo/internal/pkg/shared/app/wms"
)

type MockWorkflow struct {
	mock.Mock
}

func (m *MockWorkflow) StepIterator() iter.Seq[wms_types.Step] {
	args := m.Called()

	if s, ok := args.Get(0).(func(yield func(wms_types.Step) bool)); ok {
		return s
	}

	return nil
}
