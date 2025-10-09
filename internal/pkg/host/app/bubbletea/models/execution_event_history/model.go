package execution_event_history

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/spaulg/solo/internal/pkg/host/app/context"
	"github.com/spaulg/solo/internal/pkg/host/domain"

	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/components/table"
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/messages"
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

	soloCtx                  *context.CliContext
	executionEventRepository domain.ExecutionEventRepository
	table                    table.Component
}

func NewModel(
	soloCtx *context.CliContext,
	executionEventRepository domain.ExecutionEventRepository,
) Model {
	return Model{
		soloCtx:                  soloCtx,
		width:                    defaultWidth,
		height:                   defaultHeight,
		executionEventRepository: executionEventRepository,
		table: table.NewComponent(
			table.WithRowChangeCallback(func(selectedRow table.RowID) func() tea.Msg {
				return func() tea.Msg {
					return messages.ExecutionEventSelected{
						ExecutionEvent: string(selectedRow),
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
		rows := make([]table.Row, 0)
		basePath := t.soloCtx.Project.ResolveStateDirectory("audit_logs")

		for id, eventData := range t.executionEventRepository.ReverseWalk(basePath, "event.json") {
			status := "failed"
			if eventData.Error == "" {
				status = "succeeded"
			}

			eventDateTime, err := time.Parse("2006-01-02T15-04-05.999999999Z", id)
			if err != nil {
				return err
			}

			rows = append(rows, table.Row{
				ID: table.RowID(id),
				Cells: []table.Cell{
					{ColumnID: "DATE", Value: eventDateTime.Format(time.RFC1123)},
					{ColumnID: "EVENT", Value: eventData.CommandPath},
					{ColumnID: "COMMAND", Value: strings.Join(eventData.CommandArgs, " ")},
					{ColumnID: "STATUS", Value: status},
					{ColumnID: "ERROR", Value: eventData.Error},
				},
			})
		}

		return ExecutionEventHistoryDataLoaded{
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

	case ExecutionEventHistoryDataLoaded:
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
