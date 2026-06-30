package workflow_event_tree

import (
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

type Styles struct {
	Node     lipgloss.Style
	Selected lipgloss.Style
}

type KeyMap struct {
	LineUp   key.Binding
	LineDown key.Binding
}
