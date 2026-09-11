package execution_event_overview

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/layout"
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/messages"
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/models/workflow_event_output"
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/models/workflow_event_tree"
	"github.com/spaulg/solo/internal/pkg/host/domain"
)

type Model struct {
	layoutManager *layout.Manager

	width  int
	height int

	workflowEventTree   tea.Model
	workflowEventOutput tea.Model
}

func NewModel(
	activeModel *layout.ActiveModel,
	containerStepMapRepository domain.ContainerStepMapRepository,
) Model {
	return Model{
		layoutManager: layout.NewLayoutManager(
			layout.HorizontalLayoutDirection,
			[]layout.Spec{
				layout.NewPercentageLayoutSpec(25, lipgloss.NewStyle().Border(lipgloss.RoundedBorder())),
				layout.NewFillLayoutSpec(lipgloss.NewStyle().Border(lipgloss.RoundedBorder())),
			},
		),
		workflowEventTree:   workflow_event_tree.NewModel(activeModel, containerStepMapRepository),
		workflowEventOutput: workflow_event_output.NewModel(),
	}
}

func (t Model) Init() tea.Cmd {
	return nil
}

func (t Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m := msg.(type) {
	case messages.ComponentSizeMsg:
		t.width, t.height = m.Width, m.Height

		cmds = append(cmds, t.layoutManager.ResizeComponents(
			t.width,
			t.height,
			&t.workflowEventTree,
			&t.workflowEventOutput,
		)...)

	default:
		// Forward message
		for _, model := range []*tea.Model{
			&t.workflowEventTree,
			&t.workflowEventOutput,
		} {
			*model, cmd = (*model).Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return t, tea.Batch(cmds...)
}

func (t Model) View() tea.View {
	return tea.NewView(t.layoutManager.Render(
		t.workflowEventTree.View().Content,
		t.workflowEventOutput.View().Content,
	))
}
