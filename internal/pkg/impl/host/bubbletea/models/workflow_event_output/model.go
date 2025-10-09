package workflow_event_output

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/messages"
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
	}

	return t, nil
}

func (t Model) View() tea.View {
	return tea.NewView(lipgloss.NewStyle().Width(t.width).Height(t.height).Render("Output"))
}
