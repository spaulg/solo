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

	offset  int
	columns Columns
	rows    Rows

	keyMap KeyMap
	styles Styles

	dirty bool
	view  string
}

func DefaultStyles() Styles {
	return Styles{
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Inline(true),
		Row:      lipgloss.NewStyle().Inline(true),
		Header:   lipgloss.NewStyle().Bold(true).Padding(0, 1).Inline(true),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
		Box:      lipgloss.NewStyle(),
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

		t.styles.Box = t.styles.Box.
			Width(t.width).
			MaxWidth(t.width).
			Height(t.height).
			MaxHeight(t.height)

		t.styles.Selected = t.styles.Selected.
			MaxWidth(t.width)

		t.styles.Row = t.styles.Row.
			MaxWidth(t.width)
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
		rowChanged := false

		switch {
		case key.Matches(m, t.keyMap.LineUp):
			rowChanged = t.selectPrevRow()
		case key.Matches(m, t.keyMap.LineDown):
			rowChanged = t.selectNextRow()
		case key.Matches(m, t.keyMap.PageUp):
			rowChanged = t.selectPrevPage()
		case key.Matches(m, t.keyMap.PageDown):
			rowChanged = t.selectNextPage()
		}

		if rowChanged && t.rowChangeCallback != nil {
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

	t.styles.Box = t.styles.Box.
		Width(t.width).
		MaxWidth(t.width).
		Height(t.height).
		MaxHeight(t.height)

	t.styles.Selected = t.styles.Selected.
		MaxWidth(t.width)

	t.styles.Row = t.styles.Row.
		MaxWidth(t.width)

	return t
}

func (t *Component) selectPrevRow() bool {
	if t.selectedIndex == 0 {
		return false
	}

	// If the selected index moves before the offset,
	// move the offset back as well
	if t.selectedIndex <= t.offset {
		t.offset--
	}

	t.selectedIndex--
	t.selectedRow = t.rows[t.selectedIndex].ID
	t.dirty = true

	return true
}

func (t *Component) selectNextRow() bool {
	if len(t.rows) <= t.selectedIndex+1 {
		return false
	}

	// If selected index moves beyond the offset+height
	// move the offset forward as well
	if t.selectedIndex >= t.offset+t.height-2 {
		t.offset++
	}

	t.selectedIndex++
	t.selectedRow = t.rows[t.selectedIndex].ID
	t.dirty = true

	return true
}

func (t *Component) selectPrevPage() bool {
	// If on the first row, ignore
	if t.selectedIndex == 0 {
		return false
	}

	if len(t.rows) <= t.height {
		// Less than one page, select the first row
		t.selectedIndex = 0
	} else {
		// More than one page left, move forward one page
		t.offset -= t.height - 1
		t.selectedIndex -= t.height - 1

		// If by moving backward one page, we end up on the first page
		// adjust the offset
		if t.offset < 0 {
			t.offset = 0
		}
	}

	t.selectedRow = t.rows[t.selectedIndex].ID
	t.dirty = true

	return true
}

func (t *Component) selectNextPage() bool {
	// If on the last row, ignore
	if len(t.rows) == t.selectedIndex+1 {
		return false
	}

	if t.offset+t.height-1 >= len(t.rows) {
		// Less than one page, select the last row, or
		// on the last page, select the last row
		t.selectedIndex = len(t.rows) - 1
	} else {
		// More than one page left, move forward one page
		t.offset += t.height - 1
		t.selectedIndex += t.height - 1

		// If by moving forward one page, we end up on the last page
		// adjust the offset
		if t.offset+t.height >= len(t.rows) {
			t.offset = len(t.rows) - t.height + 1
			t.selectedIndex = len(t.rows) - 1
		}
	}

	t.selectedRow = t.rows[t.selectedIndex].ID
	t.dirty = true

	return true
}

func (t *Component) refreshView() {
	if t.dirty {
		var view string
		rowCount := 1

		// Render columns
		renderedColumnList := make([]string, len(t.columns))
		for _, columnID := range t.columnOrder {
			if column, found := t.columns[columnID]; found {
				renderedColumn := t.styles.Header.Width(column.Width).Render(column.Title)
				renderedColumnList = append(renderedColumnList, renderedColumn)
			}
		}

		renderedColumns := lipgloss.JoinHorizontal(lipgloss.Top, renderedColumnList...) + "\n"
		renderedColumns = t.styles.Row.Render(renderedColumns)
		view += renderedColumns + "\n"

		// Render rows
		if len(t.rows) > 0 {
			for _, row := range t.rows[t.offset:] {
				if rowCount >= t.height {
					break
				}

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
				} else {
					renderedRow = t.styles.Row.Render(renderedRow)
				}

				view += renderedRow + "\n"
				rowCount++
			}
		}

		t.view = t.styles.Box.Render(view)
		t.dirty = false
	}
}
