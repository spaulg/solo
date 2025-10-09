package workflow_log

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/repository"

	layout2 "github.com/spaulg/solo/internal/pkg/impl/host/app/bubbletea/layout"
	messages2 "github.com/spaulg/solo/internal/pkg/impl/host/app/bubbletea/messages"
	workflow_event_history2 "github.com/spaulg/solo/internal/pkg/impl/host/app/bubbletea/models/workflow_event_history"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/bubbletea/models/workflow_event_overview"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/bubbletea/models/workflow_event_tree"
)

type activeComponent int

const (
	workflowEventHistory activeComponent = iota
	workflowEventOverview
)

type View struct {
	soloCtx       *context.CliContext
	layoutManager *layout2.Manager

	width  int
	height int

	headingWidth  int
	headingHeight int
	headingStyle  lipgloss.Style

	activeComponent       activeComponent
	workflowEventHistory  workflow_event_history2.Model
	workflowEventOverview workflow_event_overview.Model
}

func NewView(soloCtx *context.CliContext) (tea.Model, error) {
	return View{
		soloCtx: soloCtx,
		layoutManager: layout2.NewLayoutManager(
			layout2.VerticalLayoutDirection,
			[]layout2.Spec{
				layout2.NewFixedLayoutSpec(1, lipgloss.NewStyle()),
				layout2.NewPercentageLayoutSpec(50, lipgloss.NewStyle().Border(lipgloss.RoundedBorder())),
				layout2.NewFillLayoutSpec(lipgloss.NewStyle()),
			},
		),
		headingStyle:          lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("212")),
		activeComponent:       workflowEventHistory,
		workflowEventHistory:  workflow_event_history2.NewModel(repository.NewJSONFileRepository[*domain.ExecutionEvent]()),
		workflowEventOverview: workflow_event_overview.NewModel(),
	}, nil
}

func (t View) Init() tea.Cmd {
	return tea.Batch(
		t.workflowEventHistory.Init(),
		t.workflowEventOverview.Init(),
	)
}

func (t View) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = m.Width
		t.height = m.Height

		// Recalculate layout dimensions
		dimensions := t.layoutManager.CalculateDimensions(t.width, t.height)

		// Heading
		t.headingWidth = dimensions[0].ContentBoxWidth()
		t.headingHeight = dimensions[0].ContentBoxHeight()

		// Resize workflow history
		t.workflowEventHistory, cmd = t.workflowEventHistory.Update(messages2.ComponentSizeMsg{
			Width:  dimensions[1].ContentBoxWidth(),
			Height: dimensions[1].ContentBoxHeight(),
		})

		cmds = append(cmds, cmd)

		// Resize workflow output
		t.workflowEventOverview, cmd = t.workflowEventOverview.Update(messages2.ComponentSizeMsg{
			Width:  dimensions[2].ContentBoxWidth(),
			Height: dimensions[2].ContentBoxHeight(),
		})

		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if m.String() == "ctrl+c" {
			return t, tea.Quit
		}

		if m.String() == "enter" && t.activeComponent == workflowEventHistory {
			t.activeComponent = workflowEventOverview
		} else if m.String() == "esc" && t.activeComponent == workflowEventOverview {
			t.activeComponent = workflowEventHistory
		} else {
			switch t.activeComponent {
			case workflowEventHistory:
				t.workflowEventHistory, cmd = t.workflowEventHistory.Update(msg)
			case workflowEventOverview:
				t.workflowEventOverview, cmd = t.workflowEventOverview.Update(msg)
			}

			cmds = append(cmds, cmd)
		}

	case workflow_event_history2.WorkflowHistoryDataLoaded:
		t.workflowEventHistory, cmd = t.workflowEventHistory.Update(msg)
		cmds = append(cmds, cmd)

	case messages2.WorkflowEventSelected:
		t.workflowEventOverview, cmd = t.workflowEventOverview.Update(m)
		cmds = append(cmds, cmd)

	case workflow_event_tree.WorkflowEventLoaded:
		t.workflowEventOverview, cmd = t.workflowEventOverview.Update(m)
		cmds = append(cmds, cmd)
	}

	return t, tea.Batch(cmds...)
}

func (t View) View() tea.View {
	v := tea.NewView(t.layoutManager.Render(
		t.headingStyle.Width(t.headingWidth).Height(t.headingHeight).Render("Workflow Logs"),
		t.workflowEventHistory.View().Content,
		t.workflowEventOverview.View().Content,
	))

	v.AltScreen = true

	return v
}
