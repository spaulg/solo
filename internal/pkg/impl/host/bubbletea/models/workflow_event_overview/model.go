package workflow_event_overview

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/layout"
	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/messages"
	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/models/workflow_event_output"
	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/models/workflow_event_tree"
)

type activeComponent int

const (
	workflowEventTree activeComponent = iota
	workflowEventOutput
)

type Model struct {
	layoutManager *layout.Manager

	width  int
	height int

	activeComponent     activeComponent
	workflowEventTree   workflow_event_tree.Model
	workflowEventOutput workflow_event_output.Model
}

func NewModel() Model {
	return Model{
		layoutManager: layout.NewLayoutManager(
			layout.HorizontalLayoutDirection,
			[]layout.Spec{
				layout.NewPercentageLayoutSpec(25, lipgloss.NewStyle().Border(lipgloss.RoundedBorder())),
				layout.NewFillLayoutSpec(lipgloss.NewStyle().Border(lipgloss.RoundedBorder())),
			},
		),
		activeComponent:     workflowEventTree,
		workflowEventTree:   workflow_event_tree.NewModel(),
		workflowEventOutput: workflow_event_output.NewModel(),
	}
}

func (t Model) Init() tea.Cmd {
	return nil
}

func (t Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m := msg.(type) {
	case messages.ComponentSizeMsg:
		t.width, t.height = m.Width, m.Height

		// Recalculate layout dimensions
		dimensions := t.layoutManager.CalculateDimensions(t.width, t.height)

		t.workflowEventTree, cmd = t.workflowEventTree.Update(messages.ComponentSizeMsg{
			Width:  dimensions[0].ContentBoxWidth(),
			Height: dimensions[0].ContentBoxHeight(),
		})

		cmds = append(cmds, cmd)

		t.workflowEventOutput, cmd = t.workflowEventOutput.Update(messages.ComponentSizeMsg{
			Width:  dimensions[1].ContentBoxWidth(),
			Height: dimensions[1].ContentBoxHeight(),
		})

		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if m.String() == "enter" && t.activeComponent == workflowEventTree {
			t.activeComponent = workflowEventOutput
		} else if m.String() == "esc" && t.activeComponent == workflowEventOutput {
			t.activeComponent = workflowEventTree
		} else {
			switch t.activeComponent {
			case workflowEventTree:
				t.workflowEventTree, cmd = t.workflowEventTree.Update(msg)
			case workflowEventOutput:
				t.workflowEventOutput, cmd = t.workflowEventOutput.Update(msg)
			}

			cmds = append(cmds, cmd)
		}

	case messages.WorkflowEventSelected:
		t.workflowEventTree, cmd = t.workflowEventTree.Update(msg)
		cmds = append(cmds, cmd)

	case workflow_event_tree.WorkflowEventLoaded:
		t.workflowEventTree, cmd = t.workflowEventTree.Update(msg)
		cmds = append(cmds, cmd)
	}

	return t, tea.Batch(cmds...)
}

func (t Model) View() tea.View {
	return tea.NewView(t.layoutManager.Render(
		t.workflowEventTree.View().Content,
		t.workflowEventOutput.View().Content,
	))
}
