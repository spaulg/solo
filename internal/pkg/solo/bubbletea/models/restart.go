package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spaulg/solo/internal/pkg/solo/bubbletea/messages"
	"github.com/spaulg/solo/internal/pkg/solo/context"
)

type RestartModel struct {
	soloCtx               *context.CliContext
	projectControl        *solo.ProjectControl
	projectActionComplete bool
	progressModel         tea.Model
}

func NewRestartModel(soloCtx *context.CliContext) (*RestartModel, error) {
	projectControl, err := solo.ProjectControlFactory(soloCtx)
	if err != nil {
		return nil, err
	}

	progressModel, err := NewProgressModel(soloCtx)
	if err != nil {
		return nil, err
	}

	return &RestartModel{
		soloCtx:        soloCtx,
		projectControl: projectControl,
		progressModel:  *progressModel,
	}, nil
}

func (t RestartModel) Init() tea.Cmd {
	return tea.Batch(func() tea.Msg {
		if err := t.projectControl.Stop(); err != nil {
			return messages.ErrorMsg(err)
		}

		if err := t.projectControl.Start(); err != nil {
			return messages.ErrorMsg(err)
		}

		return ProjectControlActionCompleteMsg{}
	}, t.progressModel.Init())
}

func (t RestartModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m := msg.(type) {
	case ExitAfterRefreshMsg:
		return t, tea.Quit

	case ProjectControlActionCompleteMsg:
		t.projectActionComplete = true
		return t, func() tea.Msg {
			return ExitAfterRefreshMsg{}
		}

	case tea.KeyMsg:
		if m.String() == "ctrl+c" {
			return t, tea.Quit
		}

	default:
		t.progressModel, cmd = t.progressModel.Update(msg)
	}

	return t, cmd
}

func (t RestartModel) View() (s string) {
	s += t.progressModel.View()

	if t.projectActionComplete {
		s += "\n\nProject restarted successfully\n"
	}

	return
}
