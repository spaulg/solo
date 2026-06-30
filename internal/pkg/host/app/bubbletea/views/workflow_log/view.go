package workflow_log

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/layout"
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/messages"
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/models/execution_event_history"
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/models/execution_event_overview"
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/models/workflow_event_tree"
	"github.com/spaulg/solo/internal/pkg/host/app/context"
	"github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/internal/pkg/host/infra/repository"
)

type activeComponent int

const (
	executionEventHistory activeComponent = iota
	executionEventOverview
)

type View struct {
	soloCtx       *context.CliContext
	layoutManager *layout.Manager

	width  int
	height int

	headingWidth  int
	headingHeight int
	headingStyle  lipgloss.Style

	activeComponent        activeComponent
	executionEventHistory  execution_event_history.Model
	executionEventOverview execution_event_overview.Model
}

func NewView(soloCtx *context.CliContext) (tea.Model, error) {
	executionEventRepository := repository.NewJSONFileRepository[*domain.ExecutionEvent]()
	containerStepMapRepository := repository.NewJSONFileRepository[domain.ContainerStepMap]()

	return View{
		soloCtx: soloCtx,
		layoutManager: layout.NewLayoutManager(
			layout.VerticalLayoutDirection,
			[]layout.Spec{
				layout.NewFixedLayoutSpec(1, lipgloss.NewStyle()),
				layout.NewPercentageLayoutSpec(50, lipgloss.NewStyle().Border(lipgloss.RoundedBorder())),
				layout.NewFillLayoutSpec(lipgloss.NewStyle()),
			},
		),
		headingStyle:           lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("212")),
		activeComponent:        executionEventHistory,
		executionEventHistory:  execution_event_history.NewModel(soloCtx, executionEventRepository),
		executionEventOverview: execution_event_overview.NewModel(containerStepMapRepository),
	}, nil
}

func (t View) Init() tea.Cmd {
	return tea.Batch(
		t.executionEventHistory.Init(),
		t.executionEventOverview.Init(),
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
		t.executionEventHistory, cmd = t.executionEventHistory.Update(messages.ComponentSizeMsg{
			Width:  dimensions[1].ContentBoxWidth(),
			Height: dimensions[1].ContentBoxHeight(),
		})

		cmds = append(cmds, cmd)

		// Resize workflow output
		t.executionEventOverview, cmd = t.executionEventOverview.Update(messages.ComponentSizeMsg{
			Width:  dimensions[2].ContentBoxWidth(),
			Height: dimensions[2].ContentBoxHeight(),
		})

		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if m.String() == "ctrl+c" {
			return t, tea.Quit
		}

		if m.String() == "enter" && t.activeComponent == executionEventHistory {
			t.activeComponent = executionEventOverview
		} else if m.String() == "esc" && t.activeComponent == executionEventOverview {
			t.activeComponent = executionEventHistory
		} else {
			switch t.activeComponent {
			case executionEventHistory:
				t.executionEventHistory, cmd = t.executionEventHistory.Update(msg)
			case executionEventOverview:
				t.executionEventOverview, cmd = t.executionEventOverview.Update(msg)
			}

			cmds = append(cmds, cmd)
		}

	case execution_event_history.ExecutionEventHistoryDataLoaded:
		t.executionEventHistory, cmd = t.executionEventHistory.Update(msg)
		cmds = append(cmds, cmd)

	case messages.WorkflowEventSelected:
		t.executionEventOverview, cmd = t.executionEventOverview.Update(m)
		cmds = append(cmds, cmd)

	case messages.ExecutionEventSelected:
		t.executionEventOverview, cmd = t.executionEventOverview.Update(m)
		cmds = append(cmds, cmd)

	case workflow_event_tree.WorkflowEventLoaded:
		t.executionEventOverview, cmd = t.executionEventOverview.Update(m)
		cmds = append(cmds, cmd)
	}

	return t, tea.Batch(cmds...)
}

func (t View) View() tea.View {
	v := tea.NewView(t.layoutManager.Render(
		t.headingStyle.Width(t.headingWidth).Height(t.headingHeight).Render("Workflow Logs"),
		t.executionEventHistory.View().Content,
		t.executionEventOverview.View().Content,
		// todo: add a keymap guide that changes based on the active component at the screen bottom
	))

	v.AltScreen = true

	return v
}
