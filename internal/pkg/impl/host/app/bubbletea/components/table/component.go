package table

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Component struct {
	width  int
	height int

	rowChangeCallback func(selectedRow RowID) func() tea.Msg
	columnOrder       []ColumnID
	selectedRow       RowID
	selectedIndex     int

	columns Columns
	rows    Rows

	keyMap KeyMap
	styles Styles

	dirty bool
	view  string
}

func DefaultStyles() Styles {
	return Styles{
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		Header:   lipgloss.NewStyle().Bold(true).Padding(0, 1),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
	}
}

func WithRowChangeCallback(rowChangeCallback func(selectedRow RowID) func() tea.Msg) Option {
	return func(t *Component) {
		t.rowChangeCallback = rowChangeCallback
	}
}

func DefaultKeyMap() KeyMap {
	const spacebar = " "

	return KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("b", "pgup"),
			key.WithHelp("b/pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("f", "pgdown", spacebar),
			key.WithHelp("f/pgdn", "page down"),
		),
	}
}

func WithColumnOrder(columnOrder []ColumnID) Option {
	return func(t *Component) {
		t.columnOrder = columnOrder
	}
}

func WithColumns(columns Columns) Option {
	return func(t *Component) {
		t.columns = columns
	}
}

func WithRows(rows Rows) Option {
	return func(t *Component) {
		t.rows = rows
	}
}

func WithDimensions(w, h int) Option {
	return func(t *Component) {
		t.width = w
		t.height = h
	}
}

func WithKeyMap(keymap KeyMap) Option {
	return func(t *Component) {
		t.keyMap = keymap
	}
}

func NewComponent(opts ...Option) Component {
	component := Component{
		dirty:  true,
		styles: DefaultStyles(),
		keyMap: DefaultKeyMap(),
	}

	for _, opt := range opts {
		opt(&component)
	}

	return component
}

func (t Component) Init() tea.Cmd {
	return nil
}

func (t Component) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch m := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(m, t.keyMap.LineUp):
			t.selectPrevRow()
		case key.Matches(m, t.keyMap.LineDown):
			t.selectNextRow()
		case key.Matches(m, t.keyMap.PageUp):
			// todo: handle page up
		case key.Matches(m, t.keyMap.PageDown):
			// todo: handle page down
		}

		if t.rowChangeCallback != nil {
			cmds = append(cmds, t.rowChangeCallback(t.selectedRow))
		}

	case SetComponentSizeMsg:
		for columnID, width := range m.ColumnWidths {
			t.setColumnWidth(columnID, width)
		}

		t.setDimensions(m.Width, m.Height)

	case SetRowsMsg:
		t.setRows(m.Rows)

		if t.rowChangeCallback != nil {
			cmds = append(cmds, t.rowChangeCallback(t.selectedRow))
		}
	}

	t.refreshView()

	return t, tea.Batch(cmds...)
}

func (t Component) View() tea.View {
	return tea.NewView(t.view)
}

func (t *Component) setColumnWidth(columnID ColumnID, w int) *Component {
	if column, found := t.columns[columnID]; found {
		t.dirty = true
		column.Width = w
	}

	return t
}

func (t *Component) setRows(rows Rows) *Component {
	t.dirty = true
	t.rows = rows
	t.selectedRow = rows[0].ID
	t.selectedIndex = 0

	return t
}

func (t *Component) setDimensions(w, h int) *Component {
	t.dirty = true
	t.width = w
	t.height = h

	return t
}

func (t *Component) selectNextRow() *Component {
	if len(t.rows) > t.selectedIndex+1 {
		t.dirty = true
		t.selectedIndex++
		t.selectedRow = t.rows[t.selectedIndex].ID
	}

	return t
}

func (t *Component) selectPrevRow() *Component {
	if t.selectedIndex > 0 {
		t.dirty = true
		t.selectedIndex--
		t.selectedRow = t.rows[t.selectedIndex].ID
	}

	return t
}

func (t *Component) refreshView() {
	if t.dirty {
		var view string

		// Render columns
		renderedColumns := make([]string, len(t.columns))
		for _, columnID := range t.columnOrder {
			if column, found := t.columns[columnID]; found {
				renderedColumn := t.styles.Header.Width(column.Width).Render(column.Title)
				renderedColumns = append(renderedColumns, renderedColumn)
			}
		}

		view += lipgloss.JoinHorizontal(lipgloss.Top, renderedColumns...) + "\n"

		// Render rows
		if len(t.rows) > 0 {
			for _, row := range t.rows {
				renderedCells := make([]string, len(t.columns))

				for _, cell := range row.Cells {
					if column, found := t.columns[cell.ColumnID]; found {
						style := t.styles.Cell

						renderedCell := style.Width(column.Width).Render(cell.Value)
						renderedCells = append(renderedCells, renderedCell)
					}
				}

				renderedRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedCells...)

				if row.ID == t.selectedRow {
					renderedRow = t.styles.Selected.Render(renderedRow)
				}

				view += renderedRow + "\n"
			}
		}

		t.view = lipgloss.NewStyle().Width(t.width).Height(t.height).Render(view)
		t.dirty = false
	}
}
