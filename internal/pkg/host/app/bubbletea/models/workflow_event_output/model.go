package workflow_event_output

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/messages"
)

type Model struct {
	width  int
	height int
}

func NewModel() Model {
	return Model{}
}

func (t Model) Init() tea.Cmd {
	return nil
}

func (t Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch m := msg.(type) {
	case messages.ComponentSizeMsg:
		t.width, t.height = m.Width, m.Height

	case messages.WorkflowEventSelected:
		// todo: load the output from logs for the specified container and workflow name
		// m.ExecutionEventName
		// m.ContainerName
		// m.WorkflowName

		// todo: load .solo/audit_logs/{ExecutionEventName}/{WorkflowName}/workflow.meta.json
		// todo: use ContainerName in the loaded json to get executed steps
		// todo: for each step, load the meta.json file and load the .out files
	}

	return t, nil
}

func (t Model) View() tea.View {
	return tea.NewView(
		lipgloss.NewStyle().
			Width(t.width).
			MaxWidth(t.width).
			Height(t.height).
			MaxHeight(t.height).
			Render("Output"),
	)
}
