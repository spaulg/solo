package workflow_event_output

import (
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

type Styles struct {
	StepHeader lipgloss.Style
	StepFailed lipgloss.Style
	Muted      lipgloss.Style
	Error      lipgloss.Style
}

type KeyMap struct {
	Escape key.Binding
	Top    key.Binding
	Bottom key.Binding
}

// workflowEventOutputLoaded carries the rendered output for a selected workflow/container.
type workflowEventOutputLoaded struct {
	selection selection
	content   string
	err       error
}

type selection struct {
	executionEventName string
	workflowName       string
	containerName      string
}
