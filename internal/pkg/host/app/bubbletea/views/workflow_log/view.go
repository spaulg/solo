package workflow_log

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/components/heading"
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/layout"
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/models/execution_event_history"
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/models/execution_event_overview"
	"github.com/spaulg/solo/internal/pkg/host/app/context"
	"github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/internal/pkg/host/infra/repository"
)

type View struct {
	soloCtx       *context.CliContext
	layoutManager *layout.Manager

	width  int
	height int

	heading                tea.Model
	executionEventHistory  tea.Model
	executionEventOverview tea.Model
}

func NewView(soloCtx *context.CliContext) (tea.Model, error) {
	executionEventRepository := repository.NewJSONFileRepository[*domain.ExecutionEvent]()
	containerStepMapRepository := repository.NewJSONFileRepository[domain.ContainerStepMap]()
	activeModel := layout.NewActiveModel()

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
		heading: heading.NewComponent("Workflow Logs"),
		executionEventHistory: execution_event_history.NewModel(
			soloCtx,
			activeModel,
			executionEventRepository,
			execution_event_history.WithActive(),
		),
		executionEventOverview: execution_event_overview.NewModel(
			activeModel,
			containerStepMapRepository,
		),
	}, nil
}

func (t View) Init() tea.Cmd {
	return tea.Batch(
		t.heading.Init(),
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

		cmds = append(cmds, t.layoutManager.ResizeComponents(
			t.width,
			t.height,
			&t.heading,
			&t.executionEventHistory,
			&t.executionEventOverview,
		)...)

	default:
		if m, ok := msg.(tea.KeyMsg); ok && m.String() == "ctrl+c" {
			return t, tea.Quit
		}

		// Forward message
		for _, model := range []*tea.Model{
			&t.heading,
			&t.executionEventHistory,
			&t.executionEventOverview,
		} {
			*model, cmd = (*model).Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return t, tea.Batch(cmds...)
}

func (t View) View() tea.View {
	v := tea.NewView(t.layoutManager.Render(
		t.heading.View().Content,
		t.executionEventHistory.View().Content,
		t.executionEventOverview.View().Content,
	))

	v.AltScreen = true

	return v
}
