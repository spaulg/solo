package progress

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	progresscommon "github.com/spaulg/solo/internal/pkg/impl/common/container/progress"
	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/messages"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
)

type ComposeProgressModel struct {
	soloCtx *context.CliContext
	spinner spinner.Model

	contextId         string
	action            progresscommon.ComposeProgressAction
	entityType        progresscommon.ProgressEntityTypeName
	projectEntityName string
	status            progresscommon.ComposeProgressStatus
}

func NewComposeProgressModel(soloCtx *context.CliContext, contextId string) tea.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("202"))

	return &ComposeProgressModel{
		soloCtx:   soloCtx,
		spinner:   s,
		contextId: contextId,
	}
}

func (t *ComposeProgressModel) Init() tea.Cmd {
	return t.spinner.Tick
}

func (t *ComposeProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case spinner.TickMsg:
		if t.status == progresscommon.Complete {
			return t, nil
		}

		var cmd tea.Cmd
		t.spinner, cmd = t.spinner.Update(msg)
		return t, cmd

	case messages.ComposeProgressMsg:
		if t.contextId != m.ContextId {
			t.soloCtx.Logger.Error("Context id does not match message, ignoring")
			return t, nil
		}

		t.action = m.Action
		t.entityType = m.EntityType
		t.projectEntityName = m.ProjectEntityName
		t.status = m.Status
	}

	return t, nil
}

func (t *ComposeProgressModel) View() (s string) {
	if t.status == progresscommon.InProgress {
		s += t.spinner.View()
	} else {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
		s += style.Render("\u2714 ")
	}

	s += t.action.String()

	return
}
