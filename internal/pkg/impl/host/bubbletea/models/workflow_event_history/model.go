package workflow_event_history

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/spaulg/solo/internal/pkg/impl/host/audit"
	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/components/table"
	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/messages"
)

const minEventColumnWidth = 15
const minCommandColumnWidth = 40
const minStatusColumnWidth = 20
const minDateColumnWidth = 30
const minErrorColumnWidth = 30
const defaultWidth = 0
const defaultHeight = 0

type Model struct {
	width  int
	height int

	table table.Component
}

func NewModel() Model {
	return Model{
		width:  defaultWidth,
		height: defaultHeight,
		table: table.NewComponent(
			table.WithRowChangeCallback(func(selectedRow table.RowID) func() tea.Msg {
				return func() tea.Msg {
					return messages.WorkflowEventSelected{
						WorkflowEvent: string(selectedRow),
					}
				}
			}),
			table.WithDimensions(defaultWidth, defaultHeight),
			table.WithColumnOrder([]table.ColumnID{"DATE", "EVENT", "COMMAND", "STATUS", "ERROR"}),
			table.WithColumns(
				table.Columns{
					"DATE":    {Title: "DATE", Width: minDateColumnWidth},
					"EVENT":   {Title: "EVENT", Width: minEventColumnWidth},
					"COMMAND": {Title: "COMMAND", Width: minCommandColumnWidth},
					"STATUS":  {Title: "STATUS", Width: minStatusColumnWidth},
					"ERROR":   {Title: "ERROR", Width: minErrorColumnWidth},
				},
			),
		),
	}
}

func (t Model) Init() tea.Cmd {
	return func() tea.Msg {
		logEntries, err := os.ReadDir(".solo/workflow_logs")
		if err != nil {
			return err
		}

		rows := make([]table.Row, 0)
		for _, entry := range logEntries {
			eventData, err := audit.LoadWorkflowEvent(filepath.Join(".solo/workflow_logs", entry.Name(), "event.meta.json"))
			if err != nil {
				return err
			}

			status := "failed"
			if eventData.GetError() == "" {
				status = "succeeded"
			}

			eventDateTime, err := time.Parse("2006-01-02T15-04-05.999999999Z", entry.Name())
			if err != nil {
				return err
			}

			rows = append(rows, table.Row{
				ID: table.RowID(entry.Name()),
				Cells: []table.Cell{
					{ColumnID: "DATE", Value: eventDateTime.Format(time.RFC1123)},
					{ColumnID: "EVENT", Value: eventData.GetCommandPath()},
					{ColumnID: "COMMAND", Value: strings.Join(eventData.GetCommandArgs(), " ")},
					{ColumnID: "STATUS", Value: status},
					{ColumnID: "ERROR", Value: eventData.GetError()},
				},
			})
		}

		// Reverse the list
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}

		return WorkflowHistoryDataLoaded{
			rows: rows,
		}
	}
}

func (t Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m := msg.(type) {
	case messages.ComponentSizeMsg:
		t.width, t.height = m.Width, m.Height

		t.table, cmd = t.table.Update(table.SetComponentSizeMsg{
			Width:  t.width,
			Height: t.height,
			ColumnWidths: map[table.ColumnID]int{
				"DATE":    max(minDateColumnWidth, (t.width/100)*20),
				"EVENT":   max(minEventColumnWidth, (t.width/100)*7),
				"COMMAND": max(minCommandColumnWidth, (t.width/100)*16),
				"STATUS":  max(minStatusColumnWidth, (t.width/100)*7),
				"ERROR":   max(minErrorColumnWidth, (t.width/100)*50),
			},
		})

		cmds = append(cmds, cmd)

	case WorkflowHistoryDataLoaded:
		t.table, cmd = t.table.Update(table.SetRowsMsg{Rows: m.rows})
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		t.table, cmd = t.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	return t, tea.Batch(cmds...)
}

func (t Model) View() tea.View {
	return tea.NewView(t.table.View().Content)
}
