package models

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spaulg/solo/internal/pkg/solo/bubbletea/messages"
	"github.com/spaulg/solo/internal/pkg/solo/container/progress"
	"github.com/spaulg/solo/internal/pkg/solo/context"
)

type ComposeProgressServiceStatusMap map[string]progress.ComposeProgressStatus
type ComposeProgressActionNameMap map[progress.ComposeProgressAction][]string
type ComposeProgressActionStatusMap map[progress.ComposeProgressAction]ComposeProgressServiceStatusMap

type ProjectControlActionCompleteMsg struct{}
type ExitAfterRefreshMsg struct{}

type WorkflowStatus struct {
	Output string
	Error  string
	Exit   *uint8
}

type ProgressModel struct {
	soloCtx *context.CliContext
	spinner spinner.Model

	progressOrder  map[progress.ComposeProgressEntityTypeName]ComposeProgressActionNameMap
	progressStatus map[progress.ComposeProgressEntityTypeName]ComposeProgressActionStatusMap
}

func NewProgressModel(soloCtx *context.CliContext) (*ProgressModel, error) {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("202"))

	return &ProgressModel{
		soloCtx: soloCtx,
		spinner: s,

		progressOrder:  make(map[progress.ComposeProgressEntityTypeName]ComposeProgressActionNameMap),
		progressStatus: make(map[progress.ComposeProgressEntityTypeName]ComposeProgressActionStatusMap),
	}, nil
}

func (t ProgressModel) Init() tea.Cmd {
	return t.spinner.Tick
}

func (t ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		t.spinner, cmd = t.spinner.Update(msg)
		return t, cmd

	case messages.ComposeProgressMsg:
		t.updateComposeProgressMsg(m)
	}

	return t, nil
}

func (t ProgressModel) View() (s string) {

	s += t.renderSection(progress.Container, progress.Stop, "Stopping containers")
	s += t.renderSection(progress.Container, progress.Remove, "Removing containers")
	s += t.renderSection(progress.Volume, progress.Remove, "Removing volumes")
	s += t.renderSection(progress.Network, progress.Remove, "Removing networks")

	s += t.renderSection(progress.Network, progress.Create, "Creating networks")
	s += t.renderSection(progress.Volume, progress.Create, "Creating volumes")
	s += t.renderSection(progress.Container, progress.Create, "Creating containers")
	s += t.renderSection(progress.Container, progress.Start, "Starting containers")

	return
}

func (t ProgressModel) updateComposeProgressMsg(m messages.ComposeProgressMsg) {
	if _, exists := t.progressOrder[m.Type]; !exists {
		t.progressOrder[m.Type] = make(ComposeProgressActionNameMap)
	}

	if _, exists := t.progressOrder[m.Type][m.Action]; !exists {
		t.progressOrder[m.Type][m.Action] = make([]string, 0)
	}

	if _, exists := t.progressStatus[m.Type]; !exists {
		t.progressStatus[m.Type] = make(ComposeProgressActionStatusMap)
	}

	if _, exists := t.progressStatus[m.Type][m.Action]; !exists {
		t.progressStatus[m.Type][m.Action] = make(ComposeProgressServiceStatusMap)
	}

	if _, exists := t.progressStatus[m.Type][m.Action][m.ProjectEntityName]; !exists {
		t.progressOrder[m.Type][m.Action] = append(t.progressOrder[m.Type][m.Action], m.ProjectEntityName)
	}

	t.progressStatus[m.Type][m.Action][m.ProjectEntityName] = m.Status
}

func (t ProgressModel) renderSection(
	entityTypeName progress.ComposeProgressEntityTypeName,
	action progress.ComposeProgressAction,
	message string,
) (s string) {
	if len(t.progressOrder[entityTypeName][action]) > 0 {
		s += message + "\n"
		for _, entityName := range t.progressOrder[entityTypeName][action] {
			if t.progressStatus[entityTypeName][action][entityName] == progress.InProgress {
				s += "  " + t.spinner.View() + entityName + "\n"
			} else {
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
				s += "  " + style.Render("\u2714 ") + entityName + "\n"
			}
		}
	}

	return
}
