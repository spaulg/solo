package wms

import (
	"iter"

	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/internal/pkg/impl/host/app/wms/wf"
)

type MockWorkflow struct {
	mock.Mock
}

func (m *MockWorkflow) StepIterator() iter.Seq[wf.Step] {
	args := m.Called()

	if s, ok := args.Get(0).(func(yield func(wf.Step) bool)); ok {
		return s
	}

	return nil
}
