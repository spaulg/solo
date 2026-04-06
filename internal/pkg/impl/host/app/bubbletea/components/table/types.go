package table

import (
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

type RowID string
type ColumnID string

type Column struct {
	Title string
	Width int
}

type Cell struct {
	ColumnID ColumnID
	Value    string
}

type Row struct {
	ID    RowID
	Cells []Cell
}

type Columns map[ColumnID]*Column
type Rows []Row

type Styles struct {
	Header   lipgloss.Style
	Cell     lipgloss.Style
	Selected lipgloss.Style
}

type Option func(component *Component)

type KeyMap struct {
	LineUp       key.Binding
	LineDown     key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	GotoTop      key.Binding
	GotoBottom   key.Binding
}
