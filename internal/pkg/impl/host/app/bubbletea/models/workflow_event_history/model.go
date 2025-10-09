package workflow_event_history

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/spaulg/solo/internal/pkg/impl/host/domain"

	table2 "github.com/spaulg/solo/internal/pkg/impl/host/app/bubbletea/components/table"
	messages2 "github.com/spaulg/solo/internal/pkg/impl/host/app/bubbletea/messages"
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

	executionEventRepository domain.ExecutionEventRepository
	table                    table2.Component
}

func NewModel(
	executionEventRepository domain.ExecutionEventRepository,
) Model {
	return Model{
		width:                    defaultWidth,
		height:                   defaultHeight,
		executionEventRepository: executionEventRepository,
		table: table2.NewComponent(
			table2.WithRowChangeCallback(func(selectedRow table2.RowID) func() tea.Msg {
				return func() tea.Msg {
					return messages2.WorkflowEventSelected{
						WorkflowEvent: string(selectedRow),
					}
				}
			}),
			table2.WithDimensions(defaultWidth, defaultHeight),
			table2.WithColumnOrder([]table2.ColumnID{"DATE", "EVENT", "COMMAND", "STATUS", "ERROR"}),
			table2.WithColumns(
				table2.Columns{
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
		logEntries, err := os.ReadDir(".solo/audit_logs")
		if err != nil {
			return err
		}

		rows := make([]table2.Row, 0)
		for _, entry := range logEntries {
			executionEventPath := filepath.Join(".solo/audit_logs", entry.Name(), "event.json")
			eventData, err := t.executionEventRepository.Load(executionEventPath)
			if err != nil {
				return err
			}

			status := "failed"
			if eventData.Error == "" {
				status = "succeeded"
			}

			eventDateTime, err := time.Parse("2006-01-02T15-04-05.999999999Z", entry.Name())
			if err != nil {
				return err
			}

			rows = append(rows, table2.Row{
				ID: table2.RowID(entry.Name()),
				Cells: []table2.Cell{
					{ColumnID: "DATE", Value: eventDateTime.Format(time.RFC1123)},
					{ColumnID: "EVENT", Value: eventData.CommandPath},
					{ColumnID: "COMMAND", Value: strings.Join(eventData.CommandArgs, " ")},
					{ColumnID: "STATUS", Value: status},
					{ColumnID: "ERROR", Value: eventData.Error},
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
	case messages2.ComponentSizeMsg:
		t.width, t.height = m.Width, m.Height

		t.table, cmd = t.table.Update(table2.SetComponentSizeMsg{
			Width:  t.width,
			Height: t.height,
			ColumnWidths: map[table2.ColumnID]int{
				"DATE":    max(minDateColumnWidth, (t.width/100)*20),
				"EVENT":   max(minEventColumnWidth, (t.width/100)*7),
				"COMMAND": max(minCommandColumnWidth, (t.width/100)*16),
				"STATUS":  max(minStatusColumnWidth, (t.width/100)*7),
				"ERROR":   max(minErrorColumnWidth, (t.width/100)*50),
			},
		})

		cmds = append(cmds, cmd)

	case WorkflowHistoryDataLoaded:
		t.table, cmd = t.table.Update(table2.SetRowsMsg{Rows: m.rows})
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
