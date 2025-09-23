package wms

import (
	"github.com/stretchr/testify/mock"

	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type MockStep struct {
	mock.Mock
}

func (m *MockStep) Trigger(
	trigger wms_types.StepTriggerLambda,
	progress wms_types.StepProgressLambda,
	complete wms_types.StepCompleteLambda,
) error {
	args := m.Called(trigger, progress, complete)
	return args.Error(0)
}

func (m *MockStep) GetId() string {
	args := m.Called()
	if s, ok := args.Get(0).(string); ok {
		return s
	} else {
		return ""
	}
}

func (m *MockStep) GetName() string {
	args := m.Called()
	if n, ok := args.Get(0).(string); ok {
		return n
	} else {
		return ""
	}
}

func (m *MockStep) GetCommand() string {
	args := m.Called()
	if c, ok := args.Get(0).(string); ok {
		return c
	} else {
		return ""
	}
}

func (m *MockStep) GetArguments() []string {
	args := m.Called()
	if a, ok := args.Get(0).([]string); ok {
		return a
	} else {
		return []string{}
	}
}

func (m *MockStep) GetShell() string {
	args := m.Called()
	if c, ok := args.Get(0).(string); ok {
		return c
	} else {
		return ""
	}
}

func (m *MockStep) GetWorkingDirectory() string {
	args := m.Called()
	if wd, ok := args.Get(0).(string); ok {
		return wd
	} else {
		return ""
	}
}
