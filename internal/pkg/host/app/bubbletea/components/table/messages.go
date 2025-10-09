package table

type SetRowsMsg struct {
	Rows Rows
}

type SetComponentSizeMsg struct {
	Width        int
	Height       int
	ColumnWidths map[ColumnID]int
}
