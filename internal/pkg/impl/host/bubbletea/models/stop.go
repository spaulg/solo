package models

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/spaulg/solo/internal/pkg/impl/host"
	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/messages"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
)

type StopModel struct {
	soloCtx               *context.CliContext
	projectControl        *host.ProjectControl
	projectActionComplete bool
	progressModel         tea.Model
}

func NewStopModel(soloCtx *context.CliContext) (tea.Model, error) {
	projectControl, err := host.ProjectControlFactory(soloCtx)
	if err != nil {
		return nil, err
	}

	progressModel, err := NewProgressModel(soloCtx)
	if err != nil {
		return nil, err
	}

	return &StopModel{
		soloCtx:        soloCtx,
		projectControl: projectControl,
		progressModel:  progressModel,
	}, nil
}

func (t *StopModel) Init() tea.Cmd {
	return tea.Batch(func() tea.Msg {
		if err := t.projectControl.Stop(); err != nil {
			return messages.ErrorMsg(err)
		}

		return ProjectControlActionCompleteMsg{}
	}, t.progressModel.Init())
}

func (t *StopModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m := msg.(type) {
	//case ExitAfterRefreshMsg:
	//	return t, tea.Quit

	case messages.ErrorMsg:
		return t, func() tea.Msg {
			return ExitAfterRefreshMsg{}
		}

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

func (t *StopModel) View() (s string) {
	s += t.progressModel.View()

	if t.projectActionComplete {
		s += "\n\nProject stopped successfully\n"
	}

	return
}
